package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"inHack/internal/service"
	"io"
	"mime/multipart"
	"os"

	"net/http"
	"time"
)

type Handler struct {
	services *service.Service
}

func NewHandlers(service *service.Service) *Handler {
	return &Handler{services: service}
}

func (h *Handler) InitRouter() *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	})
	router.POST("/analyze", h.analyzeCarHandler)
	router.GET("/health", h.healthHandler)
	return router
}
func (h *Handler) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "🚀 Go server is healthy",
		"time":    time.Now().Format(time.RFC3339),
	})
}
func (h *Handler) analyzeCarHandler(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
	}
	defer src.Close()

	url := "http://localhost:5000/predict"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", file.Filename)
	io.Copy(part, src)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "FastAPI not available"})
		return
	}
	defer resp.Body.Close()
	//
	//filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	//filePath := filepath.Join("./uploads", filename)
	//
	//if err := c.SaveUploadedFile(file, filePath); err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
	//	return
	//}
	//
	//log.Printf("📨 Received car image : %s", file.Filename)
	//
	//fileData, err := os.ReadFile(filePath)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
	//	return
	//}
	//
	//result, err := analyzeCarImage(fileData)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, gin.H{"error": "Analysis failed"})
	//	return
	//}
	//
	//// Формируем ответ
	//response := gin.H{
	//	"filename":    file.Filename,
	//	"condition":   result.Condition,
	//	"confidence":  result.Confidence,
	//	"message":     "Analysis completed successfully",
	//	"issues":      result.Issues,
	//	"analysis_id": filename,
	//}
	//
	//c.JSON(http.StatusOK, response)
		
	// читаем JSON от FastAPI
	respBody, _ := io.ReadAll(resp.Body)
	var fastApiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &fastApiResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bad FastAPI response"})
		return
	}

	// Берём путь к изображению
	resultPath, ok := fastApiResp["result_image"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No image in response"})
		return
	}

	// Открываем картинку
	imgBytes, err := os.ReadFile(resultPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Image not found"})
		return
	}

	// Отправляем картинку фронту
	c.Data(http.StatusOK, "image/jpeg", imgBytes)
}
func analyzeCarImage(imageData []byte) (*AnalysisResult, error) {
	// Формируем multipart/form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "car.jpg")
	part.Write(imageData)
	writer.Close()

	// Отправляем на Python FastAPI
	resp, err := http.Post("http://localhost:5000/analyze", writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to call FastAPI: %w", err)
	}
	defer resp.Body.Close()

	log.Printf("➡️ Sent image to FastAPI, status: %s", resp.Status)

	// читаем тело ответа
	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("⬅️ Response body: %s", string(respBody))

	var result AnalysisResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("❌ failed to unmarshal response: %w", err)
	}
	return &result, nil

}

type AnalysisResult struct {
	Condition  string     `json:"condition"`
	Confidence int        `json:"confidence"`
	Issues     []CarIssue `json:"issues"`
}
type CarIssue struct {
	Type        string `json:"type"`
	Location    string `json:"location"`
	Severity    int    `json:"severity"`
	Description string `json:"description"`
}

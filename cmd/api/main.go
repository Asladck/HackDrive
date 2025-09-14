package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"inHack/internal/handler"
	"inHack/internal/repository"
	"inHack/internal/service"
	"inHack/scripts/pb"
	"mime/multipart"
	"net"
	"net/http"
	"os"
)

type Server struct {
	mlServiceURL string
	pb.UnimplementedCarAnalysisServer
}

type AnalysisResult struct {
	Condition  string `json:"condition"`
	Confidence int    `json:"confidence"`
}

func main() {
	log.SetFormatter(new(log.JSONFormatter))

	if err := initConfig(); err != nil {
		log.Fatal("error initializing configs", err)
	}
	if err := godotenv.Load(); err != nil {
		log.Fatal("error initializing configs", err)
	}

	//db, err := repository.NewPostgresDB(repository.Config{
	//	Host:     viper.GetString("db.host"),
	//	Port:     viper.GetString("db.port"),
	//	Username: viper.GetString("db.username"),
	//	Password: os.Getenv("DB_PASSWORD"),
	//	DBName:   viper.GetString("db.dbname"),
	//	SSLMode:  viper.GetString("db.sslmode"),
	//})
	//if err != nil {
	//	log.Fatalf("failed to initializate a db: %s", err.Error())
	//}
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandlers(services)

	go startGRPCServer()

	router := handlers.InitRouter()

	log.Println("🚗 Car Analysis Server starting on port 8080...")
	if err := router.Run(":8000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}

func startGRPCServer() {
	grpcServer := grpc.NewServer()
	server := &Server{mlServiceURL: "http://localhost:5000"}
	pb.RegisterCarAnalysisServer(grpcServer, server)

	listener, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	log.Println("🚗 gRPC Server starting on port 8080...")
	log.Fatal(grpcServer.Serve(listener))
}

func (s *Server) AnalyzeCar(ctx context.Context, req *pb.AnalyzeRequest) (*pb.AnalyzeResponse, error) {
	log.Printf("Received car image: %s (%d bytes)", req.Filename, len(req.ImageData))

	filePath := "./storage/" + req.Filename
	if err := os.WriteFile(filePath, req.ImageData, 0644); err != nil {
		log.Printf("Failed to save image: %v", err)
		return nil, err
	}

	condition, confidence, err := analyzeCarCondition(req.ImageData)
	if err != nil {
		log.Printf("❌ Analysis failed: %v", err)
		return nil, err
	}

	log.Printf("✅ Analysis result: %s (confidence %d)", condition, confidence)

	return &pb.AnalyzeResponse{
		Filename:   req.Filename,
		Condition:  condition,
		Confidence: int32(confidence),
		Message:    "Analysis completed successfully",
	}, nil
}
func analyzeCarCondition(imageData []byte) (string, int, error) {
	// Создаём multipart/form-data запрос
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "upload.jpg")
	if err != nil {
		return "", 0, err
	}
	part.Write(imageData)
	writer.Close()

	resp, err := http.Post("http://localhost:5000/analyze", writer.FormDataContentType(), body)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Condition   string `json:"condition"`
		Confidence  int    `json:"confidence_demo"`
		ResultImage string `json:"result_image"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	return result.Condition, result.Confidence, nil
}
func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

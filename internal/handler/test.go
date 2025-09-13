package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
)

func (h *Handler) test(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		NewErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	filename := filepath.Base(file.Filename)
	uploadDir := "./files"
	if err := c.SaveUploadedFile(file, filepath.Join(uploadDir, filename)); err != nil {
		NewErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, "file with name %s loaded", file.Filename)
}

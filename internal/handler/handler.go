package handler

import (
	"github.com/gin-gonic/gin"
	files "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	//"inHack/internal/service"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}
func (h *Handler) InitRouter() *gin.Engine {
	router := gin.New()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(files.Handler))
	router.GET("/", func(c *gin.Context) {
		c.File("index.html")
	})
	h.InitTest(router)
	return router
}
func (h *Handler) InitTest(router *gin.Engine) {
	router.POST("/test", h.test)
}

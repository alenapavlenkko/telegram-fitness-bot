package api

import (
	"fmt"
	"net/http"

	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/gin-gonic/gin"
)

func StartServer(ts *service.TrainingService) {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("ngrok-skip-browser-warning", "true")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	r.Static("/assets", "./frontend/dist/assets")

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/app")
	})

	r.GET("/app", func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	r.GET("/api/trainings", func(c *gin.Context) {
		trainings, err := ts.ListTrainings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, trainings)
	})

	fmt.Println("🌐 Сервер запущен: http://localhost:8080/app")

	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

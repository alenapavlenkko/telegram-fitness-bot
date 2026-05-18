package api

import (
	"fmt"
	"net/http"

	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/gin-gonic/gin"
)

type ServerDeps struct {
	TrainingService  *service.TrainingService
	NutritionService *service.NutritionService
}

func StartServer(deps ServerDeps) {
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
		trainings, err := deps.TrainingService.ListTrainings()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, trainings)
	})

	r.GET("/api/nutrition", func(c *gin.Context) {
		nutrition, err := deps.NutritionService.ListNutrition()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, nutrition)
	})

	r.GET("/api/stats", func(c *gin.Context) {
		trainings, _ := deps.TrainingService.ListTrainings()
		nutrition, _ := deps.NutritionService.ListNutrition()

		totalMinutes := 0
		for _, t := range trainings {
			totalMinutes += t.Duration
		}

		totalCalories := 0
		for _, n := range nutrition {
			totalCalories += n.Calories
		}

		c.JSON(http.StatusOK, gin.H{
			"trainingsCount": len(trainings),
			"nutritionCount": len(nutrition),
			"totalMinutes":   totalMinutes,
			"totalCalories":  totalCalories,
		})
	})

	fmt.Println("🌐 Сервер запущен: http://localhost:8080/app")

	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

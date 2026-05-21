package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/gin-gonic/gin"
)

type ServerDeps struct {
	TrainingService  *service.TrainingService
	NutritionService *service.NutritionService
	UserService      *service.UserService
	WeightService    *service.WeightService
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

	r.GET("/api/profile/:telegramId", func(c *gin.Context) {
		telegramID, err := strconv.ParseInt(c.Param("telegramId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid telegram id"})
			return
		}

		user, err := deps.UserService.GetUserByTelegramID(telegramID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	r.POST("/api/profile", func(c *gin.Context) {
		var input service.UpdateProfileDTO

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := deps.UserService.UpdateProfile(input)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	})

	r.POST("/api/weight", func(c *gin.Context) {
		var input struct {
			TelegramID int64   `json:"telegramId"`
			Weight     float64 `json:"weight"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user, err := deps.UserService.GetUserByTelegramID(input.TelegramID)
		if err != nil {
			user, err = deps.UserService.CreateUser(service.CreateUserDTO{
				TelegramID: input.TelegramID,
				Name:       "User",
				Role:       "user",
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		if err := deps.WeightService.LogWeight(uint(user.ID), input.Weight); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "weight saved"})
	})

	r.GET("/api/weight/:telegramId", func(c *gin.Context) {
		telegramID, err := strconv.ParseInt(c.Param("telegramId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid telegram id"})
			return
		}

		user, err := deps.UserService.GetUserByTelegramID(telegramID)
		if err != nil {
			c.JSON(http.StatusOK, []any{})
			return
		}

		logs, err := deps.WeightService.GetUserHistory(uint(user.ID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, logs)
	})

	r.GET("/api/progress/:telegramId", func(c *gin.Context) {
		telegramID, err := strconv.ParseInt(c.Param("telegramId"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid telegram id"})
			return
		}

		user, err := deps.UserService.GetUserByTelegramID(telegramID)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"currentWeight": 0,
				"startWeight":   0,
				"targetWeight":  0,
				"change":        0,
				"logsCount":     0,
			})
			return
		}

		logs, err := deps.WeightService.GetUserHistory(uint(user.ID))
		if err != nil || len(logs) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"currentWeight": user.Weight,
				"startWeight":   user.Weight,
				"targetWeight":  user.TargetWeight,
				"change":        0,
				"logsCount":     0,
			})
			return
		}

		startWeight := logs[0].Weight
		currentWeight := logs[len(logs)-1].Weight

		c.JSON(http.StatusOK, gin.H{
			"currentWeight": currentWeight,
			"startWeight":   startWeight,
			"targetWeight":  user.TargetWeight,
			"change":        startWeight - currentWeight,
			"logsCount":     len(logs),
		})
	})

	fmt.Println("🌐 Сервер запущен: http://localhost:8080/app")

	if err := r.Run("0.0.0.0:8080"); err != nil {
		panic(err)
	}
}

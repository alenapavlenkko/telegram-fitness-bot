package admin

import (
	"net/http"

	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/gin-gonic/gin"
)

// Handlers содержит зависимости от сервисов
type Handlers struct {
	trainingService *service.TrainingService
	// Добавьте сюда другие сервисы по мере необходимости
}

// Используем все сервисы в админке
func SetupRoutes(r *gin.Engine,
	trainingService *service.TrainingService,
	categoryService *service.CategoryService,
	nutritionService *service.NutritionService,
	userService *service.UserService,
	progressService *service.ProgressService,
) {
	adminGroup := r.Group("/admin")

	// Тренировки
	adminGroup.GET("/trainings", func(c *gin.Context) {
		t, err := trainingService.ListTrainings()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, t)
	})

	// Категории
	adminGroup.GET("/categories", func(c *gin.Context) {
		cats, err := categoryService.ListCategories()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, cats)
	})

	// Питание
	adminGroup.GET("/nutrition", func(c *gin.Context) {
		n, err := nutritionService.List()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, n)
	})

	// Пользователи — если нет метода для списка, можно временно вернуть пустой массив
	adminGroup.GET("/users", func(c *gin.Context) {
		users, err := userService.ListUsers()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, users)
	})

	// Прогресс
	adminGroup.GET("/progress", func(c *gin.Context) {
		p, err := progressService.ListProgress()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, p)
	})
}

// Пример обработчика: Получить список всех тренировок
func (h *Handlers) ListTrainings(c *gin.Context) {

	trainings, err := h.trainingService.GetAllTrainings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch trainings"})
		return
	}

	// Если используете REST API для фронтенда (React/Vue):
	c.JSON(http.StatusOK, trainings)

	// Если используете серверный рендеринг (Go templates):
	// c.HTML(http.StatusOK, "trainings_list.html", gin.H{"Trainings": trainings})
}

// Пример обработчика: Создать тренировку через REST API (используя DTO бота)
func (h *Handlers) CreateTraining(c *gin.Context) {
	var input service.CreateTrainingDTO
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	training, err := h.trainingService.CreateTraining(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create training"})
		return
	}

	c.JSON(http.StatusCreated, training)
}

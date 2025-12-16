package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/alenapavlenkko/telegramfitnes/internal/admin"
	"github.com/alenapavlenkko/telegramfitnes/internal/database"
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	// Подключение к базе
	db, err := database.NewPostgres(dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Авто-миграция
	if err := database.AutoMigrateTables(db,
		&models.Category{},
		&models.TrainingProgram{},
		&models.NutritionPlan{},
		&models.User{},
		&models.UserProgress{},
	); err != nil {
		log.Fatal("Failed to migrate tables:", err)
	}

	// Репозитории
	trainingRepo := repository.NewTrainingRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	nutritionRepo := repository.NewNutritionRepo(db)
	userRepo := repository.NewUserRepo(db)
	progressRepo := repository.NewProgressRepo(db)

	// Сервисы
	trainingService := service.NewTrainingService(trainingRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	nutritionService := service.NewNutritionService(nutritionRepo)
	userService := service.NewUserService(userRepo)
	progressService := service.NewProgressService(progressRepo)

	// Gin router
	router := gin.Default()

	// Используем все сервисы через SetupRoutes
	admin.SetupRoutes(router, trainingService, categoryService, nutritionService, userService, progressService)

	log.Println("Admin panel starting on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to run admin panel:", err)
	}
}

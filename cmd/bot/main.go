package main

import (
	"os"

	"github.com/alenapavlenkko/telegramfitnes/internal/bot"
	"github.com/alenapavlenkko/telegramfitnes/internal/database"
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	"github.com/alenapavlenkko/telegramfitnes/pkg/utils"
	"github.com/joho/godotenv"
)

func main() {
	// -----------------------
	// ENV
	if err := godotenv.Load(); err != nil {
		utils.Log.Info("No .env file found")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		utils.Log.Error("DATABASE_URL not set")
		os.Exit(1)
	}

	// -----------------------
	// DATABASE
	db, err := database.NewPostgres(dsn)
	if err != nil {
		utils.Log.Error("Failed to connect to database: " + err.Error())
		os.Exit(1)
	}
	utils.Log.Info("Database connected")

	// Выполнение миграций для ВСЕХ моделей
	if err := database.AutoMigrateTables(db,
		&models.Category{},
		&models.TrainingProgram{},
		&models.NutritionPlan{},
		&models.User{},
		&models.UserProgress{},
		&models.WeeklyMenu{}, // Добавьте эту модель
		&models.MenuDay{},    // Добавьте эту модель
		&models.DayMeal{},    // Добавьте эту модель
	); err != nil {
		utils.Log.Error("Failed to migrate database: " + err.Error())
		os.Exit(1)
	}

	// -----------------------
	// REPOSITORIES
	trainingRepo := repository.NewTrainingRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	nutritionRepo := repository.NewNutritionRepo(db)
	weeklyMenuRepo := repository.NewWeeklyMenuRepo(db)
	userRepo := repository.NewUserRepo(db)
	progressRepo := repository.NewProgressRepo(db)

	// -----------------------
	// SERVICES
	trainingService := service.NewTrainingService(trainingRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	nutritionService := service.NewNutritionService(nutritionRepo, weeklyMenuRepo)
	userService := service.NewUserService(userRepo)
	progressService := service.NewProgressService(progressRepo)

	// -----------------------
	// BOT
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		utils.Log.Error("TELEGRAM_TOKEN not set")
		os.Exit(1)
	}

	adminIDs := bot.ParseAdminIDs(os.Getenv("ADMIN_IDS"))
	utils.Log.Info("Loaded admin IDs") // или добавьте отдельную функцию форматирования

	botApp, err := bot.NewBotApp(
		token,
		trainingService,
		nutritionService,
		categoryService,
		userService,
		progressService,
		adminIDs,
	)
	if err != nil {
		utils.Log.Error("Failed to create bot: " + err.Error())
		os.Exit(1)
	}

	utils.Log.Info("Telegram bot starting...")
	botApp.Run()
}

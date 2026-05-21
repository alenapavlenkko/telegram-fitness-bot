package main

import (
	"fmt"
	"os"

	"github.com/alenapavlenkko/telegramfitnes/internal/api"
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
		&models.WeeklyMenu{},
		&models.MenuDay{},
		&models.DayMeal{},
		&models.WeightLog{},
	); err != nil {
		utils.Log.Error("Failed to migrate database: " + err.Error())
		os.Exit(1)
	}
	if err := database.SeedData(db); err != nil {
		utils.Log.Error("Failed to seed database: " + err.Error())
		os.Exit(1)
	}
	utils.Log.Info("Seed data loaded")

	// -----------------------
	// REPOSITORIES
	trainingRepo := repository.NewTrainingRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	nutritionRepo := repository.NewNutritionRepo(db)
	weeklyMenuRepo := repository.NewWeeklyMenuRepo(db)
	userRepo := repository.NewUserRepo(db)
	weightRepo := repository.NewWeightRepository(db)

	// -----------------------
	// SERVICES
	trainingService := service.NewTrainingService(trainingRepo)
	categoryService := service.NewCategoryService(categoryRepo)
	nutritionService := service.NewNutritionService(nutritionRepo, weeklyMenuRepo)
	userService := service.NewUserService(userRepo)
	weightService := service.NewWeightService(weightRepo)

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
		weightService,
		adminIDs,
	)
	if err != nil {
		utils.Log.Error("Failed to create bot: " + err.Error())
		os.Exit(1)
	}

	fmt.Println("🚀 Попытка запуска сервера на порту 8080...")
	go api.StartServer(api.ServerDeps{
		TrainingService:  trainingService,
		NutritionService: nutritionService,
		UserService:      userService,
		WeightService:    weightService,
	})

	utils.Log.Info("Telegram bot starting...")
	botApp.Run()
}

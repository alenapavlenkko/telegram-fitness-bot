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

	// ЗАГРУЗКА .ENV
	if err := godotenv.Load(); err != nil {

		utils.Log.Info(
			"No .env file found",
		)
	}

	// Получаем строку подключения к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")

	// Если DATABASE_URL отсутствует — завершаем программу
	if dsn == "" {

		utils.Log.Error(
			"DATABASE_URL not set",
		)

		os.Exit(1)
	}

	// Создаем подключение к PostgreSQL
	db, err := database.NewPostgres(dsn)

	if err != nil {

		utils.Log.Error(
			"Failed to connect to database: " + err.Error(),
		)

		os.Exit(1)
	}

	utils.Log.Info(
		"Database connected",
	)

	// Создаем таблицы в базе данных
	// если они еще не существуют
	if err := database.AutoMigrateTables(

		db,

		// Категории
		&models.Category{},

		// Тренировки
		&models.TrainingProgram{},

		// Питание
		&models.NutritionPlan{},

		// Пользователи
		&models.User{},

		// Недельное меню
		&models.WeeklyMenu{},

		// Дни меню
		&models.MenuDay{},

		// Приемы пищи
		&models.DayMeal{},

		// История веса
		&models.WeightLog{},
	); err != nil {

		utils.Log.Error(
			"Failed to migrate database: " + err.Error(),
		)

		os.Exit(1)
	}

	// Заполняем базу начальными данными
	if err := database.SeedData(db); err != nil {

		utils.Log.Error(
			"Failed to seed database: " + err.Error(),
		)

		os.Exit(1)
	}

	utils.Log.Info(
		"Seed data loaded",
	)

	// Репозиторий тренировок
	trainingRepo := repository.NewTrainingRepo(db)

	// Репозиторий категорий
	categoryRepo := repository.NewCategoryRepository(db)

	// Репозиторий питания
	nutritionRepo := repository.NewNutritionRepo(db)

	// Репозиторий недельного меню
	weeklyMenuRepo := repository.NewWeeklyMenuRepo(db)

	// Репозиторий пользователей
	userRepo := repository.NewUserRepo(db)

	// Репозиторий веса
	weightRepo := repository.NewWeightRepository(db)

	// Бизнес-логика тренировок
	trainingService := service.NewTrainingService(
		trainingRepo,
	)

	// Бизнес-логика категорий
	categoryService := service.NewCategoryService(
		categoryRepo,
	)

	// Бизнес-логика питания
	nutritionService := service.NewNutritionService(
		nutritionRepo,
		weeklyMenuRepo,
	)

	// Бизнес-логика пользователей
	userService := service.NewUserService(
		userRepo,
	)

	// Бизнес-логика веса
	weightService := service.NewWeightService(
		weightRepo,
	)

	// Получаем токен Telegram бота
	token := os.Getenv("TELEGRAM_TOKEN")

	if token == "" {

		utils.Log.Error(
			"TELEGRAM_TOKEN not set",
		)

		os.Exit(1)
	}

	// Получаем список Telegram ID администраторов
	adminIDs := bot.ParseAdminIDs(

		os.Getenv("ADMIN_IDS"),
	)

	utils.Log.Info(
		"Loaded admin IDs",
	)

	// Создаем экземпляр Telegram бота
	botApp, err := bot.NewBotApp(

		// Telegram token
		token,

		// Services
		trainingService,
		nutritionService,
		categoryService,
		userService,
		weightService,

		// Администраторы
		adminIDs,
	)

	if err != nil {

		utils.Log.Error(
			"Failed to create bot: " + err.Error(),
		)

		os.Exit(1)
	}

	fmt.Println(
		"🚀 Попытка запуска сервера на порту 8080...",
	)

	// Запускаем API сервер параллельно
	go api.StartServer(

		api.ServerDeps{

			TrainingService:  trainingService,
			NutritionService: nutritionService,
			UserService:      userService,
			WeightService:    weightService,
			AdminIDs:         adminIDs,
		},
	)

	utils.Log.Info(
		"Telegram bot starting...",
	)

	// Запускаем бота
	botApp.Run()
}

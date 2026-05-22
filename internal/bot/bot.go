package bot

import (
	"log"

	"github.com/alenapavlenkko/telegramfitnes/internal/admin"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotApp — главная структура Telegram бота
type BotApp struct {
	API *tgbotapi.BotAPI

	Admins []int64

	trainingService  *service.TrainingService
	nutritionService *service.NutritionService
	categoryService  *service.CategoryService
	userService      *service.UserService
	weightService    *service.WeightService

	adminHandler *admin.AdminHandler

	// Состояния пользователей
	userStates map[int64]string

	// Данные калькулятора
	calculatorData map[int64]map[string]string
}

// NewBotApp создает нового Telegram бота
func NewBotApp(
	token string,
	trainingService *service.TrainingService,
	nutritionService *service.NutritionService,
	categoryService *service.CategoryService,
	userService *service.UserService,
	weightService *service.WeightService,
	adminIDs []int64,
) (*BotApp, error) {

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	bot := &BotApp{
		API:              botAPI,
		Admins:           adminIDs,
		trainingService:  trainingService,
		nutritionService: nutritionService,
		categoryService:  categoryService,
		userService:      userService,
		weightService:    weightService,
		userStates:       make(map[int64]string),
		calculatorData:   make(map[int64]map[string]string),
	}

	// Создаем admin handler
	bot.adminHandler = admin.NewAdminHandler(
		trainingService,
		nutritionService,
		categoryService,
		userService,
		bot.sendText,
		func(chatID int64, text string, rows [][]tgbotapi.InlineKeyboardButton) {
			bot.sendTextWithKeyboard(chatID, text, rows)
		},
	)

	return bot, nil
}

// Run запускает Telegram бота
func (b *BotApp) Run() {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	log.Println("🤖 Bot started")

	for update := range updates {

		// Callback кнопки
		if update.CallbackQuery != nil {
			b.handleCallback(update.CallbackQuery)
			continue
		}

		if update.Message == nil {
			continue
		}

		// Команды
		if update.Message.IsCommand() {
			b.handleCommand(update)
			continue
		}

		// Обычные сообщения
		b.handleRegularMessage(update)
	}
}

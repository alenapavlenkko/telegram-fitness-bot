package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/alenapavlenkko/telegramfitnes/internal/admin"
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// BotApp — основная структура бота
type BotApp struct {
	API *tgbotapi.BotAPI

	Admins           []int64
	Handlers         map[string]func(tgbotapi.Update)
	trainingService  *service.TrainingService
	nutritionService *service.NutritionService
	categoryService  *service.CategoryService
	userService      *service.UserService
	weightService    *service.WeightService
	userStates       map[int64]string
	calculatorData   map[int64]map[string]string

	// Админ-панель
	adminHandler *admin.AdminHandler
}

// Конструктор бота
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

	// Создаем админ-хендлер с функцией отправки сообщений
	bot.adminHandler = admin.NewAdminHandler(
		trainingService,
		nutritionService,
		categoryService,
		userService,
		bot.sendText, // передаем функцию отправки сообщений
		func(chatID int64, text string, rows [][]tgbotapi.InlineKeyboardButton) {
			bot.sendTextWithKeyboard(chatID, text, rows)
		},
	)

	// Добавьте проверку после создания
	if bot.adminHandler == nil {
		log.Println("ERROR: AdminHandler is nil after creation!")
	} else {
		log.Printf("AdminHandler created successfully: %v", bot.adminHandler)
	}
	return bot, nil
}

// Запуск бота
func (b *BotApp) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.API.GetUpdatesChan(u)
	log.Println("🤖 Bot started")

	for update := range updates {
		// Обработка callback-кнопок
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery

			if b.isAuthorized(callback.From.ID, "admin") {
				b.adminHandler.HandleAdminCallback(callback)
			} else {
				b.answerCallback(callback.ID, "⛔ Доступ запрещен")
			}
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

// Проверка админа
func (b *BotApp) isAdmin(userID int64) bool {
	for _, id := range b.Admins {
		if id == userID {
			return true
		}
	}
	return false
}

// Команды
func (b *BotApp) handleCommand(update tgbotapi.Update) {
	cmd := update.Message.Command()
	chatID := update.Message.Chat.ID

	switch cmd {
	case "start":
		_, err := b.authenticateUser(update)
		if err != nil {
			b.sendText(chatID, "❌ Ошибка авторизации")
			return
		}

		// Отправляем приветственное сообщение
		b.sendText(chatID, "👋 Рад вас видеть! Сейчас открою главное меню...")
		b.showMainMenu(chatID)
	case "help":
		helpMsg := `📚 *Помощь по использованию Fitness Bot*

*Основные команды:*
/start - Главное меню
/help - Эта справка
/admin - Панель администратора (только для админов)

*Как пользоваться:*
1. Используйте кнопки меню для навигации
2. Тренировки - выбирайте программы упражнений
3. Питание - изучайте планы питания
4. Категории - фильтруйте контент по темам

*Для администраторов:*
• Добавляйте новые тренировки и блюда
• Создавайте недельные меню
• Управляйте категориями
• Активируйте меню для пользователей

*Поддержка:* Если возникли проблемы, свяжитесь с администратором.`

		b.sendText(chatID, helpMsg)

	case "admin":
		log.Printf("[DEBUG] Admin command received from user ID: %d", update.Message.From.ID)
		log.Printf("[DEBUG] Admin list: %v", b.Admins)
		log.Printf("[DEBUG] Is admin? %v", b.isAdmin(int64(update.Message.From.ID)))
		log.Printf("[DEBUG] Admin handler: %v", b.adminHandler)

		if !b.isAuthorized(int64(update.Message.From.ID), "admin") {
			b.sendText(chatID, "⛔ Недостаточно прав")
			return
		}

		log.Printf("[DEBUG] Calling ShowAdminPanel for chat %d", chatID)
		b.adminHandler.ShowAdminPanel(chatID)

	case "log_weight":
		b.sendText(chatID, "⚖️ Введите ваш вес в кг (например: 75.5):")
		b.userStates[chatID] = "awaiting_weight"
	case "stats":
		b.handleStatsCommand(update, chatID)

	case "checkdb":
		if b.isAdmin(int64(update.Message.From.ID)) {
			b.checkDatabase(chatID)
		}
	case "foodlist":
		if b.isAdmin(int64(update.Message.From.ID)) {
			b.adminHandler.ShowNutritionListForSelection(chatID)
		}
	case "test":
		log.Printf("[TEST] Testing admin handler")
		log.Printf("[TEST] AdminHandler is nil? %v", b.adminHandler == nil)
		log.Printf("[TEST] ChatID: %d", chatID)

		// Просто отправьте тестовое сообщение
		b.sendText(chatID, "Тест работает! Бот активен.")

		// Попробуйте вызвать админ-панель напрямую
		if b.adminHandler != nil {
			log.Printf("[TEST] Calling ShowAdminPanel")
			b.adminHandler.ShowAdminPanel(chatID)
		} else {
			b.sendText(chatID, "AdminHandler is nil!")
		}
	default:
		b.sendText(chatID, "Неизвестная команда. Используйте /help")
	}
}
func (b *BotApp) handleStatsCommand(update tgbotapi.Update, chatID int64) {
	user, err := b.authenticateUser(update)
	if err != nil {
		b.sendText(chatID, "❌ Ошибка авторизации")
		return
	}

	logs, err := b.weightService.GetUserHistory(uint(user.ID))
	if err != nil || len(logs) == 0 {
		b.sendText(chatID, "📊 У вас пока нет записей о весе.\n\nИспользуйте /log_weight чтобы записать первый вес!")
		return
	}

	// Статистика
	var min, max, sum float64
	min = logs[0].Weight
	max = logs[0].Weight

	for _, log := range logs {
		sum += log.Weight
		if log.Weight < min {
			min = log.Weight
		}
		if log.Weight > max {
			max = log.Weight
		}
	}

	avg := sum / float64(len(logs))
	current := logs[len(logs)-1].Weight
	start := logs[0].Weight
	change := start - current

	msg := fmt.Sprintf("📊 *Ваша статистика веса*\n\n"+
		"⚖️ Текущий вес: *%.1f кг*\n"+
		"📍 Начальный вес: *%.1f кг*\n"+
		"📈 Изменение: *%.1f кг*\n"+
		"📊 Мин: *%.1f кг*\n"+
		"📊 Макс: *%.1f кг*\n"+
		"📈 Средний: *%.1f кг*\n"+
		"📅 Всего записей: *%d*\n",
		current, start, change, min, max, avg, len(logs))

	if change > 0 {
		msg += "\n🎉 Отлично! Вы сбросили вес!"
	} else if change < 0 {
		msg += "\n💪 Вы набираете массу!"
	}

	b.sendText(chatID, msg)
}

func (b *BotApp) checkDatabase(chatID int64) {
	trainings, err := b.trainingService.ListTrainings()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка БД: "+err.Error())
		return
	}

	if len(trainings) == 0 {
		b.sendText(chatID, "📭 В БД нет тренировок")
		return
	}

	msg := fmt.Sprintf("✅ В БД найдено %d тренировок:\n\n", len(trainings))
	for i, t := range trainings {
		// Используем рефлексию для безопасного доступа к полям
		title := t.Title
		duration := t.Duration

		msg += fmt.Sprintf("%d. %s - %d мин\n", i+1, title, duration)
	}

	b.sendText(chatID, msg)
}
func (b *BotApp) handleRegularMessage(update tgbotapi.Update) {
	userID := int64(update.Message.From.ID)
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	log.Printf("Regular message: userID=%d, text='%s'", userID, text)

	state, isAdminAction := b.adminHandler.GetState(userID)
	if isAdminAction {
		log.Println("Admin action detected")
		b.adminHandler.HandleAdminActions(chatID, userID, state, text)
		return
	}

	if userState, exists := b.userStates[userID]; exists {
		switch userState {
		case "awaiting_weight":
			weight, err := strconv.ParseFloat(text, 64)
			if err != nil || weight <= 20 || weight >= 300 {
				b.sendText(chatID, "❌ Введите корректный вес \\(20-300 кг\\), например: 75.5")
				return
			}

			user, err := b.authenticateUser(update)
			if err != nil {
				b.sendText(chatID, "❌ Ошибка авторизации")
				return
			}

			err = b.weightService.LogWeight(uint(user.ID), weight)
			if err != nil {
				b.sendText(chatID, "❌ Ошибка сохранения: "+err.Error())
			} else {
				b.sendText(chatID, fmt.Sprintf("✅ Вес *%.1f кг* успешно записан!", weight))
			}

			delete(b.userStates, userID)
			return

		case "awaiting_calculator":
			b.handleCalculatorInput(chatID, text)
			delete(b.userStates, userID)
			return
		}
	}
	if state, exists := b.userStates[userID]; exists {

		switch state {

		case "calc_weight":
			b.calculatorData[userID]["weight"] = text
			b.userStates[userID] = "calc_height"
			b.sendText(chatID, "📏 Введите рост в см:")
			return

		case "calc_height":
			b.calculatorData[userID]["height"] = text
			b.userStates[userID] = "calc_age"
			b.sendText(chatID, "🎂 Введите возраст:")
			return

		case "calc_age":
			b.calculatorData[userID]["age"] = text
			b.userStates[userID] = "calc_gender"

			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("👨 Мужской"),
					tgbotapi.NewKeyboardButton("👩 Женский"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, "Выберите пол:")
			msg.ReplyMarkup = keyboard
			b.API.Send(msg)

			return

		case "calc_gender":
			b.calculatorData[userID]["gender"] = text
			b.userStates[userID] = "calc_activity"

			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("🛋 Низкая"),
					tgbotapi.NewKeyboardButton("🚶 Средняя"),
					tgbotapi.NewKeyboardButton("🏃 Высокая"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, "Выберите активность:")
			msg.ReplyMarkup = keyboard
			b.API.Send(msg)

			return

		case "calc_activity":
			b.calculatorData[userID]["activity"] = text
			b.userStates[userID] = "calc_goal"

			keyboard := tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("📉 Похудение"),
					tgbotapi.NewKeyboardButton("⚖ Поддержание"),
					tgbotapi.NewKeyboardButton("📈 Набор массы"),
				),
			)

			msg := tgbotapi.NewMessage(chatID, "Выберите цель:")
			msg.ReplyMarkup = keyboard
			b.API.Send(msg)

			return

		case "calc_goal":
			b.calculatorData[userID]["goal"] = text

			b.finishCalculator(chatID, userID)

			delete(b.userStates, userID)

			return
		}
	}

	if b.isAdmin(userID) {
		log.Println("Admin regular message")
		b.handleAdminRegularMessage(chatID, text)
		return
	}

	log.Println("User action")
	b.handleUserActions(chatID, text)
}

// Обработка обычных сообщений админа (не в админ-панели)
func (b *BotApp) handleAdminRegularMessage(chatID int64, text string) {
	log.Printf("[handleAdminRegularMessage] chatID=%d, text='%s'", chatID, text)

	switch text {
	case "/panel":
		b.adminHandler.ShowAdminPanel(chatID)

	case "/trainings":
		b.adminHandler.ShowTrainingsAdmin(chatID)

	case "/nutrition":
		b.adminHandler.ShowNutritionAdmin(chatID)

	case "/categories":
		b.adminHandler.ShowCategoriesAdmin(chatID)

	case "/foodlist":
		b.adminHandler.ShowNutritionListForSelection(chatID)

	case "🏋️ Тренировки":
		b.showTrainingsForUser(chatID)

	case "🍎 Питание":
		b.showNutritionForUser(chatID)

	case "📂 Категории":
		b.showCategoriesForUser(chatID)

	case "📊 Статистика":
		b.showUserStats(chatID)

	case "👤 Профиль":
		b.showProfile(chatID)

	case "🧮 Калькулятор":
		b.startCalculator(chatID)

	case "⚖️ Записать вес":
		b.sendText(chatID, "⚖️ Введите ваш вес в кг, например: 55.5")
		b.userStates[chatID] = "awaiting_weight"

	case "ℹ️ Помощь":
		b.sendText(chatID,
			"📚 *Помощь*\n\n"+
				"Выберите раздел в меню:\n"+
				"🏋️ Тренировки\n"+
				"🍎 Питание\n"+
				"📊 Статистика\n"+
				"👤 Профиль\n"+
				"🧮 Калькулятор\n"+
				"⚖️ Записать вес",
		)

	default:
		b.showMainMenu(chatID)
	}
}

// Обработка действий обычных пользователей
// В handleUserActions оставьте только вызовы функций show...ForUser
func (b *BotApp) handleUserActions(chatID int64, text string) {
	log.Printf("[handleUserActions] chatID=%d, text='%s'", chatID, text)

	switch text {
	case "🏋️ Тренировки":
		b.showTrainingsForUser(chatID)

	case "🍎 Питание":
		b.showNutritionForUser(chatID)

	case "📂 Категории":
		b.showCategoriesForUser(chatID)

	case "📊 Статистика":
		b.showUserStats(chatID)

	case "👤 Профиль":
		b.showProfile(chatID)

	case "🧮 Калькулятор":
		b.startCalculator(chatID)

	case "⚖️ Записать вес":
		b.sendText(chatID, "⚖️ Введите ваш вес в кг, например: 55.5")
		b.userStates[chatID] = "awaiting_weight"

	case "ℹ️ Помощь":
		b.sendText(chatID, "📚 *Помощь*\n\nВыберите раздел в меню:\n🏋️ Тренировки\n🍎 Питание\n📊 Статистика\n👤 Профиль\n🧮 Калькулятор\n⚖️ Записать вес")

	default:
		b.showMainMenu(chatID)
	}
}

// Методы для пользователей (нужно будет реализовать)
func (b *BotApp) showTrainingsForUser(chatID int64) {
	log.Printf("[showTrainingsForUser] START for chatID=%d", chatID)

	trainings, err := b.trainingService.ListTrainings()
	if err != nil {
		log.Printf("[showTrainingsForUser] ERROR loading trainings: %v", err)
		b.sendText(chatID, "❌ Не удалось загрузить тренировки")
		return
	}

	log.Printf("[showTrainingsForUser] Loaded %d trainings", len(trainings))

	if len(trainings) == 0 {
		log.Println("[showTrainingsForUser] No trainings found")
		b.sendText(chatID, "🏋️ Тренировок пока нет. Следите за обновлениями!")
		return
	}

	escape := func(s string) string {
		specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, c := range specialChars {
			s = strings.ReplaceAll(s, c, "\\"+c)
		}
		return s
	}

	msg := "🏋️ *Доступные тренировки:*\n\n"
	for i, t := range trainings {
		log.Printf("[showTrainingsForUser] Processing training %d: ID=%d, Title='%s'",
			i+1, t.ID, t.Title)

		msg += fmt.Sprintf("%d. *%s* - %d мин\n", i+1, escape(t.Title), t.Duration)
		if t.Description != "" {
			msg += fmt.Sprintf("   %s\n", escape(t.Description))
		}
		if t.YouTubeLink != "" {
			link := strings.ReplaceAll(t.YouTubeLink, "[", "\\[")
			link = strings.ReplaceAll(link, "]", "\\]")
			msg += fmt.Sprintf("   🎥 [Смотреть на YouTube](%s)\n", link)
		}
		msg += "\n"
	}

	log.Printf("[showTrainingsForUser] Sending message of length %d", len(msg))
	b.sendText(chatID, msg)
}

func (b *BotApp) showNutritionForUser(chatID int64) {
	nutritionList, err := b.nutritionService.ListNutrition()
	if err != nil {
		b.sendText(chatID, "❌ Не удалось загрузить планы питания")
		return
	}

	if len(nutritionList) == 0 {
		b.sendText(chatID, "🍎 Планов питания пока нет. Следите за обновлениями!")
		return
	}

	msg := "🍎 *Планы питания:*\n\n"
	for i, n := range nutritionList {
		msg += fmt.Sprintf("%d. *%s* - %d ккал\n", i+1, n.Title, n.Calories)
		if n.Description != "" {
			msg += fmt.Sprintf("   %s\n", n.Description)
		}
		msg += fmt.Sprintf("   Б:%.1fг, У:%.1fг, Ж:%.1fг\n\n", n.Protein, n.Carbs, n.Fats)
	}

	b.sendText(chatID, msg)
}

func (b *BotApp) showCategoriesForUser(chatID int64) {
	categories, err := b.categoryService.ListCategories()
	if err != nil {
		b.sendText(chatID, "❌ Не удалось загрузить категории")
		return
	}

	if len(categories) == 0 {
		b.sendText(chatID, "📂 Категорий пока нет")
		return
	}

	// Группируем по типам
	trainingCats := []string{}
	nutritionCats := []string{}
	generalCats := []string{}

	for _, c := range categories {
		switch c.Type {
		case "training":
			trainingCats = append(trainingCats, c.Name)
		case "nutrition":
			nutritionCats = append(nutritionCats, c.Name)
		default:
			generalCats = append(generalCats, c.Name)
		}
	}

	msg := "📂 *Категории:*\n\n"

	if len(trainingCats) > 0 {
		msg += "🏋️ *Тренировки:*\n"
		for _, name := range trainingCats {
			msg += fmt.Sprintf("• %s\n", name)
		}
		msg += "\n"
	}

	if len(nutritionCats) > 0 {
		msg += "🍎 *Питание:*\n"
		for _, name := range nutritionCats {
			msg += fmt.Sprintf("• %s\n", name)
		}
		msg += "\n"
	}

	if len(generalCats) > 0 {
		msg += "📋 *Общие:*\n"
		for _, name := range generalCats {
			msg += fmt.Sprintf("• %s\n", name)
		}
	}

	b.sendText(chatID, msg)
}

// Отправка сообщений
func (b *BotApp) sendText(chatID int64, text string) {
	log.Printf("[sendText] chatID=%d, text length=%d", chatID, len(text))

	msg := tgbotapi.NewMessage(chatID, text)

	// Для обычных сообщений тоже включаем Markdown
	// Но экранируем спецсимволы в тексте пользователя
	msg.ParseMode = "Markdown"

	if _, err := b.API.Send(msg); err != nil {
		log.Printf("[sendText] ERROR: %v", err)

		// Если Markdown вызывает ошибку, пробуем отправить без него
		msg2 := tgbotapi.NewMessage(chatID, text)
		msg2.ParseMode = ""
		if _, err2 := b.API.Send(msg2); err2 != nil {
			log.Printf("[sendText] ERROR without Markdown: %v", err2)
		}
	} else {
		log.Printf("[sendText] SUCCESS")
	}
}

func (b *BotApp) editMessage(chatID int64, messageID int, text string, rows [][]tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	editMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, messageID, text, keyboard)
	editMsg.ParseMode = "Markdown"
	b.API.Send(editMsg)
}

// Главное меню
func (b *BotApp) showMainMenu(chatID int64) {
	welcomeMsg := `🏃‍♀️ *Добро пожаловать в Fitness Bot!*

🌟 *Ваш персональный помощник на пути к здоровью и красоте!*

🎯 *Просто нажмите кнопку:*

🏋️ *Тренировки* → Готовые программы упражнений с видеоуроками
🍎 *Питание* → Планы питания с подсчетом калорий
📂 *Категории* → Удобная навигация по материалам
ℹ️ *Помощь* → Инструкция и справка

📅 *Что вас ждет:*
✅ Ежедневные тренировки с пошаговыми инструкциями
✅ Сбалансированное питание с учетом КБЖУ
✅ Недельные меню - готовый рацион на 7 дней
✅ Регулярные обновления - новый контент каждый день

🚀 *Начните прямо сейчас!*
1. Выберите раздел в меню ниже 👇
2. Изучайте программы тренировок
3. Составляйте свое идеальное меню
4. Достигайте результатов вместе с нами!

---

*"Путь в тысячу миль начинается с первого шага"*
*Сделайте свой первый шаг к здоровью прямо сейчас!* 💪`
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏋️ Тренировки"),
			tgbotapi.NewKeyboardButton("🍎 Питание"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 Статистика"),
			tgbotapi.NewKeyboardButton("👤 Профиль"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧮 Калькулятор"),
			tgbotapi.NewKeyboardButton("⚖️ Записать вес"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("ℹ️ Помощь"),
		),
	)
	// Добавляем настройки клавиатуры
	keyboard.ResizeKeyboard = true   // Клавиатура занимает меньше места
	keyboard.OneTimeKeyboard = false // Клавиатура остается постоянно

	msg := tgbotapi.NewMessage(chatID, welcomeMsg)
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"

	b.API.Send(msg)
}

func (b *BotApp) sendTextWithKeyboard(chatID int64, text string, rows [][]tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.API.Send(msg)
}

func (b *BotApp) answerCallback(callbackID string, text string) {
	b.API.Request(tgbotapi.NewCallback(callbackID, text))
}

// parseAdminIDs преобразует строку вида "123,456,789" в срез int64
func ParseAdminIDs(ids string) []int64 {
	var result []int64
	if ids == "" {
		return result
	}
	for _, s := range strings.Split(ids, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64); err == nil {
			result = append(result, id)
		}
	}
	return result
}

func (b *BotApp) authenticateUser(update tgbotapi.Update) (*models.User, error) {
	tgUser := update.Message.From

	user, err := b.userService.GetUserByTelegramID(int64(tgUser.ID))
	if err == nil {
		return user, nil
	}

	// Пользователь не найден — создаём
	return b.userService.CreateUser(service.CreateUserDTO{
		TelegramID: int64(tgUser.ID),
		Name:       tgUser.UserName,
		Role:       "user",
	})
}
func (b *BotApp) isAuthorized(userID int64, requiredRole string) bool {
	if requiredRole == "admin" && b.isAdmin(userID) {
		return true
	}

	user, err := b.userService.GetUserByTelegramID(userID)
	if err != nil {
		return false
	}

	if requiredRole == "user" {
		return true
	}

	return user.Role == requiredRole
}

// Отправка Markdown-сообщений с экранированием спецсимволов
func (b *BotApp) sendMarkdown(chatID int64, text string) {
	// Экранируем спецсимволы MarkdownV2
	escape := func(s string) string {
		specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, c := range specialChars {
			s = strings.ReplaceAll(s, c, "\\"+c)
		}
		return s
	}
	msg := tgbotapi.NewMessage(chatID, escape(text))
	msg.ParseMode = "MarkdownV2"
	b.API.Send(msg)
}

func (b *BotApp) testTrainings(chatID int64) {
	log.Println("[testTrainings] Using hardcoded data")

	// Создаем тестовые данные
	trainings := []*models.TrainingProgram{
		{
			Title:       "Тестовая утренняя зарядка",
			Description: "Базовые упражнения для пробуждения",
			Duration:    15,
		},
		{
			Title:       "Силовая тренировка",
			Description: "Упражнения с весом",
			Duration:    45,
			YouTubeLink: "https://youtube.com/watch?v=test123",
		},
	}

	msg := "🏋️ *ТЕСТ: Доступные тренировки:*\n\n"
	for i, t := range trainings {
		msg += fmt.Sprintf("%d. *%s* - %d мин\n", i+1, t.Title, t.Duration)
		if t.Description != "" {
			msg += fmt.Sprintf("   %s\n", t.Description)
		}
		if t.YouTubeLink != "" {
			msg += fmt.Sprintf("   🎥 Ссылка на YouTube\n")
		}
		msg += "\n"
	}

	b.sendText(chatID, msg)
}

// В main.go или отдельном тестовом файле
func testTrainingFlow(bot *BotApp, chatID int64) {
	log.Println("=== TESTING TRAINING FLOW ===")

	// 1. Проверяем сервис
	trainings, err := bot.trainingService.ListTrainings()
	log.Printf("Service result: %d trainings, error: %v", len(trainings), err)

	// 2. Пробуем отправить простое сообщение
	bot.sendText(chatID, "Тестовое сообщение 123")

	// 3. Пробуем показать тренировки
	bot.showTrainingsForUser(chatID)
}
func (b *BotApp) showNutritionListForSelection(chatID int64) {
	nutritionList, err := b.nutritionService.ListNutrition()
	if err != nil {
		b.sendText(chatID, "❌ Не удалось загрузить список блюд")
		return
	}

	if len(nutritionList) == 0 {
		b.sendText(chatID, "🍎 Блюд пока нет. Сначала добавьте блюда через админ-панель питания.")
		return
	}

	// Экранируем спецсимволы для Markdown
	escape := func(s string) string {
		specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
		for _, c := range specialChars {
			s = strings.ReplaceAll(s, c, "\\"+c)
		}
		return s
	}

	msg := "🍎 *Список блюд (ID - Название - Калории):*\n\n"
	for _, n := range nutritionList {
		msg += fmt.Sprintf("*%d*\\. %s \\- %d ккал\n",
			n.ID, escape(n.Title), n.Calories)
	}

	msg += "\nПри добавлении приема пищи введите ID блюда из этого списка\\."
	b.sendText(chatID, msg)
	b.showMainMenu(chatID)
}

func (b *BotApp) showUserStats(chatID int64) {
	user, err := b.userService.GetUserByTelegramID(chatID)
	if err != nil {
		b.sendText(chatID, "📊 Статистика пока недоступна.\n\nСначала нажмите /start или заполните профиль.")
		return
	}

	logs, err := b.weightService.GetUserHistory(uint(user.ID))
	if err != nil || len(logs) == 0 {
		b.sendText(chatID, "📊 У вас пока нет записей веса.\n\nНажмите «⚖️ Записать вес», чтобы добавить первую запись.")
		return
	}

	var min, max, sum float64
	min = logs[0].Weight
	max = logs[0].Weight

	for _, item := range logs {
		sum += item.Weight
		if item.Weight < min {
			min = item.Weight
		}
		if item.Weight > max {
			max = item.Weight
		}
	}

	start := logs[0].Weight
	current := logs[len(logs)-1].Weight
	avg := sum / float64(len(logs))
	change := start - current

	msg := fmt.Sprintf(
		"📊 *Ваша статистика*\n\n"+
			"⚖️ Текущий вес: *%.1f кг*\n"+
			"📍 Начальный вес: *%.1f кг*\n"+
			"🎯 Желаемый вес: *%.1f кг*\n"+
			"📉 Изменение: *%.1f кг*\n"+
			"📊 Минимум: *%.1f кг*\n"+
			"📊 Максимум: *%.1f кг*\n"+
			"📈 Средний вес: *%.1f кг*\n"+
			"📝 Записей: *%d*",
		current,
		start,
		user.TargetWeight,
		change,
		min,
		max,
		avg,
		len(logs),
	)

	b.sendText(chatID, msg)
}

func (b *BotApp) showProfile(chatID int64) {
	user, err := b.userService.GetUserByTelegramID(chatID)
	if err != nil {
		b.sendText(chatID, "👤 Профиль пока не заполнен.\n\nОткройте Mini App и заполните данные в разделе «Профиль».")
		return
	}

	msg := fmt.Sprintf(
		"👤 *Ваш профиль*\n\n"+
			"Имя: *%s*\n"+
			"Возраст: *%d*\n"+
			"Рост: *%.0f см*\n"+
			"Вес: *%.1f кг*\n"+
			"Цель: *%s*\n"+
			"Активность: *%s*\n"+
			"Уровень: *%s*\n"+
			"Желаемый вес: *%.1f кг*",
		user.Name,
		user.Age,
		user.Height,
		user.Weight,
		user.Goal,
		user.Activity,
		user.FitnessLevel,
		user.TargetWeight,
	)

	b.sendText(chatID, msg)
}

func (b *BotApp) startCalculator(chatID int64) {
	b.sendText(chatID,
		"🧮 *Калькулятор калорий и ИМТ*\n\n"+
			"Введите ваш вес в кг:")

	b.userStates[chatID] = "calc_weight"
	b.calculatorData[chatID] = make(map[string]string)
}

func (b *BotApp) handleCalculatorInput(chatID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) != 6 {
		b.sendText(chatID, "❌ Неверный формат.\n\nПример:\n`55 165 20 female medium loss`")
		return
	}

	weight, err1 := strconv.ParseFloat(parts[0], 64)
	height, err2 := strconv.ParseFloat(parts[1], 64)
	age, err3 := strconv.Atoi(parts[2])
	gender := parts[3]
	activity := parts[4]
	goal := parts[5]

	if err1 != nil || err2 != nil || err3 != nil || weight <= 0 || height <= 0 || age <= 0 {
		b.sendText(chatID, "❌ Проверьте числа.\n\nПример:\n`55 165 20 female medium loss`")
		return
	}

	heightM := height / 100
	bmi := weight / (heightM * heightM)

	genderValue := -161.0
	if gender == "male" {
		genderValue = 5
	}

	bmr := 10*weight + 6.25*height - 5*float64(age) + genderValue

	multiplier := 1.2
	if activity == "medium" {
		multiplier = 1.55
	}
	if activity == "high" {
		multiplier = 1.725
	}

	maintenance := bmr * multiplier
	target := maintenance

	if goal == "loss" {
		target = maintenance - 300
	}
	if goal == "gain" {
		target = maintenance + 300
	}

	status := "Норма"
	if bmi < 18.5 {
		status = "Недостаточный вес"
	} else if bmi >= 25 && bmi < 30 {
		status = "Избыточный вес"
	} else if bmi >= 30 {
		status = "Ожирение"
	}

	msg := fmt.Sprintf(
		"🧮 *Ваш результат*\n\n"+
			"ИМТ: *%.1f* — %s\n"+
			"Базовый обмен: *%.0f ккал/день*\n"+
			"Поддержание: *%.0f ккал/день*\n"+
			"Для цели: *%.0f ккал/день*",
		bmi,
		status,
		bmr,
		maintenance,
		target,
	)

	b.sendText(chatID, msg)
}

func (b *BotApp) finishCalculator(chatID int64, userID int64) {

	data := b.calculatorData[userID]

	weight, _ := strconv.ParseFloat(data["weight"], 64)
	height, _ := strconv.ParseFloat(data["height"], 64)
	age, _ := strconv.Atoi(data["age"])

	gender := data["gender"]
	activity := data["activity"]
	goal := data["goal"]

	heightM := height / 100
	bmi := weight / (heightM * heightM)

	genderValue := -161.0

	if strings.Contains(strings.ToLower(gender), "муж") {
		genderValue = 5
	}

	bmr := 10*weight + 6.25*height - 5*float64(age) + genderValue

	multiplier := 1.2

	if strings.Contains(activity, "Сред") {
		multiplier = 1.55
	}

	if strings.Contains(activity, "Выс") {
		multiplier = 1.725
	}

	maintenance := bmr * multiplier
	target := maintenance

	if strings.Contains(goal, "Пох") {
		target -= 300
	}

	if strings.Contains(goal, "Наб") {
		target += 300
	}

	status := "Норма"

	if bmi < 18.5 {
		status = "Недостаточный вес"
	} else if bmi >= 25 && bmi < 30 {
		status = "Избыточный вес"
	} else if bmi >= 30 {
		status = "Ожирение"
	}

	msg := fmt.Sprintf(
		"🧮 *Ваш результат*\n\n"+
			"📊 ИМТ: *%.1f* — %s\n"+
			"🔥 Базовый обмен: *%.0f ккал*\n"+
			"⚡ Поддержание веса: *%.0f ккал*\n"+
			"🎯 Для вашей цели: *%.0f ккал*",
		bmi,
		status,
		bmr,
		maintenance,
		target,
	)

	b.sendText(chatID, msg)

	delete(b.calculatorData, userID)
}

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

	Admins   []int64
	Handlers map[string]func(tgbotapi.Update)

	trainingService  *service.TrainingService
	nutritionService *service.NutritionService
	categoryService  *service.CategoryService
	userService      *service.UserService

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
		// Обработка CallbackQuery
		if update.CallbackQuery != nil {
			callback := update.CallbackQuery

			// Проверяем, админ или обычный пользователь
			if b.isAuthorized(callback.From.ID, "admin") {
				b.adminHandler.HandleAdminCallback(callback)
			} else {
				// Обычные пользователи не должны получать callback
				b.answerCallback(callback.ID, "⛔ Доступ запрещен")
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		// Обработка команд
		if update.Message.IsCommand() {
			b.handleCommand(update)
			continue
		}

		// Обработка обычных сообщений
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

	// 1. Сначала проверяем состояние админ-панели
	state, isAdminAction := b.adminHandler.GetState(userID)
	if isAdminAction {
		log.Println("Admin action detected")
		// АДМИНСКИЕ ДЕЙСТВИЯ
		b.adminHandler.HandleAdminActions(chatID, userID, state, text)
		return
	}

	// 2. Проверяем, является ли пользователь админом
	if b.isAdmin(userID) {
		log.Println("Admin regular message")
		// Админ, но не в режиме админ-панели
		b.handleAdminRegularMessage(chatID, text)
		return
	}

	// 3. ОБЫЧНЫЕ ПОЛЬЗОВАТЕЛИ
	log.Println("User action")
	b.handleUserActions(chatID, text)
}

// Обработка обычных сообщений админа (не в админ-панели)
func (b *BotApp) handleAdminRegularMessage(chatID int64, text string) {
	log.Printf("[handleAdminRegularMessage] chatID=%d, text='%s'", chatID, text)
	// Админ может вводить специальные команды
	switch text {
	case "/panel":
		b.adminHandler.ShowAdminPanel(chatID)
	case "/trainings":
		b.adminHandler.ShowTrainingsAdmin(chatID)
	case "/nutrition":
		b.adminHandler.ShowNutritionAdmin(chatID)
	case "/categories":
		b.adminHandler.ShowCategoriesAdmin(chatID)
	case "🏋️ Тренировки":
		// Админ тоже может смотреть тренировки как обычный пользователь
		log.Println("[handleAdminRegularMessage] Showing trainings for admin")
		b.showTrainingsForUser(chatID)
	case "🍎 Питание":
		log.Println("[handleAdminRegularMessage] Showing nutrition for admin")
		b.showNutritionForUser(chatID)
	case "📅 Недельное меню":
		b.showWeeklyMenuForUser(chatID)
	case "📂 Категории":
		log.Println("[handleAdminRegularMessage] Showing categories for admin")
		b.showCategoriesForUser(chatID)
	case "ℹ️ Помощь":
		b.sendText(chatID, "🏃‍♀️ Fitness Bot Помощь:\n\nВыберите раздел в меню:\n"+
			"• Тренировки - программы упражнений\n"+
			"• Питание - планы питания\n"+
			"• Категории - фильтрация контента\n\n"+
			"Используйте /start для возврата в меню")
		// В handleAdminRegularMessage добавьте:
	case "/foodlist":
		b.adminHandler.ShowNutritionListForSelection(chatID)
	default:
		// Если админ просто что-то пишет, показываем главное меню
		b.showMainMenu(chatID)
	}
}

// Обработка действий обычных пользователей
// В handleUserActions оставьте только вызовы функций show...ForUser
func (b *BotApp) handleUserActions(chatID int64, text string) {
	log.Printf("[handleUserActions] chatID=%d, text='%s'", chatID, text)

	switch text {
	case "🏋️ Тренировки":
		log.Println("[handleUserActions] Calling showTrainingsForUser")
		b.showTrainingsForUser(chatID)
	case "🍎 Питание":
		log.Println("[handleUserActions] Calling showNutritionForUser")
		b.showNutritionForUser(chatID)
	case "📅 Недельное меню":
		log.Println("[handleUserActions] Calling showWeeklyMenuForUser")
		b.showWeeklyMenuForUser(chatID)
	case "📂 Категории":
		log.Println("[handleUserActions] Calling showCategoriesForUser")
		b.showCategoriesForUser(chatID)
	case "ℹ️ Помощь":
		helpMsg := `📚 *Помощь по использованию Fitness Bot*

*Как пользоваться ботом:*
🏋️ *Тренировки* - Готовые программы упражнений с видео
🍎 *Питание* - Планы питания и недельные меню
📂 *Категории* - Фильтрация контента по темам

*Основные команды:*
/start - Главное меню
/help - Подробная справка

*Советы:*
• Регулярно проверяйте обновления
• Составляйте свое меню на неделю
• Следуйте программам тренировок

*Нужна помощь?*
Используйте команду /help для подробной информации
или свяжитесь с администратором.`

		b.sendText(chatID, helpMsg)
	case "/testtrainings":
		b.testTrainings(chatID)
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
			tgbotapi.NewKeyboardButton("📅 Недельное меню"),
			tgbotapi.NewKeyboardButton("📂 Категории"),
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
}

func (b *BotApp) showWeeklyMenuForUser(chatID int64) {
	log.Printf("[showWeeklyMenuForUser] START for chatID=%d", chatID)

	// Получаем активное недельное меню
	activeMenu, err := b.nutritionService.GetActiveWeeklyMenu()
	if err != nil {
		log.Printf("[showWeeklyMenuForUser] ERROR: %v", err)
		b.sendText(chatID, "❌ Не удалось загрузить недельное меню")
		return
	}

	if activeMenu == nil {
		b.sendText(chatID, "📭 Активное недельное меню еще не создано.\nОжидайте обновлений от администратора!")
		return
	}

	// Загружаем полное меню с днями и приемами пищи
	fullMenu, err := b.nutritionService.GetFullWeeklyMenu(activeMenu.ID)
	if err != nil {
		log.Printf("[showWeeklyMenuForUser] ERROR loading full menu: %v", err)
		b.sendText(chatID, "📅 *"+activeMenu.Name+"*\n\n"+activeMenu.Description)
		return
	}

	// Формируем красивое сообщение
	msg := fmt.Sprintf("📅 *%s*\n\n", fullMenu.Name)
	if fullMenu.Description != "" {
		msg += fmt.Sprintf("%s\n\n", fullMenu.Description)
	}

	msg += fmt.Sprintf("🍽 Всего калорий за неделю: *%d ккал*\n\n", fullMenu.TotalCalories)

	if len(fullMenu.Days) == 0 {
		msg += "📭 Дни меню еще не добавлены\n"
	} else {
		msg += "📋 *Рацион на неделю:*\n\n"

		// Группируем дни по номерам для удобного доступа
		daysMap := make(map[int]models.MenuDay)
		for _, day := range fullMenu.Days {
			daysMap[day.DayNumber] = day
		}

		// Показываем дни от 1 до 7
		for dayNum := 1; dayNum <= 7; dayNum++ {
			if day, exists := daysMap[dayNum]; exists {
				msg += fmt.Sprintf("*%d. %s* - %d ккал\n",
					day.DayNumber, day.DayName, day.TotalCalories)

				if len(day.Meals) > 0 {
					for _, meal := range day.Meals {
						if meal.Nutrition.ID != 0 {
							msg += fmt.Sprintf("   🕐 %s: %s - %s (%d ккал)\n",
								meal.MealTime, meal.MealType,
								meal.Nutrition.Title, meal.Nutrition.Calories)
							if meal.Notes != "" {
								msg += fmt.Sprintf("     📝 %s\n", meal.Notes)
							}
						}
					}
				} else {
					msg += "   📭 Приемы пищи не добавлены\n"
				}
				msg += "\n"
			}
		}
	}

	msg += "\n🍎 *Приятного аппетита!* 🍴"

	b.sendText(chatID, msg)
}

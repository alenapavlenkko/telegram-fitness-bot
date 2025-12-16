package bot

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
	progressService  *service.ProgressService

	// Админ-панель
	adminStates    map[int64]*AdminState
	adminCallbacks map[string]func(*tgbotapi.CallbackQuery)
}

// AdminState хранит состояние админ-панели
type AdminState struct {
	Action   string
	Step     int
	EntityID uint
	TempData map[string]interface{}
}

// Конструктор бота
func NewBotApp(
	token string,
	trainingService *service.TrainingService,
	nutritionService *service.NutritionService,
	categoryService *service.CategoryService,
	userService *service.UserService,
	progressService *service.ProgressService,
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
		progressService:  progressService,
		adminStates:      make(map[int64]*AdminState),
		adminCallbacks:   make(map[string]func(*tgbotapi.CallbackQuery)),
	}

	bot.registerAdminCallbacks()
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
				b.handleAdminCallback(callback)
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
func (b *BotApp) requireAdmin(
	handler func(*tgbotapi.CallbackQuery),
) func(*tgbotapi.CallbackQuery) {

	return func(c *tgbotapi.CallbackQuery) {
		if !b.isAuthorized(c.From.ID, "admin") {
			b.answerCallback(c.ID, "⛔ Нет прав")
			return
		}
		handler(c)
	}
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
		if !b.isAuthorized(int64(update.Message.From.ID), "admin") {
			b.sendText(chatID, "⛔ Недостаточно прав")
			return
		}
		b.showAdminPanel(chatID)
	case "checkdb":
		if b.isAdmin(int64(update.Message.From.ID)) {
			b.checkDatabase(chatID)
		}
	case "foodlist":
		if b.isAdmin(int64(update.Message.From.ID)) {
			b.showNutritionListForSelection(chatID)
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
	state, isAdminAction := b.adminStates[userID]
	if isAdminAction {
		log.Println("Admin action detected")
		// АДМИНСКИЕ ДЕЙСТВИЯ
		b.handleAdminActions(chatID, userID, state, text)
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

// Обработка действий админ-панели
func (b *BotApp) handleAdminActions(chatID, userID int64, state *AdminState, text string) {
	log.Println("ADMIN FSM:", state.Action, "STEP:", state.Step, "TEXT:", text)
	switch state.Action {
	// ==================== Тренировки ====================
	case "add_training":
		if state.Step == 1 {
			state.TempData["title"] = text
			state.Step = 2
			b.sendText(chatID, "Введите длительность (минуты):")
		} else if state.Step == 2 {
			dur, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Пожалуйста, введите число!")
				return
			}
			state.TempData["duration"] = dur
			state.Step = 3
			b.sendText(chatID, "Введите ссылку на YouTube (или оставьте пустым):")
		} else if state.Step == 3 {
			state.TempData["youtube_link"] = text

			_, err := b.trainingService.CreateTraining(service.CreateTrainingDTO{
				Title:       state.TempData["title"].(string),
				Duration:    state.TempData["duration"].(int),
				YouTubeLink: state.TempData["youtube_link"].(string),
				CategoryID:  nil,
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при создании тренировки: "+err.Error())
				return
			}

			b.sendText(chatID, "✅ Тренировка создана")
			delete(b.adminStates, userID)
			b.showTrainingsAdmin(chatID)
		}

	case "edit_training":
		if state.Step == 1 {
			state.TempData["title"] = text
			state.Step = 2
			b.sendText(chatID, "Введите новую длительность (минуты):")
		} else if state.Step == 2 {
			dur, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Пожалуйста, введите число!")
				return
			}
			state.TempData["duration"] = dur
			state.Step = 3
			b.sendText(chatID, "Введите ссылку на YouTube (или оставьте пустым):")
		} else if state.Step == 3 {
			state.TempData["youtube_link"] = text
			_, err := b.trainingService.CreateTraining(service.CreateTrainingDTO{
				Title:       state.TempData["title"].(string),
				Duration:    state.TempData["duration"].(int),
				YouTubeLink: state.TempData["youtube_link"].(string),
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при создании тренировки: "+err.Error())
				return
			}

			b.sendText(chatID, "✅ Тренировка обновлена")
			delete(b.adminStates, userID)
			b.showTrainingsAdmin(chatID)
		}

	// ==================== Питание ====================
	case "add_nutrition":
		switch state.Step {
		case 1:
			state.TempData["title"] = text
			state.Step = 2
			b.sendText(chatID, "Введите описание:")
		case 2:
			state.TempData["description"] = text
			state.Step = 3
			b.sendText(chatID, "Введите калорийность (ккал):")
		case 3:
			calories, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для калорийности!")
				return
			}
			state.TempData["calories"] = calories
			state.Step = 4
			b.sendText(chatID, "Введите белки (г):")
		case 4:
			protein, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для белков!")
				return
			}
			state.TempData["protein"] = protein
			state.Step = 5
			b.sendText(chatID, "Введите углеводы (г):")
		case 5:
			carbs, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для углеводов!")
				return
			}
			state.TempData["carbs"] = carbs
			state.Step = 6
			b.sendText(chatID, "Введите жиры (г):")
		case 6:
			fats, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для жиров!")
				return
			}
			state.TempData["fats"] = fats
			state.Step = 7
			b.sendText(chatID, "Введите ID категории (или 0 если нет):")
		case 7:
			categoryID, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для ID категории!")
				return
			}
			state.TempData["category_id"] = categoryID

			_, err = b.nutritionService.CreateNutrition(service.CreateNutritionDTO{
				Title:       state.TempData["title"].(string),
				Description: state.TempData["description"].(string),
				Calories:    state.TempData["calories"].(int),
				Protein:     state.TempData["protein"].(float64),
				Carbs:       state.TempData["carbs"].(float64),
				Fats:        state.TempData["fats"].(float64),
				CategoryID:  uint(state.TempData["category_id"].(int)),
			})

			if err != nil {
				b.sendText(chatID, "❌ Ошибка при создании питания: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Запись о питании создана")
			}
			delete(b.adminStates, userID)
			b.showNutritionAdmin(chatID)
		}

	case "edit_nutrition":
		switch state.Step {
		case 1:
			state.TempData["title"] = text
			state.Step = 2
			b.sendText(chatID, "Введите новое описание:")
		case 2:
			state.TempData["description"] = text
			state.Step = 3
			b.sendText(chatID, "Введите новую калорийность (ккал):")
		case 3:
			calories, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для калорийности!")
				return
			}
			state.TempData["calories"] = calories
			state.Step = 4
			b.sendText(chatID, "Введите новые белки (г):")
		case 4:
			protein, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для белков!")
				return
			}
			state.TempData["protein"] = protein
			state.Step = 5
			b.sendText(chatID, "Введите новые углеводы (г):")
		case 5:
			carbs, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для углеводов!")
				return
			}
			state.TempData["carbs"] = carbs
			state.Step = 6
			b.sendText(chatID, "Введите новые жиры (г):")
		case 6:
			fats, err := strconv.ParseFloat(text, 64)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для жиров!")
				return
			}
			state.TempData["fats"] = fats
			state.Step = 7
			b.sendText(chatID, "Введите новый ID категории (или 0 если нет):")
		case 7:
			categoryID, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для ID категории!")
				return
			}
			state.TempData["category_id"] = categoryID

			err = b.nutritionService.UpdateNutrition(state.EntityID, service.UpdateNutritionDTO{
				Title:       state.TempData["title"].(string),
				Description: state.TempData["description"].(string),
				Calories:    state.TempData["calories"].(int),
				Protein:     state.TempData["protein"].(float64),
				Carbs:       state.TempData["carbs"].(float64),
				Fats:        state.TempData["fats"].(float64),
				CategoryID:  uint(state.TempData["category_id"].(int)),
			})

			if err != nil {
				b.sendText(chatID, "❌ Ошибка при обновлении питания: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Запись о питании обновлена")
			}
			delete(b.adminStates, userID)
			b.showNutritionAdmin(chatID)
		}

	// ==================== Категории ====================
	case "add_category":
		switch state.Step {
		case 1:
			state.TempData["name"] = text
			state.Step = 2
			b.sendText(chatID, "Введите описание категории:")
		case 2:
			state.TempData["description"] = text
			state.Step = 3
			b.sendText(chatID, "Введите тип (training/nutrition/general):")
		case 3:
			state.TempData["type"] = text

			_, err := b.categoryService.CreateCategory(service.CreateCategoryDTO{
				Name:        state.TempData["name"].(string),
				Description: state.TempData["description"].(string),
				Type:        state.TempData["type"].(string),
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при создании категории: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Категория создана")
			}
			delete(b.adminStates, userID)
			b.showCategoriesAdmin(chatID)
		}

	case "edit_category":
		switch state.Step {
		case 1:
			state.TempData["name"] = text
			state.Step = 2
			b.sendText(chatID, "Введите новое описание:")
		case 2:
			state.TempData["description"] = text
			state.Step = 3
			b.sendText(chatID, "Введите новый тип (training/nutrition/general):")
		case 3:
			state.TempData["type"] = text

			err := b.categoryService.UpdateCategory(state.EntityID, service.UpdateCategoryDTO{
				Name:        state.TempData["name"].(string),
				Description: state.TempData["description"].(string),
				Type:        state.TempData["type"].(string),
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при обновлении категории: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Категория обновлена")
			}
			delete(b.adminStates, userID)
			b.showCategoriesAdmin(chatID)
		}
	case "add_weekly_menu":
		switch state.Step {
		case 1:
			state.TempData["name"] = text
			state.Step = 2
			b.sendText(chatID, "Введите описание меню:")
		case 2:
			state.TempData["description"] = text

			_, err := b.nutritionService.CreateWeeklyMenu(service.CreateWeeklyMenuDTO{
				Name:        state.TempData["name"].(string),
				Description: state.TempData["description"].(string),
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при создании меню: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Недельное меню создано")
			}
			delete(b.adminStates, userID)
			b.showWeeklyMenusAdmin(chatID)
		}

	case "add_day_to_menu":
		switch state.Step {
		case 1:
			dayNum, err := strconv.Atoi(text)
			if err != nil || dayNum < 1 || dayNum > 7 {
				b.sendText(chatID, "❌ Введите номер дня от 1 до 7")
				return
			}
			state.TempData["day_number"] = dayNum
			state.Step = 2

			// Автоматически определяем название дня
			dayNames := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
			state.TempData["day_name"] = dayNames[dayNum-1]
			b.sendText(chatID, fmt.Sprintf("📅 День %d: %s\nТеперь вы можете добавить приемы пищи",
				dayNum, state.TempData["day_name"].(string)))

			// Создаем день
			_, err = b.nutritionService.AddDayToWeeklyMenu(service.AddDayToMenuDTO{
				MenuID:    state.EntityID,
				DayNumber: dayNum,
				DayName:   state.TempData["day_name"].(string),
			})
			if err != nil {
				b.sendText(chatID, "❌ Ошибка при добавлении дня: "+err.Error())
				delete(b.adminStates, userID)
				return
			}
			// Задаем вопрос о добавлении приема пищи
			state.Step = 3
			b.sendText(chatID, "День добавлен! Хотите добавить прием пищи? (Да/Нет)")

		case 3: // Это шаг для ответа на вопрос "Хотите добавить прием пищи?"
			if strings.ToLower(text) == "да" {
				state.Action = "add_meal_to_day"
				state.Step = 1
				b.sendText(chatID, "Выберите тип приема пищи:\n1. Завтрак\n2. Обед\n3. Ужин\n4. Перекус")
			} else {
				delete(b.adminStates, userID)
				b.showWeeklyMenuDetails(chatID, state.EntityID)
			}
		}
	case "add_meal_to_day":
		switch state.Step {
		case 1:
			mealType := ""
			switch text {
			case "1":
				mealType = "Завтрак"
			case "2":
				mealType = "Обед"
			case "3":
				mealType = "Ужин"
			case "4":
				mealType = "Перекус"
			default:
				mealType = text
			}
			state.TempData["meal_type"] = mealType
			state.Step = 2
			b.sendText(chatID, "Введите время приема пищи (например, 09:00):")
		case 2:
			state.TempData["meal_time"] = text
			state.Step = 3
			b.sendText(chatID, "Введите ID блюда из списка питания:")
		case 3: // Когда запрашивается ID блюда
			if text == "/foodlist" {
				b.showNutritionListForSelection(chatID)
				return
			}
			nutritionID, err := strconv.Atoi(text)
			if err != nil {
				b.sendText(chatID, "❌ Введите число для ID блюда. Используйте /foodlist для просмотра списка")
				return
			}
			state.TempData["nutrition_id"] = uint(nutritionID)
			state.Step = 4
			b.sendText(chatID, "Введите заметки (или оставьте пустым):")
		case 4:
			// Получаем ID последнего дня в меню
			menu, err := b.nutritionService.GetFullWeeklyMenu(state.EntityID)
			if err != nil || len(menu.Days) == 0 {
				b.sendText(chatID, "❌ Ошибка при получении дней меню")
				delete(b.adminStates, userID)
				return
			}

			// Берем последний добавленный день
			lastDay := menu.Days[len(menu.Days)-1]

			_, err = b.nutritionService.AddMealToDay(service.AddMealToDayDTO{
				DayID:       lastDay.ID,
				MealType:    state.TempData["meal_type"].(string),
				MealTime:    state.TempData["meal_time"].(string),
				NutritionID: state.TempData["nutrition_id"].(uint),
				Notes:       text,
			})

			if err != nil {
				b.sendText(chatID, "❌ Ошибка при добавлении приема пищи: "+err.Error())
			} else {
				b.sendText(chatID, "✅ Прием пищи добавлен!")
			}

			// Спрашиваем, добавить еще один прием пищи
			state.Step = 5
			b.sendText(chatID, "Хотите добавить еще один прием пищи в этот день? (Да/Нет)")

		case 5:
			if strings.ToLower(text) == "да" {
				state.Step = 1 // Снова спрашиваем тип приема пищи
				b.sendText(chatID, "Выберите тип приема пищи:\n1. Завтрак\n2. Обед\n3. Ужин\n4. Перекус")
			} else {
				delete(b.adminStates, userID)
				b.showWeeklyMenuDetails(chatID, state.EntityID)
			}
		}
	default:
		b.sendText(chatID, "⚠️ Неизвестное действие")
		delete(b.adminStates, userID)
	}
}

// Обработка обычных сообщений админа (не в админ-панели)
func (b *BotApp) handleAdminRegularMessage(chatID int64, text string) {
	log.Printf("[handleAdminRegularMessage] chatID=%d, text='%s'", chatID, text)
	// Админ может вводить специальные команды
	switch text {
	case "/panel":
		b.showAdminPanel(chatID)
	case "/trainings":
		b.showTrainingsAdmin(chatID)
	case "/nutrition":
		b.showNutritionAdmin(chatID)
	case "/categories":
		b.showCategoriesAdmin(chatID)
	case "🏋️ Тренировки":
		// Админ тоже может смотреть тренировки как обычный пользователь
		log.Println("[handleAdminRegularMessage] Showing trainings for admin")
		b.showTrainingsForUser(chatID)
	case "🍎 Питание":
		log.Println("[handleAdminRegularMessage] Showing nutrition for admin")
		b.showNutritionForUser(chatID)
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
		b.showNutritionListForSelection(chatID)
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
			tgbotapi.NewKeyboardButton("📂 Категории"),
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

// Админ-панель
func (b *BotApp) showAdminPanel(chatID int64) {
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🏋️ Тренировки", "admin_trainings"),
			tgbotapi.NewInlineKeyboardButtonData("🍎 Питание", "admin_nutrition"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📂 Категории", "admin_categories"),
			tgbotapi.NewInlineKeyboardButtonData("📅 Недельные меню", "admin_weekly_menus"),
		),
	}
	b.sendTextWithKeyboard(chatID, "⚙️ Панель администратора", rows)
}

func (b *BotApp) sendTextWithKeyboard(chatID int64, text string, rows [][]tgbotapi.InlineKeyboardButton) {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	b.API.Send(msg)
}

// Callback обработка CRUD
func (b *BotApp) handleAdminCallback(callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	chatID := callback.Message.Chat.ID

	// Отвечаем на callback, чтобы убрать "часики"
	b.answerCallback(callback.ID, "")

	log.Printf("Admin callback from %d: %s", callback.From.ID, data)

	// 1. Сначала проверяем зарегистрированные callback
	if callbackFn, ok := b.adminCallbacks[data]; ok {
		callbackFn(callback)
		return
	}

	// 2. Обработка недельных меню
	if strings.HasPrefix(data, "admin_view_weekly_menu_") {
		idStr := strings.TrimPrefix(data, "admin_view_weekly_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID меню")
			return
		}
		b.showWeeklyMenuDetails(chatID, uint(id))
		return
	}

	if strings.HasPrefix(data, "admin_add_day_to_menu_") {
		idStr := strings.TrimPrefix(data, "admin_add_day_to_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID меню")
			return
		}

		b.adminStates[callback.From.ID] = &AdminState{
			Action:   "add_day_to_menu",
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		}
		b.sendText(chatID, "Введите номер дня (1-7, где 1 - понедельник):")
		return
	}

	if strings.HasPrefix(data, "admin_activate_menu_") {
		idStr := strings.TrimPrefix(data, "admin_activate_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID меню")
			return
		}

		err = b.nutritionService.ActivateWeeklyMenu(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Ошибка активации: "+err.Error())
		} else {
			b.sendText(chatID, "✅ Меню активировано")
		}
		b.showWeeklyMenusAdmin(chatID)
		return
	}

	if strings.HasPrefix(data, "admin_delete_weekly_menu_") {
		idStr := strings.TrimPrefix(data, "admin_delete_weekly_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Да, удалить",
					fmt.Sprintf("admin_confirm_delete_weekly_menu_%d", id),
				),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "admin_weekly_menus"),
			),
		}

		b.sendTextWithKeyboard(
			chatID,
			fmt.Sprintf("⚠️ Вы уверены, что хотите удалить недельное меню #%d?", id),
			rows,
		)
		return
	}

	if strings.HasPrefix(data, "admin_confirm_delete_weekly_menu_") {
		idStr := strings.TrimPrefix(data, "admin_confirm_delete_weekly_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		err = b.nutritionService.DeleteWeeklyMenu(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			b.sendText(chatID, "✅ Недельное меню удалено")
		}

		b.showWeeklyMenusAdmin(chatID)
		return
	}

	// 3. Затем обрабатываем префиксы для просмотра тренировок
	if strings.HasPrefix(data, "admin_view_training_") {
		idStr := strings.TrimPrefix(data, "admin_view_training_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID тренировки")
			return
		}

		training, err := b.trainingService.GetTrainingByID(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Тренировка не найдена")
			return
		}

		msg := fmt.Sprintf("🏋️ *%s*\n\nДлительность: %d мин\nID: %d",
			training.Title, training.Duration, training.ID)
		b.sendText(chatID, msg)
		return
	}

	// 4. Обрабатываем префиксы для редактирования/удаления
	if strings.HasPrefix(data, "admin_edit_training_") || strings.HasPrefix(data, "admin_delete_training_") {
		// Извлекаем ID из строки
		parts := strings.Split(data, "_")
		if len(parts) < 4 {
			b.sendText(chatID, "❌ Неверный формат команды")
			return
		}

		idStr := parts[3]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID тренировки")
			return
		}

		if strings.HasPrefix(data, "admin_edit_training_") {
			// Режим редактирования
			b.adminStates[callback.From.ID] = &AdminState{
				Action:   "edit_training",
				EntityID: uint(id),
				Step:     1,
				TempData: make(map[string]interface{}),
			}
			b.sendText(chatID, "✏️ Введите новое название тренировки:")
		} else {
			// Режим удаления - сначала запрашиваем подтверждение
			rows := [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить",
						fmt.Sprintf("admin_confirm_delete_training_%d", id)),
					tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "admin_trainings"),
				},
			}
			b.sendTextWithKeyboard(chatID,
				fmt.Sprintf("⚠️ Вы уверены, что хотите удалить тренировку #%d?", id),
				rows)
		}
		return
	}

	// 5. Обрабатываем подтверждение удаления
	if strings.HasPrefix(data, "admin_confirm_delete_training_") {
		idStr := strings.TrimPrefix(data, "admin_confirm_delete_training_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID тренировки")
			return
		}

		err = b.trainingService.DeleteTraining(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Ошибка при удалении тренировки: "+err.Error())
		} else {
			b.sendText(chatID, "✅ Тренировка удалена")
		}
		b.showTrainingsAdmin(chatID)
		return
	}

	// Обработка редактирования питания
	if strings.HasPrefix(data, "admin_edit_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_edit_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		// Загружаем существующую запись для отображения текущих данных
		nutrition, err := b.nutritionService.GetNutritionByID(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Запись о питании не найдена")
			return
		}

		// Показываем текущие данные и начинаем редактирование
		b.adminStates[callback.From.ID] = &AdminState{
			Action:   "edit_nutrition",
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		}

		msg := fmt.Sprintf("✏️ Редактирование: %s\n\nТекущее название: %s\nВведите новое название:",
			nutrition.Title, nutrition.Title)
		b.sendText(chatID, msg)
		return

	} else if strings.HasPrefix(data, "admin_edit_category_") {
		idStr := strings.TrimPrefix(data, "admin_edit_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		category, err := b.categoryService.GetCategoryByID(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Категория не найдена")
			return
		}

		b.adminStates[callback.From.ID] = &AdminState{
			Action:   "edit_category",
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		}

		msg := fmt.Sprintf("✏️ Редактирование: %s\n\nТекущее название: %s\nВведите новое название:",
			category.Name, category.Name)
		b.sendText(chatID, msg)
		return
	}

	if strings.HasPrefix(data, "admin_view_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_view_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		n, err := b.nutritionService.GetNutritionByID(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Запись не найдена")
			return
		}

		msg := fmt.Sprintf(
			"🍎 *%s*\n\n"+
				"Описание: %s\n"+
				"Калории: %d ккал\n"+
				"Белки: %.1f г\n"+
				"Углеводы: %.1f г\n"+
				"Жиры: %.1f г\n"+
				"ID категории: %d\n\n"+
				"ID: %d",
			n.Title,
			n.Description,
			n.Calories,
			n.Protein,
			n.Carbs,
			n.Fats,
			n.CategoryID,
			n.ID,
		)

		b.sendText(chatID, msg)
		return
	}

	if strings.HasPrefix(data, "admin_view_category_") {
		idStr := strings.TrimPrefix(data, "admin_view_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		c, err := b.categoryService.GetCategoryByID(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Категория не найдена")
			return
		}

		msg := fmt.Sprintf(
			"📂 *%s*\n\n"+
				"Описание: %s\n"+
				"Тип: %s\n"+
				"ID: %d",
			c.Name,
			c.Description,
			c.Type,
			c.ID,
		)

		b.sendText(chatID, msg)
		return
	}
	if strings.HasPrefix(data, "admin_delete_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_delete_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Да, удалить",
					fmt.Sprintf("admin_confirm_delete_nutrition_%d", id),
				),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "admin_nutrition"),
			),
		}

		b.sendTextWithKeyboard(
			chatID,
			fmt.Sprintf("⚠️ Вы уверены, что хотите удалить запись о питании #%d?", id),
			rows,
		)
		return
	}
	if strings.HasPrefix(data, "admin_confirm_delete_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_confirm_delete_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		err = b.nutritionService.DeleteNutrition(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			b.sendText(chatID, "✅ Запись о питании удалена")
		}

		b.showNutritionAdmin(chatID)
		return
	}
	if strings.HasPrefix(data, "admin_delete_category_") {
		idStr := strings.TrimPrefix(data, "admin_delete_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Да, удалить",
					fmt.Sprintf("admin_confirm_delete_category_%d", id),
				),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "admin_categories"),
			),
		}

		b.sendTextWithKeyboard(
			chatID,
			fmt.Sprintf("⚠️ Вы уверены, что хотите удалить категорию #%d?", id),
			rows,
		)
		return
	}
	if strings.HasPrefix(data, "admin_confirm_delete_category_") {
		idStr := strings.TrimPrefix(data, "admin_confirm_delete_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			b.sendText(chatID, "❌ Неверный ID")
			return
		}

		err = b.categoryService.DeleteCategory(uint(id))
		if err != nil {
			b.sendText(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			b.sendText(chatID, "✅ Категория удалена")
		}

		b.showCategoriesAdmin(chatID)
		return
	}

	// 5. Если команда не распознана
	b.sendText(chatID, "⚠️ Неизвестная команда")
}

func (b *BotApp) answerCallback(callbackID string, text string) {
	b.API.Request(tgbotapi.NewCallback(callbackID, text))
}

// CRUD-заглушки (реализовать через сервисы)
func (b *BotApp) showTrainingsAdmin(chatID int64) {
	trainings, err := b.trainingService.ListTrainings()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка при получении тренировок: "+err.Error())
		return
	}

	if len(trainings) == 0 {
		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить тренировку", "admin_add_training"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
			),
		}
		b.sendTextWithKeyboard(chatID, "📭 Тренировок пока нет", rows)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}
	for i, t := range trainings {
		// Кнопка для просмотра деталей
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🏋️ %s (%d мин)", t.Title, t.Duration),
				fmt.Sprintf("admin_view_training_%d", t.ID),
			),
		))

		// Кнопки действий
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать",
				fmt.Sprintf("admin_edit_training_%d", t.ID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑️ Удалить",
				fmt.Sprintf("admin_delete_training_%d", t.ID)),
		))

		// Разделитель (только между элементами)
		if i < len(trainings)-1 {
			separator := strings.Repeat("─", 20)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(separator, "noop"),
			))
		}
	}

	// Кнопка добавления новой тренировки
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("➕ Добавить тренировку", "admin_add_training"),
	))

	// Кнопка возврата в админ-панель
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
	))

	b.sendTextWithKeyboard(chatID, fmt.Sprintf("🏋️ Тренировки (Admin) - всего: %d", len(trainings)), rows)
}

func (b *BotApp) showNutritionAdmin(chatID int64) {
	nutritions, err := b.nutritionService.ListNutrition()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка при получении списка питания")
		return
	}

	if len(nutritions) == 0 {
		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить питание", "admin_add_nutrition"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
			),
		}
		b.sendTextWithKeyboard(chatID, "📭 Записей о питании пока нет", rows)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}

	for _, n := range nutritions {
		// Кнопка просмотра
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("🍎 %s (%d ккал)", n.Title, n.Calories),
				fmt.Sprintf("admin_view_nutrition_%d", n.ID),
			),
		))

		// Кнопки действий
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать",
				fmt.Sprintf("admin_edit_nutrition_%d", n.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑 Удалить",
				fmt.Sprintf("admin_delete_nutrition_%d", n.ID),
			),
		))
	}

	// Кнопки управления
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить питание", "admin_add_nutrition"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
		),
	)

	b.sendTextWithKeyboard(chatID, "🍎 Питание (Admin)", rows)
}

func (b *BotApp) showCategoriesAdmin(chatID int64) {
	categories, err := b.categoryService.ListCategories()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка при получении категорий")
		return
	}

	if len(categories) == 0 {
		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Добавить категорию", "admin_add_category"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
			),
		}
		b.sendTextWithKeyboard(chatID, "📭 Категорий пока нет", rows)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}

	for _, c := range categories {
		// Кнопка просмотра
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("📂 %s (%s)", c.Name, c.Type),
				fmt.Sprintf("admin_view_category_%d", c.ID),
			),
		))

		// Кнопки действий
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Редактировать",
				fmt.Sprintf("admin_edit_category_%d", c.ID),
			),
			tgbotapi.NewInlineKeyboardButtonData(
				"🗑 Удалить",
				fmt.Sprintf("admin_delete_category_%d", c.ID),
			),
		))
	}

	// Управляющие кнопки
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить категорию", "admin_add_category"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
		),
	)

	b.sendTextWithKeyboard(chatID, "📂 Категории (Admin)", rows)
}

// Показать недельные меню (админ)
func (b *BotApp) showWeeklyMenusAdmin(chatID int64) {
	menus, err := b.nutritionService.ListWeeklyMenus()
	if err != nil {
		b.sendText(chatID, "❌ Ошибка при получении меню: "+err.Error())
		return
	}

	if len(menus) == 0 {
		rows := [][]tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("➕ Создать меню", "admin_add_weekly_menu"),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
			),
		}
		b.sendTextWithKeyboard(chatID, "📭 Недельных меню пока нет", rows)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}

	// Показать активное меню
	activeMenu, err := b.nutritionService.GetActiveWeeklyMenu()
	if err == nil && activeMenu != nil {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ АКТИВНО: %s", activeMenu.Name),
				fmt.Sprintf("admin_view_weekly_menu_%d", activeMenu.ID),
			),
		))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Подробнее",
				fmt.Sprintf("admin_view_weekly_menu_%d", activeMenu.ID)),
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить день",
				fmt.Sprintf("admin_add_day_to_menu_%d", activeMenu.ID)),
		))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("────────────", "noop"),
		))
	}

	// Показать все меню
	for _, menu := range menus {
		// Пропускаем активное меню, если оно уже показано
		if activeMenu != nil && menu.ID == activeMenu.ID {
			continue
		}

		status := "🔘"
		if menu.Active {
			status = "✅"
		}

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("%s %s (%d ккал)", status, menu.Name, menu.TotalCalories),
				fmt.Sprintf("admin_view_weekly_menu_%d", menu.ID),
			),
		))

		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📋 Подробнее",
				fmt.Sprintf("admin_view_weekly_menu_%d", menu.ID)),
			tgbotapi.NewInlineKeyboardButtonData("✅ Активировать",
				fmt.Sprintf("admin_activate_menu_%d", menu.ID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить",
				fmt.Sprintf("admin_delete_weekly_menu_%d", menu.ID)),
		))

		if menu.ID != menus[len(menus)-1].ID {
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("────────────", "noop"),
			))
		}
	}

	// Управляющие кнопки
	rows = append(rows,
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Создать меню", "admin_add_weekly_menu"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад в админ-панель", "admin_panel"),
		),
	)

	b.sendTextWithKeyboard(chatID, "📅 Недельные меню (Admin)", rows)
}

// Показать детали недельного меню
func (b *BotApp) showWeeklyMenuDetails(chatID int64, menuID uint) {
	menu, err := b.nutritionService.GetFullWeeklyMenu(menuID)
	if err != nil {
		b.sendText(chatID, "❌ Ошибка при получении меню: "+err.Error())
		return
	}

	msg := fmt.Sprintf("📅 *%s*\n\n", menu.Name)
	if menu.Description != "" {
		msg += fmt.Sprintf("Описание: %s\n\n", menu.Description)
	}

	msg += fmt.Sprintf("🍽 Всего калорий за неделю: *%d ккал*\n", menu.TotalCalories)
	msg += fmt.Sprintf("Статус: ")
	if menu.Active {
		msg += "✅ *АКТИВНО*\n\n"
	} else {
		msg += "🔘 Неактивно\n\n"
	}

	if len(menu.Days) == 0 {
		msg += "📭 Дни не добавлены\n"
	} else {
		msg += "📋 *Дни недели:*\n\n"

		// Исправляем: создаем map с использованием не-указателей
		daysMap := make(map[int]models.MenuDay)
		for _, day := range menu.Days {
			daysMap[day.DayNumber] = day
		}

		for dayNum := 1; dayNum <= 7; dayNum++ {
			if day, exists := daysMap[dayNum]; exists {
				msg += fmt.Sprintf("*%d. %s* - %d ккал\n",
					day.DayNumber, day.DayName, day.TotalCalories)

				if len(day.Meals) > 0 {
					for _, meal := range day.Meals {
						// Проверяем, загружена ли информация о питании
						if meal.Nutrition.ID != 0 {
							msg += fmt.Sprintf("   🕐 %s: %s - %s (%d ккал)\n",
								meal.MealTime, meal.MealType,
								meal.Nutrition.Title, meal.Nutrition.Calories)
						} else {
							// Если питание не загружено, показываем только ID
							msg += fmt.Sprintf("   🕐 %s: %s (ID питания: %d)\n",
								meal.MealTime, meal.MealType, meal.NutritionID)
						}
					}
				} else {
					msg += "   📭 Приемы пищи не добавлены\n"
				}
				msg += "\n"
			}
		}
	}

	// Кнопки управления
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➕ Добавить день",
				fmt.Sprintf("admin_add_day_to_menu_%d", menuID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Активировать",
				fmt.Sprintf("admin_activate_menu_%d", menuID)),
			tgbotapi.NewInlineKeyboardButtonData("🗑 Удалить",
				fmt.Sprintf("admin_delete_weekly_menu_%d", menuID)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад к меню", "admin_weekly_menus"),
		),
	}

	b.sendTextWithKeyboard(chatID, msg, rows)
}

func (b *BotApp) startAddTrainingFlow(chatID int64, userID int64) {
	b.adminStates[userID] = &AdminState{
		Action:   "add_training",
		Step:     1,
		TempData: make(map[string]interface{}),
	}
	b.sendText(chatID, "Введите название тренировки:")
}
func (b *BotApp) startAddNutritionFlow(chatID int64, userID int64) {
	b.adminStates[userID] = &AdminState{
		Action:   "add_nutrition",
		Step:     1,
		TempData: make(map[string]interface{}),
	}
	b.sendText(chatID, "Введите название блюда/продукта:")
}

func (b *BotApp) startAddCategoryFlow(chatID int64, userID int64) {
	b.adminStates[userID] = &AdminState{
		Action:   "add_category",
		Step:     1,
		TempData: make(map[string]interface{}),
	}
	b.sendText(chatID, "Введите название категории:")
}

// Начало потока добавления недельного меню
func (b *BotApp) startAddWeeklyMenuFlow(chatID int64, userID int64) {
	b.adminStates[userID] = &AdminState{
		Action:   "add_weekly_menu",
		Step:     1,
		TempData: make(map[string]interface{}),
	}
	b.sendText(chatID, "Введите название недельного меню:")
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

func (b *BotApp) registerAdminCallbacks() {
	b.adminCallbacks["admin_panel"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			b.showAdminPanel(c.Message.Chat.ID)
		})

	b.adminCallbacks["admin_add_training"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			chatID := c.Message.Chat.ID
			userID := c.From.ID
			b.startAddTrainingFlow(chatID, userID)
		})

	// Для редактирования
	b.adminCallbacks["admin_trainings"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			b.showTrainingsAdmin(c.Message.Chat.ID)
		})

	// Обработчик для кнопки-разделителя (ничего не делает)
	b.adminCallbacks["noop"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			// Просто отвечаем на callback, но ничего не делаем
			b.answerCallback(c.ID, "")
		})

	b.adminCallbacks["admin_nutrition"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			b.showNutritionAdmin(c.Message.Chat.ID)
		})

	b.adminCallbacks["admin_categories"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			b.showCategoriesAdmin(c.Message.Chat.ID)
		})

	b.adminCallbacks["admin_weekly_menus"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			b.showWeeklyMenusAdmin(c.Message.Chat.ID)
		})

	b.adminCallbacks["admin_add_nutrition"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			chatID := c.Message.Chat.ID
			userID := c.From.ID
			b.startAddNutritionFlow(chatID, userID)
		})

	b.adminCallbacks["admin_add_category"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			chatID := c.Message.Chat.ID
			userID := c.From.ID
			b.startAddCategoryFlow(chatID, userID)
		})
	b.adminCallbacks["admin_add_weekly_menu"] =
		b.requireAdmin(func(c *tgbotapi.CallbackQuery) {
			chatID := c.Message.Chat.ID
			userID := c.From.ID
			b.startAddWeeklyMenuFlow(chatID, userID)
		})
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

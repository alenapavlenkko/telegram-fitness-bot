package admin

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AdminHandler обрабатывает все админ-действия
type AdminHandler struct {
	trainingService      *service.TrainingService
	nutritionService     *service.NutritionService
	categoryService      *service.CategoryService
	userService          *service.UserService
	progressService      *service.ProgressService
	Fsm                  *AdminFSM
	sendTextFunc         func(chatID int64, text string)
	sendTextWithKeyboard func(chatID int64, text string, rows [][]tgbotapi.InlineKeyboardButton)

	// Callbacks (переносим из BotApp)
	adminCallbacks map[string]func(*tgbotapi.CallbackQuery)
}

func (ah *AdminHandler) RegisterAdminCallbacks() {
	ah.adminCallbacks = make(map[string]func(*tgbotapi.CallbackQuery))

	ah.adminCallbacks["admin_panel"] = func(c *tgbotapi.CallbackQuery) {
		ah.ShowAdminPanel(c.Message.Chat.ID)
	}

	ah.adminCallbacks["admin_add_training"] = func(c *tgbotapi.CallbackQuery) {
		ah.StartAddTrainingFlow(c.Message.Chat.ID, c.From.ID)
	}

	ah.adminCallbacks["admin_trainings"] = func(c *tgbotapi.CallbackQuery) {
		ah.ShowTrainingsAdmin(c.Message.Chat.ID)
	}

	ah.adminCallbacks["noop"] = func(c *tgbotapi.CallbackQuery) {
		// Ничего не делаем
	}

	ah.adminCallbacks["admin_nutrition"] = func(c *tgbotapi.CallbackQuery) {
		ah.ShowNutritionAdmin(c.Message.Chat.ID)
	}

	ah.adminCallbacks["admin_categories"] = func(c *tgbotapi.CallbackQuery) {
		ah.ShowCategoriesAdmin(c.Message.Chat.ID)
	}

	ah.adminCallbacks["admin_weekly_menus"] = func(c *tgbotapi.CallbackQuery) {
		ah.ShowWeeklyMenusAdmin(c.Message.Chat.ID)
	}

	ah.adminCallbacks["admin_add_nutrition"] = func(c *tgbotapi.CallbackQuery) {
		chatID := c.Message.Chat.ID
		userID := c.From.ID
		ah.StartAddNutritionFlow(chatID, userID)
	}

	ah.adminCallbacks["admin_add_category"] = func(c *tgbotapi.CallbackQuery) {
		chatID := c.Message.Chat.ID
		userID := c.From.ID
		ah.StartAddCategoryFlow(chatID, userID)
	}

	ah.adminCallbacks["admin_add_weekly_menu"] = func(c *tgbotapi.CallbackQuery) {
		chatID := c.Message.Chat.ID
		userID := c.From.ID
		ah.StartAddWeeklyMenuFlow(chatID, userID)
	}
}

// Добавляем в AdminHandler:
func (ah *AdminHandler) ShowAdminPanel(chatID int64) {
	log.Printf("[ADMIN DEBUG] ShowAdminPanel called for chat %d", chatID)
	log.Printf("[ADMIN DEBUG] sendTextFunc is nil? %v", ah.sendTextFunc == nil)
	log.Printf("[ADMIN DEBUG] sendTextWithKeyboard is nil? %v", ah.sendTextWithKeyboard == nil)
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
	log.Printf("[ADMIN DEBUG] Sending panel with %d rows", len(rows))

	if ah.sendTextWithKeyboard != nil {
		ah.sendTextWithKeyboard(chatID, "⚙️ Панель администратора", rows)
		log.Printf("[ADMIN DEBUG] Panel sent successfully")
	} else {
		log.Printf("[ADMIN DEBUG] ERROR: sendTextWithKeyboard is nil!")
	}
}
func (ah *AdminHandler) ShowTrainingsAdmin(chatID int64) {
	trainings, err := ah.trainingService.ListTrainings()
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Ошибка при получении тренировок: "+err.Error())
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
		ah.sendTextWithKeyboard(chatID, "📭 Тренировок пока нет", rows)
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

	ah.sendTextWithKeyboard(chatID, fmt.Sprintf("🏋️ Тренировки (Admin) - всего: %d", len(trainings)), rows)
}

func (ah *AdminHandler) ShowNutritionAdmin(chatID int64) {
	nutritions, err := ah.nutritionService.ListNutrition()
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Ошибка при получении списка питания")
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
		ah.sendTextWithKeyboard(chatID, "📭 Записей о питании пока нет", rows)
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

	ah.sendTextWithKeyboard(chatID, "🍎 Питание (Admin)", rows)
}

func (ah *AdminHandler) ShowCategoriesAdmin(chatID int64) {
	categories, err := ah.categoryService.ListCategories()
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Ошибка при получении категорий")
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
		ah.sendTextWithKeyboard(chatID, "📭 Категорий пока нет", rows)
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

	ah.sendTextWithKeyboard(chatID, "📂 Категории (Admin)", rows)
}

func (ah *AdminHandler) ShowWeeklyMenusAdmin(chatID int64) {
	menus, err := ah.nutritionService.ListWeeklyMenus()
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Ошибка при получении меню: "+err.Error())
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
		ah.sendTextWithKeyboard(chatID, "📭 Недельных меню пока нет", rows)
		return
	}

	rows := [][]tgbotapi.InlineKeyboardButton{}

	// Показать активное меню
	activeMenu, err := ah.nutritionService.GetActiveWeeklyMenu()
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

	ah.sendTextWithKeyboard(chatID, "📅 Недельные меню (Admin)", rows)
}

func (ah *AdminHandler) ShowWeeklyMenuDetails(chatID int64, menuID uint) {
	menu, err := ah.nutritionService.GetFullWeeklyMenu(menuID)
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Ошибка при получении меню: "+err.Error())
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

	ah.sendTextWithKeyboard(chatID, msg, rows)
}

// Также нужно добавить метод showNutritionListForSelection если он используется админом
func (ah *AdminHandler) ShowNutritionListForSelection(chatID int64) {
	nutritionList, err := ah.nutritionService.ListNutrition()
	if err != nil {
		ah.sendTextFunc(chatID, "❌ Не удалось загрузить список блюд")
		return
	}

	if len(nutritionList) == 0 {
		ah.sendTextFunc(chatID, "🍎 Блюд пока нет. Сначала добавьте блюда через админ-панель питания.")
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
	ah.sendTextFunc(chatID, msg)
}
func (ah *AdminHandler) HandleAdminCallback(callback *tgbotapi.CallbackQuery) {
	data := callback.Data
	chatID := callback.Message.Chat.ID

	if data == "admin_cancel" {
		ah.Fsm.DeleteState(callback.From.ID)
		ah.sendTextFunc(callback.Message.Chat.ID, "❌ Действие отменено")
		ah.ShowAdminPanel(callback.Message.Chat.ID)
		return
	}

	// 1. Сначала проверяем зарегистрированные callback
	if callbackFn, ok := ah.adminCallbacks[data]; ok {
		callbackFn(callback)
		return
	}

	// 2. Обработка недельных меню (переносим все if-блоки)
	if strings.HasPrefix(data, "admin_view_weekly_menu_") {
		idStr := strings.TrimPrefix(data, "admin_view_weekly_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID меню")
			return
		}
		ah.ShowWeeklyMenuDetails(chatID, uint(id))
		return
	}

	if strings.HasPrefix(data, "admin_add_day_to_menu_") {
		idStr := strings.TrimPrefix(data, "admin_add_day_to_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID меню")
			return
		}

		state := &AdminState{
			Action:   "add_day_to_menu", // Исправляем на правильный action
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		}
		ah.Fsm.SetState(callback.From.ID, state)
		ah.sendTextFunc(chatID, "Введите номер дня (1-7, где 1 - понедельник):")
		return
	}

	if strings.HasPrefix(data, "admin_activate_menu_") {
		idStr := strings.TrimPrefix(data, "admin_activate_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID меню")
			return
		}

		err = ah.nutritionService.ActivateWeeklyMenu(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка активации: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Меню активировано")
		}
		ah.ShowWeeklyMenusAdmin(chatID)
		return
	}

	if strings.HasPrefix(data, "admin_delete_weekly_menu_") {
		idStr := strings.TrimPrefix(data, "admin_delete_weekly_menu_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
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

		ah.sendTextWithKeyboard(
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
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		err = ah.nutritionService.DeleteWeeklyMenu(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Недельное меню удалено")
		}

		ah.ShowWeeklyMenusAdmin(chatID)
		return
	}

	// 3. Обработка тренировок
	if strings.HasPrefix(data, "admin_view_training_") {
		idStr := strings.TrimPrefix(data, "admin_view_training_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID тренировки")
			return
		}

		training, err := ah.trainingService.GetTrainingByID(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Тренировка не найдена")
			return
		}

		msg := fmt.Sprintf("🏋️ *%s*\n\nДлительность: %d мин\nID: %d",
			training.Title, training.Duration, training.ID)
		ah.sendTextFunc(chatID, msg)
		return
	}

	// 4. Обрабатываем префиксы для редактирования/удаления тренировок
	if strings.HasPrefix(data, "admin_edit_training_") || strings.HasPrefix(data, "admin_delete_training_") {
		parts := strings.Split(data, "_")
		if len(parts) < 4 {
			ah.sendTextFunc(chatID, "❌ Неверный формат команды")
			return
		}

		idStr := parts[3]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID тренировки")
			return
		}

		if strings.HasPrefix(data, "admin_edit_training_") {
			state := &AdminState{
				Action:   "edit_training", // Исправляем на правильный action
				EntityID: uint(id),
				Step:     1,
				TempData: make(map[string]interface{}),
			}
			ah.Fsm.SetState(callback.From.ID, state)
			ah.sendTextFunc(chatID, "✏️ Введите новое название тренировки:")
		} else {
			rows := [][]tgbotapi.InlineKeyboardButton{
				{
					tgbotapi.NewInlineKeyboardButtonData("✅ Да, удалить",
						fmt.Sprintf("admin_confirm_delete_training_%d", id)),
					tgbotapi.NewInlineKeyboardButtonData("❌ Отмена", "admin_trainings"),
				},
			}
			ah.sendTextWithKeyboard(chatID,
				fmt.Sprintf("⚠️ Вы уверены, что хотите удалить тренировку #%d?", id),
				rows)
		}
		return
	}

	// 5. Обрабатываем подтверждение удаления тренировки
	if strings.HasPrefix(data, "admin_confirm_delete_training_") {
		idStr := strings.TrimPrefix(data, "admin_confirm_delete_training_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID тренировки")
			return
		}

		err = ah.trainingService.DeleteTraining(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при удалении тренировки: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Тренировка удалена")
		}
		ah.ShowTrainingsAdmin(chatID)
		return
	}

	// 6. Обработка питания
	if strings.HasPrefix(data, "admin_edit_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_edit_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		nutrition, err := ah.nutritionService.GetNutritionByID(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Запись о питании не найдена")
			return
		}

		ah.Fsm.SetState(callback.From.ID, &AdminState{
			Action:   "edit_nutrition",
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		})

		msg := fmt.Sprintf("✏️ Редактирование: %s\n\nТекущее название: %s\nВведите новое название:",
			nutrition.Title, nutrition.Title)
		ah.sendTextFunc(chatID, msg)
		return

	} else if strings.HasPrefix(data, "admin_edit_category_") {
		idStr := strings.TrimPrefix(data, "admin_edit_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		category, err := ah.categoryService.GetCategoryByID(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Категория не найдена")
			return
		}

		state := &AdminState{
			Action:   "edit_category",
			EntityID: uint(id),
			Step:     1,
			TempData: make(map[string]interface{}),
		}
		ah.Fsm.SetState(callback.From.ID, state)

		msg := fmt.Sprintf("✏️ Редактирование: %s\n\nТекущее название: %s\nВведите новое название:",
			category.Name, category.Name)
		ah.sendTextFunc(chatID, msg)
		return
	}

	if strings.HasPrefix(data, "admin_view_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_view_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		n, err := ah.nutritionService.GetNutritionByID(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Запись не найдена")
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

		ah.sendTextFunc(chatID, msg)
		return
	}

	if strings.HasPrefix(data, "admin_view_category_") {
		idStr := strings.TrimPrefix(data, "admin_view_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		c, err := ah.categoryService.GetCategoryByID(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Категория не найдена")
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

		ah.sendTextFunc(chatID, msg)
		return
	}

	if strings.HasPrefix(data, "admin_delete_nutrition_") {
		idStr := strings.TrimPrefix(data, "admin_delete_nutrition_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
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

		ah.sendTextWithKeyboard(
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
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		err = ah.nutritionService.DeleteNutrition(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Запись о питании удалена")
		}

		ah.ShowNutritionAdmin(chatID)
		return
	}

	if strings.HasPrefix(data, "admin_delete_category_") {
		idStr := strings.TrimPrefix(data, "admin_delete_category_")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Неверный ID")
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

		ah.sendTextWithKeyboard(
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
			ah.sendTextFunc(chatID, "❌ Неверный ID")
			return
		}

		err = ah.categoryService.DeleteCategory(uint(id))
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при удалении: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Категория удалена")
		}

		ah.ShowCategoriesAdmin(chatID)
		return
	}

	// Если команда не распознана
	ah.sendTextFunc(chatID, "⚠️ Неизвестная команда")
}
func (ah *AdminHandler) StartAddTrainingFlow(chatID int64, userID int64) {
	ah.Fsm.SetState(userID, &AdminState{
		Action:   "add_training",
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	ah.sendTextFunc(chatID, "Введите название тренировки:")
}

func (ah *AdminHandler) StartAddNutritionFlow(chatID int64, userID int64) {
	ah.Fsm.SetState(userID, &AdminState{
		Action:   "add_nutrition",
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	ah.sendTextFunc(chatID, "Введите название блюда/продукта:")
}

func (ah *AdminHandler) StartAddCategoryFlow(chatID int64, userID int64) {
	ah.Fsm.SetState(userID, &AdminState{
		Action:   "add_category",
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	ah.sendTextFunc(chatID, "Введите название категории:")
}

func (ah *AdminHandler) StartAddWeeklyMenuFlow(chatID int64, userID int64) {
	ah.Fsm.SetState(userID, &AdminState{
		Action:   "add_weekly_menu",
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	ah.sendTextFunc(chatID, "Введите название недельного меню:")
}
func NewAdminHandler(
	trainingService *service.TrainingService,
	nutritionService *service.NutritionService,
	categoryService *service.CategoryService,
	userService *service.UserService,
	progressService *service.ProgressService,
	sendText func(int64, string),
	sendTextWithKeyboard func(int64, string, [][]tgbotapi.InlineKeyboardButton),
) *AdminHandler {
	// Исправленная версия:
	handler := &AdminHandler{
		trainingService:      trainingService,
		nutritionService:     nutritionService,
		categoryService:      categoryService,
		userService:          userService,
		progressService:      progressService,
		Fsm:                  NewAdminFSM(),
		sendTextFunc:         sendText,
		sendTextWithKeyboard: sendTextWithKeyboard,
		adminCallbacks:       make(map[string]func(*tgbotapi.CallbackQuery)),
	}

	handler.RegisterAdminCallbacks()

	return handler
}

// ==================== МЕТОДЫ ДЛЯ СТАРТА ПОТОКОВ ====================

func (h *AdminHandler) StartEditCategoryFlow(chatID, userID int64, categoryID uint) {
	h.Fsm.SetState(userID, &AdminState{
		Action:   "edit_category",
		EntityID: categoryID,
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	h.sendTextFunc(chatID, fmt.Sprintf("✏️ Редактирование категории #%d\nВведите новое название:", categoryID))
}

func (h *AdminHandler) StartEditNutritionFlow(chatID, userID int64, nutritionID uint) {
	h.Fsm.SetState(userID, &AdminState{
		Action:   "edit_nutrition",
		EntityID: nutritionID,
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	h.sendTextFunc(chatID, fmt.Sprintf("✏️ Редактирование питания #%d\nВведите новое название:", nutritionID))
}

func (h *AdminHandler) StartEditTrainingFlow(chatID, userID int64, trainingID uint) {
	h.Fsm.SetState(userID, &AdminState{
		Action:   "edit_training",
		EntityID: trainingID,
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	h.sendTextFunc(chatID, fmt.Sprintf("✏️ Редактирование тренировки #%d\nВведите новое название:", trainingID))
}

func (h *AdminHandler) StartAddDayToMenuFlow(chatID, userID int64, menuID uint) {
	h.Fsm.SetState(userID, &AdminState{
		Action:   "add_day_to_menu",
		EntityID: menuID,
		Step:     1,
		TempData: make(map[string]interface{}),
	})
	h.sendTextFunc(chatID, "Введите номер дня (1-7, где 1 - понедельник):")
}

// ==================== БАЗОВЫЕ МЕТОДЫ ====================

func (h *AdminHandler) GetState(userID int64) (*AdminState, bool) {
	return h.Fsm.GetState(userID)
}

func (h *AdminHandler) SetState(userID int64, state *AdminState) {
	h.Fsm.SetState(userID, state)
}

func (h *AdminHandler) DeleteState(userID int64) {
	h.Fsm.DeleteState(userID)
}

func (ah *AdminHandler) HandleAdminActions(chatID, userID int64, state *AdminState, text string) {
	log.Println("ADMIN FSM:", state.Action, "STEP:", state.Step, "TEXT:", text)

	// Проверяем команду отмены
	if text == "/cancel" || text == "отмена" || text == "cancel" {
		ah.Fsm.DeleteState(userID)
		ah.sendTextFunc(chatID, "❌ Действие отменено")
		ah.ShowAdminPanel(chatID)
		return
	}

	switch state.Action {
	// ==================== Тренировки ====================
	case "add_training":
		ah.handleAddTraining(chatID, userID, state, text)
	case "edit_training":
		ah.handleEditTraining(chatID, userID, state, text)

	// ==================== Питание ====================
	case "add_nutrition":
		ah.handleAddNutrition(chatID, userID, state, text)
	case "edit_nutrition":
		ah.handleEditNutrition(chatID, userID, state, text)

	// ==================== Категории ====================
	case "add_category":
		ah.handleAddCategory(chatID, userID, state, text)
	case "edit_category":
		ah.handleEditCategory(chatID, userID, state, text)

	// ==================== Недельные меню ====================
	case "add_weekly_menu":
		ah.handleAddWeeklyMenu(chatID, userID, state, text)
	case "add_day_to_menu":
		ah.handleAddDayToMenu(chatID, userID, state, text)
	case "add_meal_to_day":
		ah.handleAddMealToDay(chatID, userID, state, text)

	default:
		ah.sendTextFunc(chatID, "⚠️ Неизвестное действие")
		ah.Fsm.DeleteState(userID)
	}
}

// ==================== ТРЕНИРОВКИ ====================

func (ah *AdminHandler) handleAddTraining(chatID, userID int64, state *AdminState, text string) {
	if state.Step == 1 {
		state.TempData["title"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите длительность (минуты):")
	} else if state.Step == 2 {
		dur, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Пожалуйста, введите число!")
			return
		}
		state.TempData["duration"] = dur
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите ссылку на YouTube (или оставьте пустым):")
	} else if state.Step == 3 {
		state.TempData["youtube_link"] = text
		state.Step = 4
		ah.sendTextFunc(chatID, "Введите описание тренировки (или оставьте пустым):")
	} else if state.Step == 4 {
		state.TempData["description"] = text

		categoryID := state.TempData["category_id"]
		var catIDPtr *uint
		if categoryID != nil {
			if catID, ok := categoryID.(uint); ok && catID > 0 {
				catIDPtr = &catID
			}
		}

		_, err := ah.trainingService.CreateTraining(service.CreateTrainingDTO{
			Title:       state.TempData["title"].(string),
			Duration:    state.TempData["duration"].(int),
			YouTubeLink: state.TempData["youtube_link"].(string),
			Description: state.TempData["description"].(string),
			CategoryID:  catIDPtr,
		})
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при создании тренировки: "+err.Error())
			return
		}

		ah.sendTextFunc(chatID, "✅ Тренировка создана")
		ah.Fsm.DeleteState(userID)
		ah.ShowTrainingsAdmin(chatID)
	}
}

func (ah *AdminHandler) handleEditTraining(chatID, userID int64, state *AdminState, text string) {
	if state.Step == 1 {
		state.TempData["title"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите новую длительность (минуты):")
	} else if state.Step == 2 {
		dur, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Пожалуйста, введите число!")
			return
		}
		state.TempData["duration"] = dur
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите новую ссылку на YouTube (или оставьте пустым):")
	} else if state.Step == 3 {
		state.TempData["youtube_link"] = text
		state.Step = 4
		ah.sendTextFunc(chatID, "Введите новое описание (или оставьте пустым):")
	} else if state.Step == 4 {
		state.TempData["description"] = text

		err := ah.trainingService.UpdateTraining(state.EntityID, service.UpdateTrainingDTO{
			Title:       state.TempData["title"].(string),
			Duration:    state.TempData["duration"].(int),
			YouTubeLink: state.TempData["youtube_link"].(string),
			Description: state.TempData["description"].(string),
		})

		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при обновлении тренировки: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Тренировка обновлена")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowTrainingsAdmin(chatID)
	}
}

// ==================== ПИТАНИЕ ====================

func (ah *AdminHandler) handleAddNutrition(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		state.TempData["title"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите описание:")
	case 2:
		state.TempData["description"] = text
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите калорийность (ккал):")
	case 3:
		calories, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для калорийности!")
			return
		}
		state.TempData["calories"] = calories
		state.Step = 4
		ah.sendTextFunc(chatID, "Введите белки (г):")
	case 4:
		protein, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для белков!")
			return
		}
		state.TempData["protein"] = protein
		state.Step = 5
		ah.sendTextFunc(chatID, "Введите углеводы (г):")
	case 5:
		carbs, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для углеводов!")
			return
		}
		state.TempData["carbs"] = carbs
		state.Step = 6
		ah.sendTextFunc(chatID, "Введите жиры (г):")
	case 6:
		fats, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для жиров!")
			return
		}
		state.TempData["fats"] = fats
		state.Step = 7
		ah.sendTextFunc(chatID, "Введите ID категории (или 0 если нет):")
	case 7:
		categoryID, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для ID категории!")
			return
		}
		state.TempData["category_id"] = uint(categoryID)

		_, err = ah.nutritionService.CreateNutrition(service.CreateNutritionDTO{
			Title:       state.TempData["title"].(string),
			Description: state.TempData["description"].(string),
			Calories:    state.TempData["calories"].(int),
			Protein:     state.TempData["protein"].(float64),
			Carbs:       state.TempData["carbs"].(float64),
			Fats:        state.TempData["fats"].(float64),
			CategoryID:  state.TempData["category_id"].(uint),
		})

		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при создании питания: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Запись о питании создана")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowNutritionAdmin(chatID)
	}
}

func (ah *AdminHandler) handleEditNutrition(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		state.TempData["title"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите новое описание:")
	case 2:
		state.TempData["description"] = text
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите новую калорийность (ккал):")
	case 3:
		calories, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для калорийности!")
			return
		}
		state.TempData["calories"] = calories
		state.Step = 4
		ah.sendTextFunc(chatID, "Введите новые белки (г):")
	case 4:
		protein, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для белков!")
			return
		}
		state.TempData["protein"] = protein
		state.Step = 5
		ah.sendTextFunc(chatID, "Введите новые углеводы (г):")
	case 5:
		carbs, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для углеводов!")
			return
		}
		state.TempData["carbs"] = carbs
		state.Step = 6
		ah.sendTextFunc(chatID, "Введите новые жиры (г):")
	case 6:
		fats, err := strconv.ParseFloat(text, 64)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для жиров!")
			return
		}
		state.TempData["fats"] = fats
		state.Step = 7
		ah.sendTextFunc(chatID, "Введите новый ID категории (или 0 если нет):")
	case 7:
		categoryID, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для ID категории!")
			return
		}
		state.TempData["category_id"] = uint(categoryID)

		err = ah.nutritionService.UpdateNutrition(state.EntityID, service.UpdateNutritionDTO{
			Title:       state.TempData["title"].(string),
			Description: state.TempData["description"].(string),
			Calories:    state.TempData["calories"].(int),
			Protein:     state.TempData["protein"].(float64),
			Carbs:       state.TempData["carbs"].(float64),
			Fats:        state.TempData["fats"].(float64),
			CategoryID:  state.TempData["category_id"].(uint),
		})

		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при обновлении питания: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Запись о питании обновлена")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowNutritionAdmin(chatID)
	}
}

// ==================== КАТЕГОРИИ ====================

func (ah *AdminHandler) handleAddCategory(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		state.TempData["name"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите описание категории:")
	case 2:
		state.TempData["description"] = text
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите тип (training/nutrition/general):")
	case 3:
		state.TempData["type"] = text

		_, err := ah.categoryService.CreateCategory(service.CreateCategoryDTO{
			Name:        state.TempData["name"].(string),
			Description: state.TempData["description"].(string),
			Type:        state.TempData["type"].(string),
		})
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при создании категории: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Категория создана")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowCategoriesAdmin(chatID)
	}
}

func (ah *AdminHandler) handleEditCategory(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		state.TempData["name"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите новое описание:")
	case 2:
		state.TempData["description"] = text
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите новый тип (training/nutrition/general):")
	case 3:
		state.TempData["type"] = text

		err := ah.categoryService.UpdateCategory(state.EntityID, service.UpdateCategoryDTO{
			Name:        state.TempData["name"].(string),
			Description: state.TempData["description"].(string),
			Type:        state.TempData["type"].(string),
		})
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при обновлении категории: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Категория обновлена")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowCategoriesAdmin(chatID)
	}
}

// ==================== НЕДЕЛЬНЫЕ МЕНЮ ====================

func (ah *AdminHandler) handleAddWeeklyMenu(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		state.TempData["name"] = text
		state.Step = 2
		ah.sendTextFunc(chatID, "Введите описание меню:")
	case 2:
		state.TempData["description"] = text

		_, err := ah.nutritionService.CreateWeeklyMenu(service.CreateWeeklyMenuDTO{
			Name:        state.TempData["name"].(string),
			Description: state.TempData["description"].(string),
		})
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при создании меню: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Недельное меню создано")
		}
		ah.Fsm.DeleteState(userID)
		ah.ShowWeeklyMenusAdmin(chatID)
	}
}

func (ah *AdminHandler) handleAddDayToMenu(chatID, userID int64, state *AdminState, text string) {
	switch state.Step {
	case 1:
		dayNum, err := strconv.Atoi(text)
		if err != nil || dayNum < 1 || dayNum > 7 {
			ah.sendTextFunc(chatID, "❌ Введите номер дня от 1 до 7")
			return
		}
		state.TempData["day_number"] = dayNum
		state.Step = 2

		// Автоматически определяем название дня
		dayNames := []string{"Понедельник", "Вторник", "Среда", "Четверг", "Пятница", "Суббота", "Воскресенье"}
		state.TempData["day_name"] = dayNames[dayNum-1]
		ah.sendTextFunc(chatID, fmt.Sprintf("📅 День %d: %s\nТеперь вы можете добавить приемы пищи",
			dayNum, state.TempData["day_name"].(string)))

		// Создаем день
		_, err = ah.nutritionService.AddDayToWeeklyMenu(service.AddDayToMenuDTO{
			MenuID:    state.EntityID,
			DayNumber: dayNum,
			DayName:   state.TempData["day_name"].(string),
		})
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при добавлении дня: "+err.Error())
			ah.Fsm.DeleteState(userID)
			return
		}
		// Задаем вопрос о добавлении приема пищи
		state.Step = 3
		ah.sendTextFunc(chatID, "День добавлен! Хотите добавить прием пищи? (Да/Нет)")

	case 3: // Это шаг для ответа на вопрос "Хотите добавить прием пищи?"
		if strings.ToLower(text) == "да" {
			state.Action = "add_meal_to_day"
			state.Step = 1
			ah.sendTextFunc(chatID, "Выберите тип приема пищи:\n1. Завтрак\n2. Обед\n3. Ужин\n4. Перекус")
		} else {
			ah.Fsm.DeleteState(userID)
			ah.ShowWeeklyMenuDetails(chatID, state.EntityID)
		}
	}
}

func (ah *AdminHandler) handleAddMealToDay(chatID, userID int64, state *AdminState, text string) {
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
		ah.sendTextFunc(chatID, "Введите время приема пищи (например, 09:00):")
	case 2:
		state.TempData["meal_time"] = text
		state.Step = 3
		ah.sendTextFunc(chatID, "Введите ID блюда из списка питания (используйте /foodlist для просмотра):")
	case 3: // Когда запрашивается ID блюда
		if text == "/foodlist" {
			ah.ShowNutritionListForSelection(chatID)
			return
		}
		nutritionID, err := strconv.Atoi(text)
		if err != nil {
			ah.sendTextFunc(chatID, "❌ Введите число для ID блюда. Используйте /foodlist для просмотра списка")
			return
		}
		state.TempData["nutrition_id"] = uint(nutritionID)
		state.Step = 4
		ah.sendTextFunc(chatID, "Введите заметки (или оставьте пустым):")
	case 4:
		// Получаем ID последнего дня в меню
		menu, err := ah.nutritionService.GetFullWeeklyMenu(state.EntityID)
		if err != nil || len(menu.Days) == 0 {
			ah.sendTextFunc(chatID, "❌ Ошибка при получении дней меню")
			ah.Fsm.DeleteState(userID)
			return
		}

		// Берем последний добавленный день
		lastDay := menu.Days[len(menu.Days)-1]

		_, err = ah.nutritionService.AddMealToDay(service.AddMealToDayDTO{
			DayID:       lastDay.ID,
			MealType:    state.TempData["meal_type"].(string),
			MealTime:    state.TempData["meal_time"].(string),
			NutritionID: state.TempData["nutrition_id"].(uint),
			Notes:       text,
		})

		if err != nil {
			ah.sendTextFunc(chatID, "❌ Ошибка при добавлении приема пищи: "+err.Error())
		} else {
			ah.sendTextFunc(chatID, "✅ Прием пищи добавлен!")
		}

		// Спрашиваем, добавить еще один прием пищи
		state.Step = 5
		ah.sendTextFunc(chatID, "Хотите добавить еще один прием пищи в этот день? (Да/Нет)")

	case 5:
		if strings.ToLower(text) == "да" {
			state.Step = 1 // Снова спрашиваем тип приема пищи
			ah.sendTextFunc(chatID, "Выберите тип приема пищи:\n1. Завтрак\n2. Обед\n3. Ужин\n4. Перекус")
		} else {
			ah.Fsm.DeleteState(userID)
			ah.ShowWeeklyMenuDetails(chatID, state.EntityID)
		}
	}
}

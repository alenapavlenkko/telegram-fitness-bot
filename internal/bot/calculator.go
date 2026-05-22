package bot

import (
	"fmt"
	"math"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startCalculator запускает калькулятор
func (b *BotApp) startCalculator(
	chatID int64,
	userID int64,
) {

	b.userStates[userID] = StateCalcWeight
	b.calculatorData[userID] = make(map[string]string)

	b.sendText(
		chatID,
		"🧮 *Калькулятор калорий и ИМТ*\n\n"+
			"Введите ваш вес в кг:",
	)
}

// handleCalculatorFlow FSM калькулятора
func (b *BotApp) handleCalculatorFlow(
	update tgbotapi.Update,
) {

	chatID := update.Message.Chat.ID
	userID := int64(update.Message.From.ID)
	text := update.Message.Text

	state := b.userStates[userID]

	switch state {

	// ========================================
	// ВЕС
	// ========================================

	case StateCalcWeight:

		b.calculatorData[userID]["weight"] = text

		b.userStates[userID] = StateCalcHeight

		b.sendText(
			chatID,
			"📏 Введите рост в см:",
		)

	// ========================================
	// РОСТ
	// ========================================

	case StateCalcHeight:

		b.calculatorData[userID]["height"] = text

		b.userStates[userID] = StateCalcAge

		b.sendText(
			chatID,
			"🎂 Введите возраст:",
		)

	// ========================================
	// ВОЗРАСТ
	// ========================================

	case StateCalcAge:

		b.calculatorData[userID]["age"] = text

		b.userStates[userID] = StateCalcGender

		keyboard := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("👩 Женский"),
				tgbotapi.NewKeyboardButton("👨 Мужской"),
			),
		)

		msg := tgbotapi.NewMessage(
			chatID,
			"Выберите пол:",
		)

		msg.ReplyMarkup = keyboard

		b.API.Send(msg)

	// ========================================
	// ПОЛ
	// ========================================

	case StateCalcGender:

		b.calculatorData[userID]["gender"] = text

		b.userStates[userID] = StateCalcActivity

		keyboard := tgbotapi.NewReplyKeyboard(

			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🛌 Низкая"),
				tgbotapi.NewKeyboardButton("🚶 Средняя"),
			),

			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("🏃 Высокая"),
			),
		)

		msg := tgbotapi.NewMessage(
			chatID,
			"Выберите активность:",
		)

		msg.ReplyMarkup = keyboard

		b.API.Send(msg)

	// ========================================
	// АКТИВНОСТЬ
	// ========================================

	case StateCalcActivity:

		b.calculatorData[userID]["activity"] = text

		b.userStates[userID] = StateCalcGoal

		keyboard := tgbotapi.NewReplyKeyboard(

			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("📉 Похудение"),
				tgbotapi.NewKeyboardButton("⚖️ Поддержание"),
			),

			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButton("📈 Набор массы"),
			),
		)

		msg := tgbotapi.NewMessage(
			chatID,
			"Выберите цель:",
		)

		msg.ReplyMarkup = keyboard

		b.API.Send(msg)

	// ========================================
	// ЦЕЛЬ
	// ========================================

	case StateCalcGoal:

		b.calculatorData[userID]["goal"] = text

		b.finishCalculator(chatID, userID)
	}
}

// finishCalculator завершает расчет
func (b *BotApp) finishCalculator(
	chatID int64,
	userID int64,
) {

	data := b.calculatorData[userID]

	weight, _ := strconv.ParseFloat(data["weight"], 64)
	height, _ := strconv.ParseFloat(data["height"], 64)
	age, _ := strconv.Atoi(data["age"])

	gender := data["gender"]
	activity := data["activity"]
	goal := data["goal"]

	// ========================================
	// BMI
	// ========================================

	heightM := height / 100

	bmi := weight / math.Pow(heightM, 2)

	bmiText := "Норма"

	if bmi < 18.5 {
		bmiText = "Недостаточный вес"
	}

	if bmi >= 25 {
		bmiText = "Избыточный вес"
	}

	// ========================================
	// BMR
	// ========================================

	bmr := 10*weight + 6.25*height - 5*float64(age)

	if gender == "👨 Мужской" {
		bmr += 5
	} else {
		bmr -= 161
	}

	// ========================================
	// ACTIVITY
	// ========================================

	activityMultiplier := 1.2

	switch activity {

	case "🚶 Средняя":
		activityMultiplier = 1.55

	case "🏃 Высокая":
		activityMultiplier = 1.725
	}

	maintenance := bmr * activityMultiplier

	targetCalories := maintenance

	switch goal {

	case "📉 Похудение":
		targetCalories -= 300

	case "📈 Набор массы":
		targetCalories += 300
	}
	// ========================================
	// СОХРАНЯЕМ ПРОФИЛЬ
	// ========================================

	user, err := b.userService.GetUserByTelegramID(userID)

	if err == nil {

		user.Age = age
		user.Height = height
		user.Weight = weight

		user.Goal = goal
		user.Activity = activity

		// Автоматически определяем уровень
		if activity == "🏃 Высокая" {

			user.FitnessLevel = "Продвинутый"

		} else if activity == "🚶 Средняя" {

			user.FitnessLevel = "Средний"

		} else {

			user.FitnessLevel = "Начальный"
		}

		// Целевой вес
		targetWeight := weight

		switch goal {

		case "📉 Похудение":
			targetWeight = weight - 3

		case "📈 Набор массы":
			targetWeight = weight + 3
		}

		user.TargetWeight = targetWeight

		// Сохраняем
		b.userService.UpdateUser(user)
	}

	// ========================================
	// RESULT
	// ========================================

	result := fmt.Sprintf(
		"🧮 *Ваш результат*\n\n"+
			"📊 ИМТ: *%.1f* — %s\n"+
			"🔥 Базовый обмен: *%.0f ккал*\n"+
			"⚡ Поддержание веса: *%.0f ккал*\n"+
			"🎯 Для вашей цели: *%.0f ккал*",
		bmi,
		bmiText,
		bmr,
		maintenance,
		targetCalories,
	)

	b.sendText(chatID, result)

	// Очищаем данные
	delete(b.userStates, userID)
	delete(b.calculatorData, userID)

	// Возвращаем главное меню
	b.showMainMenu(chatID)
}

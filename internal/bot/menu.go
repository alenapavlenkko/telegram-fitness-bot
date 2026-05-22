package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// showMainMenu показывает главное меню
func (b *BotApp) showMainMenu(chatID int64) {

	keyboard := tgbotapi.NewReplyKeyboard(

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🏋️ Тренировки"),
			tgbotapi.NewKeyboardButton("🍎 Питание"),
		),

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📂 Категории"),
			tgbotapi.NewKeyboardButton("📊 Статистика"),
		),

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👤 Профиль"),
			tgbotapi.NewKeyboardButton("⚖️ Записать вес"),
		),

		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧮 Калькулятор"),
		),
	)

	msg := tgbotapi.NewMessage(
		chatID,
		`🏃‍♀️ Добро пожаловать в Fitness Bot!

🌟 Ваш персональный фитнес-помощник с расширенным функционалом для тренировок, питания и контроля прогресса!

━━━━━━━━━━━━━━━

🎯 Возможности бота:

🏋️ Тренировки
• Готовые программы тренировок
• Упражнения с видеоуроками
• Разделение по уровням сложности
• Подбор тренировок по категориям
• Пошаговые инструкции выполнения

🍎 Питание
• Планы питания с учетом КБЖУ
• Подсчет калорий
• Сбалансированные рационы
• Недельные меню
• Информация о белках, жирах и углеводах

🧮 Умный калькулятор
• Расчет индекса массы тела (ИМТ)
• Подсчет суточной нормы калорий
• Расчет калорий для похудения, поддержания и набора массы
• Учет пола, возраста, роста и активности
• Персональные рекомендации

📊 Статистика и прогресс
• Сохранение истории веса
• Анализ изменений
• Отслеживание результатов
• Контроль прогресса похудения или набора массы

👤 Профиль пользователя
• Персональные данные
• Цели тренировок
• Уровень подготовки
• Параметры пользователя

━━━━━━━━━━━━━━━

📅 Что вас ждет:

✅ Ежедневные тренировки
✅ Готовые планы питания
✅ Поддержка здорового образа жизни
✅ Удобная навигация
✅ Красивый и понятный интерфейс
✅ Регулярные обновления контента
✅ Быстрый доступ через Telegram
✅ Персонализированные расчеты

━━━━━━━━━━━━━━━

🚀 Начните прямо сейчас!

1️⃣ Выберите нужный раздел в меню 👇
2️⃣ Получите персональные рекомендации
3️⃣ Следите за своим прогрессом
4️⃣ Достигайте целей вместе с Fitness Bot!

━━━━━━━━━━━━━━━

💪 Fitness Bot — ваш надежный помощник
на пути к здоровью, красоте и отличной форме!`,
	)

	msg.ReplyMarkup = keyboard
	msg.ParseMode = "Markdown"

	b.API.Send(msg)
}

// sendText отправляет сообщение
func (b *BotApp) sendText(
	chatID int64,
	text string,
) {

	msg := tgbotapi.NewMessage(
		chatID,
		text,
	)

	msg.ParseMode = "Markdown"

	_, err := b.API.Send(msg)

	if err != nil {
		log.Printf(
			"sendText error: %v",
			err,
		)
	}
}

// sendTextWithKeyboard отправляет сообщение с inline keyboard
func (b *BotApp) sendTextWithKeyboard(
	chatID int64,
	text string,
	rows [][]tgbotapi.InlineKeyboardButton,
) {

	keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	b.API.Send(msg)
}

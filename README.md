cat > README.md << 'EOF'
# 🏋️‍♀️ Fitness Bot - Telegram Fitness Assistant

Телеграм бот для управления тренировками и питанием с полноценной админ-панелью.

## 🚀 Возможности

### Для пользователей:
- 📋 Просмотр тренировок с видеоуроками
- 🍎 Планы питания с подсчетом КБЖУ
- 📅 Недельные меню с автоматическим калоражем
- 📂 Категории для фильтрации контента
- ⭐ Ежедневные рекомендации

### Для администраторов:
- ⚙️ Полный CRUD для всех сущностей (тренировки, питание, категории)
- 📊 Управление недельными меню (создание, активация, наполнение днями)
- 👥 Управление пользователями и их доступом
- 🔄 FSM (Finite State Machine) для многошаговых действий
- 📈 Просмотр статистики использования

## 🛠 Технологии

- **Backend:** Go 1.21+
- **Database:** PostgreSQL 15+
- **ORM:** GORM
- **Bot Framework:** go-telegram-bot-api/v5
- **Containerization:** Docker & Docker Compose
- **Environment:** godotenv
- **Message Formatting:** Markdown + HTML
- **Architecture:** Clean Architecture + Repository Pattern

## 📁 Структура проекта
telegramfitnes/
├── cmd/
│ └── bot/
│ └── main.go # Точка входа приложения
├── internal/
│ ├── bot/ # Логика Telegram бота
│ │ └── bot.go # Основная логика бота и обработчики
│ ├── admin/ # Админ-панель (Telegram-based)
│ │ ├── handler.go # Обработчик админ-действий
│ │ └── state.go # FSM для админских workflow
│ ├── models/ # Модели базы данных
│ │ ├── user.go # Пользователи
│ │ ├── training.go # Тренировочные программы
│ │ ├── nutrition.go # Планы питания
│ │ ├── category.go # Категории
│ │ ├── weekly_menu.go # Недельные меню
│ │ └── menu_day.go # Дни меню
│ ├── repository/ # Репозитории (Data Access Layer)
│ ├── service/ # Бизнес-логика (Service Layer)
│ ├── database/ # Конфигурация базы данных
│ └── pkg/utils/ # Вспомогательные утилиты
├── docker-compose.yml # Docker Compose конфигурация
├── Dockerfile # Dockerfile для бота
├── .env.example # Пример переменных окружения
├── go.mod & go.sum # Зависимости Go
└── README.md # Документация

## 🚀 Быстрый старт
### Вариант 1: Запуск с Docker Compose (рекомендуется)
1. **Клонировать репозиторий:**
   git clone https://github.com/yourusername/telegram-fitness-bot.git
   cd telegram-fitness-bot
Настроить переменные окружения:
cp .env.example .env
# Отредактировать .env файл:
# - TELEGRAM_TOKEN=ваш_токен_бота
# - ADMIN_IDS=ваш_telegram_id
Запустить сервисы:
docker-compose up -d
Проверить запуск:
docker-compose logs -f bot# Test comment

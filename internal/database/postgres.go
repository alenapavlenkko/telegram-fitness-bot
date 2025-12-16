package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect подключается к PostgreSQL с retry логикой
func NewPostgres(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	log.Printf("Attempting to connect to database...")

	// Пытаемся подключиться 15 раз с увеличением паузы
	for i := 1; i <= 15; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			// Проверяем живое подключение
			sqlDB, _ := db.DB()
			if pingErr := sqlDB.Ping(); pingErr == nil {
				log.Printf("✅ Database connected successfully (attempt %d)", i)
				return db, nil
			}
		}

		log.Printf("Attempt %d failed: %v", i, err)

		// Экспоненциальная backoff: 1, 2, 4, 8 секунд...
		waitTime := time.Duration(1<<uint(i-1)) * time.Second
		if waitTime > 10*time.Second {
			waitTime = 10 * time.Second
		}
		time.Sleep(waitTime)
	}

	return nil, fmt.Errorf("failed to connect to database after 15 attempts: %w", err)
}

// AutoMigrateTables создает таблицы
func AutoMigrateTables(db *gorm.DB, models ...interface{}) error {
	log.Println("Running database migrations...")

	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model: %w", err)
		}
	}

	log.Println("✅ Database migrations completed")
	return nil
}

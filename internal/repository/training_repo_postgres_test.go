package repository

import (
	"os"
	"testing"

	"github.com/alenapavlenkko/telegramfitnes/internal/database"
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL not set")
	}

	db, err := database.NewPostgres(dsn)
	assert.NoError(t, err)

	// Миграция только нужной таблицы
	err = db.AutoMigrate(&models.TrainingProgram{})
	assert.NoError(t, err)

	// Очистка таблицы перед тестом
	db.Exec("DELETE FROM training_programs")

	return db
}

func setupTrainingDB(t *testing.T) *gorm.DB {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL not set")
	}

	db, err := database.NewPostgres(dsn)
	assert.NoError(t, err)

	err = db.AutoMigrate(&models.TrainingProgram{})
	assert.NoError(t, err)

	db.Exec("DELETE FROM training_programs")

	return db
}

func TestTrainingRepo(t *testing.T) {
	db := setupTrainingDB(t)
	repo := NewTrainingRepo(db)

	tp := &models.TrainingProgram{Title: "TestProgram"}
	err := repo.Create(tp)
	assert.NoError(t, err)

	list, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "TestProgram", list[0].Title)

	got, err := repo.GetByID(tp.ID)
	assert.NoError(t, err)
	assert.Equal(t, "TestProgram", got.Title)
}

func TestCreateAndListTraining(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTrainingRepo(db)

	tp := &models.TrainingProgram{Title: "TestProgram"}
	err := repo.Create(tp)
	assert.NoError(t, err)

	list, err := repo.List()
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "TestProgram", list[0].Title)
}

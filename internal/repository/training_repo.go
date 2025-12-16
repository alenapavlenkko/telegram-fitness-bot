package repository

import (
	"log"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

type TrainingRepository interface {
	Create(training *models.TrainingProgram) (*models.TrainingProgram, error)
	FindAll() ([]*models.TrainingProgram, error)
	FindByID(id uint) (*models.TrainingProgram, error)
	Update(training *models.TrainingProgram) error
	Delete(id uint) error
}

type trainingRepo struct {
	db *gorm.DB
}

func NewTrainingRepo(db *gorm.DB) TrainingRepository {
	return &trainingRepo{db: db}
}

func (r *trainingRepo) Create(training *models.TrainingProgram) (*models.TrainingProgram, error) {
	err := r.db.Create(training).Error
	return training, err
}

func (r *trainingRepo) FindAll() ([]*models.TrainingProgram, error) {
	var trainings []*models.TrainingProgram

	// Добавляем SQL-логирование
	r.db = r.db.Debug() // Это покажет SQL-запросы в логах

	err := r.db.Preload("Category").Find(&trainings).Error

	// Логируем результат
	if err != nil {
		log.Printf("ERROR in FindAll: %v", err)
	} else {
		log.Printf("SUCCESS in FindAll: found %d trainings", len(trainings))
		for i, t := range trainings {
			log.Printf("  Training %d: ID=%d, Title='%s', Duration=%d",
				i+1, t.ID, t.Title, t.Duration)
		}
	}

	return trainings, err
}

func (r *trainingRepo) FindByID(id uint) (*models.TrainingProgram, error) {
	var training models.TrainingProgram
	err := r.db.Preload("Category").First(&training, id).Error
	return &training, err
}

func (r *trainingRepo) Update(training *models.TrainingProgram) error {
	return r.db.Save(training).Error
}

func (r *trainingRepo) Delete(id uint) error {
	return r.db.Delete(&models.TrainingProgram{}, id).Error
}

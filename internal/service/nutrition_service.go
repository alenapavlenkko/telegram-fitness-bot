package service

import (
	"fmt"
	"log"

	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"github.com/alenapavlenkko/telegramfitnes/internal/repository"
)

type NutritionService struct {
	repo           repository.NutritionRepository
	weeklyMenuRepo repository.WeeklyMenuRepository
}

func NewNutritionService(repo repository.NutritionRepository, weeklyMenuRepo repository.WeeklyMenuRepository) *NutritionService {
	return &NutritionService{
		repo:           repo,
		weeklyMenuRepo: weeklyMenuRepo,
	}
}

// CreateNutrition - создать план питания
func (s *NutritionService) CreateNutrition(dto CreateNutritionDTO) (*models.NutritionPlan, error) {
	plan := &models.NutritionPlan{
		Title:       dto.Title,
		Description: dto.Description,
		Calories:    dto.Calories,
		Protein:     dto.Protein,
		Carbs:       dto.Carbs,
		Fats:        dto.Fats,
		CategoryID:  dto.CategoryID,
	}
	return s.repo.Create(plan)
}

// List - список планов питания
func (s *NutritionService) ListNutrition() ([]*models.NutritionPlan, error) {
	return s.repo.FindAll()
}

// GetNutritionByID - получить план питания по ID
func (s *NutritionService) GetNutritionByID(id uint) (*models.NutritionPlan, error) {
	return s.repo.FindByID(id)
}

// DeleteNutrition - удалить план питания
func (s *NutritionService) DeleteNutrition(id uint) error {
	return s.repo.Delete(id)
}

// UpdateNutrition - обновить план питания
func (s *NutritionService) UpdateNutrition(id uint, dto UpdateNutritionDTO) error {
	plan, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if dto.Title != "" {
		plan.Title = dto.Title
	}
	if dto.Description != "" {
		plan.Description = dto.Description
	}
	if dto.Calories > 0 {
		plan.Calories = dto.Calories
	}
	if dto.Protein >= 0 {
		plan.Protein = dto.Protein
	}
	if dto.Carbs >= 0 {
		plan.Carbs = dto.Carbs
	}
	if dto.Fats >= 0 {
		plan.Fats = dto.Fats
	}
	if dto.CategoryID > 0 {
		plan.CategoryID = dto.CategoryID
	}

	return s.repo.Update(plan)
}

// ==================== МЕТОДЫ ДЛЯ НЕДЕЛЬНОГО МЕНЮ ====================

// CreateWeeklyMenu - создать недельное меню
func (s *NutritionService) CreateWeeklyMenu(dto CreateWeeklyMenuDTO) (*models.WeeklyMenu, error) {
	if dto.Name == "" {
		return nil, fmt.Errorf("название меню не может быть пустым")
	}

	menu := &models.WeeklyMenu{
		Name:          dto.Name,
		Description:   dto.Description,
		TotalCalories: 0,
		Active:        false,
	}

	return s.weeklyMenuRepo.Create(menu)
}

// ListWeeklyMenus - список всех недельных меню
func (s *NutritionService) ListWeeklyMenus() ([]*models.WeeklyMenu, error) {
	return s.weeklyMenuRepo.FindAll()
}

// GetActiveWeeklyMenu - получить активное недельное меню
func (s *NutritionService) GetActiveWeeklyMenu() (*models.WeeklyMenu, error) {
	return s.weeklyMenuRepo.FindActive()
}

// ActivateWeeklyMenu - активировать недельное меню
func (s *NutritionService) ActivateWeeklyMenu(menuID uint) error {
	// Деактивируем все меню
	if err := s.weeklyMenuRepo.DeactivateAll(); err != nil {
		return err
	}

	// Активируем выбранное
	return s.weeklyMenuRepo.Activate(menuID)
}

// AddDayToWeeklyMenu - добавить день в недельное меню
func (s *NutritionService) AddDayToWeeklyMenu(dto AddDayToMenuDTO) (*models.MenuDay, error) {
	if dto.DayNumber < 1 || dto.DayNumber > 7 {
		return nil, fmt.Errorf("номер дня должен быть от 1 до 7")
	}

	day := &models.MenuDay{
		MenuID:        dto.MenuID,
		DayNumber:     dto.DayNumber,
		DayName:       dto.DayName,
		TotalCalories: 0,
	}

	return s.weeklyMenuRepo.CreateDay(day)
}

func (s *NutritionService) AddMealToDay(dto AddMealToDayDTO) (*models.DayMeal, error) {
	// Проверяем существование питания
	_, err := s.repo.FindByID(dto.NutritionID)
	if err != nil {
		return nil, fmt.Errorf("питание не найдено: %w", err)
	}

	meal := &models.DayMeal{
		DayID:       dto.DayID,
		MealType:    dto.MealType,
		MealTime:    dto.MealTime,
		NutritionID: dto.NutritionID,
		Notes:       dto.Notes,
	}

	// Создаем прием пищи
	createdMeal, err := s.weeklyMenuRepo.CreateMeal(meal)
	if err != nil {
		return nil, err
	}

	// Обновляем калории дня
	if err := s.updateDayCalories(dto.DayID); err != nil {
		log.Printf("Warning: failed to update day calories: %v", err)
	}

	// Обновляем калории недели
	day, err := s.weeklyMenuRepo.FindDayByID(dto.DayID)
	if err == nil {
		if err := s.updateMenuCalories(day.MenuID); err != nil {
			log.Printf("Warning: failed to update menu calories: %v", err)
		}
	}

	return createdMeal, nil
}

// GetFullWeeklyMenu - получить полное меню с днями и приемами пищи
func (s *NutritionService) GetFullWeeklyMenu(menuID uint) (*models.WeeklyMenu, error) {
	menu, err := s.weeklyMenuRepo.FindByID(menuID)
	if err != nil {
		return nil, err
	}

	// Получаем дни
	days, err := s.weeklyMenuRepo.FindDaysByMenuID(menuID)
	if err != nil {
		return nil, err
	}

	var daysSlice []models.MenuDay
	for _, dayPtr := range days {
		if dayPtr != nil {
			daysSlice = append(daysSlice, *dayPtr)
		}
	}
	menu.Days = daysSlice

	// Для каждого дня загружаем приемы пищи с информацией о питании
	for i := range menu.Days {
		meals, err := s.weeklyMenuRepo.FindMealsByDayID(menu.Days[i].ID)
		if err != nil {
			continue
		}

		var mealsSlice []models.DayMeal
		for _, mealPtr := range meals {
			if mealPtr != nil {
				meal := *mealPtr
				// Загружаем информацию о питании
				nutrition, err := s.repo.FindByID(meal.NutritionID)
				if err == nil {
					meal.Nutrition = *nutrition
				}
				mealsSlice = append(mealsSlice, meal)
			}
		}
		menu.Days[i].Meals = mealsSlice
	}

	return menu, nil
}

// updateDayCalories - обновить калории дня
func (s *NutritionService) updateDayCalories(dayID uint) error {
	meals, err := s.weeklyMenuRepo.FindMealsByDayID(dayID)
	if err != nil {
		return err
	}

	totalCalories := 0
	for _, meal := range meals {
		nutrition, err := s.repo.FindByID(meal.NutritionID)
		if err == nil {
			totalCalories += nutrition.Calories
		}
	}

	return s.weeklyMenuRepo.UpdateDayCalories(dayID, totalCalories)
}

// updateMenuCalories - обновить калории меню
func (s *NutritionService) updateMenuCalories(menuID uint) error {
	days, err := s.weeklyMenuRepo.FindDaysByMenuID(menuID)
	if err != nil {
		return err
	}

	totalCalories := 0
	for _, day := range days {
		totalCalories += day.TotalCalories
	}

	return s.weeklyMenuRepo.UpdateMenuCalories(menuID, totalCalories)
}

// DeleteWeeklyMenu - удалить недельное меню
func (s *NutritionService) DeleteWeeklyMenu(id uint) error {
	return s.weeklyMenuRepo.Delete(id)
}

// DeleteDayFromMenu - удалить день из меню
func (s *NutritionService) DeleteDayFromMenu(dayID uint) error {
	return s.weeklyMenuRepo.DeleteDay(dayID)
}

// DeleteMealFromDay - удалить прием пищи из дня
func (s *NutritionService) DeleteMealFromDay(mealID uint) error {
	// Получаем информацию о приеме пищи
	meal, err := s.weeklyMenuRepo.FindMealByID(mealID)
	if err != nil {
		return err
	}

	// Удаляем прием пищи
	if err := s.weeklyMenuRepo.DeleteMeal(mealID); err != nil {
		return err
	}

	// Обновляем калории дня
	if err := s.updateDayCalories(meal.DayID); err != nil {
		log.Printf("Warning: failed to update day calories: %v", err)
	}

	// Обновляем калории недели
	day, err := s.weeklyMenuRepo.FindDayByID(meal.DayID)
	if err == nil {
		if err := s.updateMenuCalories(day.MenuID); err != nil {
			log.Printf("Warning: failed to update menu calories: %v", err)
		}
	}

	return nil
}

// internal/repository/nutrition_repo.go
package repository

import (
	"github.com/alenapavlenkko/telegramfitnes/internal/models"
	"gorm.io/gorm"
)

// NutritionRepository - интерфейс для планов питания
type NutritionRepository interface {
	Create(plan *models.NutritionPlan) (*models.NutritionPlan, error)
	FindAll() ([]*models.NutritionPlan, error)
	FindByID(id uint) (*models.NutritionPlan, error)
	Update(plan *models.NutritionPlan) error
	Delete(id uint) error
}

// WeeklyMenuRepository - интерфейс для недельных меню
type WeeklyMenuRepository interface {
	// Меню
	Create(menu *models.WeeklyMenu) (*models.WeeklyMenu, error)
	FindAll() ([]*models.WeeklyMenu, error)
	FindByID(id uint) (*models.WeeklyMenu, error)
	FindActive() (*models.WeeklyMenu, error)
	Update(menu *models.WeeklyMenu) error
	Delete(id uint) error
	Activate(id uint) error
	DeactivateAll() error
	UpdateMenuCalories(menuID uint, calories int) error

	// Дни
	CreateDay(day *models.MenuDay) (*models.MenuDay, error)
	FindDaysByMenuID(menuID uint) ([]*models.MenuDay, error)
	FindDayByID(dayID uint) (*models.MenuDay, error)
	UpdateDay(day *models.MenuDay) error
	DeleteDay(dayID uint) error
	UpdateDayCalories(dayID uint, calories int) error

	// Приемы пищи
	CreateMeal(meal *models.DayMeal) (*models.DayMeal, error)
	FindMealsByDayID(dayID uint) ([]*models.DayMeal, error)
	FindMealByID(mealID uint) (*models.DayMeal, error)
	UpdateMeal(meal *models.DayMeal) error
	DeleteMeal(mealID uint) error
}

type nutritionRepo struct {
	db *gorm.DB
}

func NewNutritionRepo(db *gorm.DB) *nutritionRepo {
	return &nutritionRepo{db: db}
}

// Реализация NutritionRepository

func (r *nutritionRepo) Create(plan *models.NutritionPlan) (*models.NutritionPlan, error) {
	result := r.db.Create(plan)
	return plan, result.Error
}

func (r *nutritionRepo) FindAll() ([]*models.NutritionPlan, error) {
	var plans []*models.NutritionPlan
	result := r.db.Preload("Category").Find(&plans)
	return plans, result.Error
}

func (r *nutritionRepo) FindByID(id uint) (*models.NutritionPlan, error) {
	var plan models.NutritionPlan
	result := r.db.Preload("Category").First(&plan, id)
	return &plan, result.Error
}

func (r *nutritionRepo) Update(plan *models.NutritionPlan) error {
	result := r.db.Save(plan)
	return result.Error
}

func (r *nutritionRepo) Delete(id uint) error {
	result := r.db.Delete(&models.NutritionPlan{}, id)
	return result.Error
}

// ==================== РЕАЛИЗАЦИЯ WeeklyMenuRepository ====================

type weeklyMenuRepo struct {
	db *gorm.DB
}

func NewWeeklyMenuRepo(db *gorm.DB) *weeklyMenuRepo {
	return &weeklyMenuRepo{db: db}
}

// Меню

func (r *weeklyMenuRepo) Create(menu *models.WeeklyMenu) (*models.WeeklyMenu, error) {
	result := r.db.Create(menu)
	return menu, result.Error
}

func (r *weeklyMenuRepo) FindAll() ([]*models.WeeklyMenu, error) {
	var menus []*models.WeeklyMenu
	result := r.db.Find(&menus)
	return menus, result.Error
}

func (r *weeklyMenuRepo) FindByID(id uint) (*models.WeeklyMenu, error) {
	var menu models.WeeklyMenu
	result := r.db.First(&menu, id)
	return &menu, result.Error
}

func (r *weeklyMenuRepo) FindActive() (*models.WeeklyMenu, error) {
	var menu models.WeeklyMenu
	result := r.db.Where("active = ?", true).First(&menu)
	return &menu, result.Error
}

func (r *weeklyMenuRepo) Update(menu *models.WeeklyMenu) error {
	result := r.db.Save(menu)
	return result.Error
}

func (r *weeklyMenuRepo) Delete(id uint) error {
	// Удаляем меню вместе с днями и приемами пищи (каскадное удаление)
	result := r.db.Delete(&models.WeeklyMenu{}, id)
	return result.Error
}

func (r *weeklyMenuRepo) Activate(id uint) error {
	result := r.db.Model(&models.WeeklyMenu{}).Where("id = ?", id).Update("active", true)
	return result.Error
}

func (r *weeklyMenuRepo) DeactivateAll() error {
	result := r.db.Model(&models.WeeklyMenu{}).Where("active = ?", true).Update("active", false)
	return result.Error
}

func (r *weeklyMenuRepo) UpdateMenuCalories(menuID uint, calories int) error {
	result := r.db.Model(&models.WeeklyMenu{}).Where("id = ?", menuID).Update("total_calories", calories)
	return result.Error
}

// Дни

func (r *weeklyMenuRepo) CreateDay(day *models.MenuDay) (*models.MenuDay, error) {
	result := r.db.Create(day)
	return day, result.Error
}

func (r *weeklyMenuRepo) FindDaysByMenuID(menuID uint) ([]*models.MenuDay, error) {
	var days []*models.MenuDay
	result := r.db.Where("menu_id = ?", menuID).Find(&days)
	return days, result.Error
}

func (r *weeklyMenuRepo) FindDayByID(dayID uint) (*models.MenuDay, error) {
	var day models.MenuDay
	result := r.db.First(&day, dayID)
	return &day, result.Error
}

func (r *weeklyMenuRepo) UpdateDay(day *models.MenuDay) error {
	result := r.db.Save(day)
	return result.Error
}

func (r *weeklyMenuRepo) DeleteDay(dayID uint) error {
	result := r.db.Delete(&models.MenuDay{}, dayID)
	return result.Error
}

func (r *weeklyMenuRepo) UpdateDayCalories(dayID uint, calories int) error {
	result := r.db.Model(&models.MenuDay{}).Where("id = ?", dayID).Update("total_calories", calories)
	return result.Error
}

// Приемы пищи

func (r *weeklyMenuRepo) CreateMeal(meal *models.DayMeal) (*models.DayMeal, error) {
	result := r.db.Create(meal)
	return meal, result.Error
}

func (r *weeklyMenuRepo) FindMealsByDayID(dayID uint) ([]*models.DayMeal, error) {
	var meals []*models.DayMeal
	result := r.db.Where("day_id = ?", dayID).Find(&meals)
	return meals, result.Error
}

func (r *weeklyMenuRepo) FindMealByID(mealID uint) (*models.DayMeal, error) {
	var meal models.DayMeal
	result := r.db.First(&meal, mealID)
	return &meal, result.Error
}

func (r *weeklyMenuRepo) UpdateMeal(meal *models.DayMeal) error {
	result := r.db.Save(meal)
	return result.Error
}

func (r *weeklyMenuRepo) DeleteMeal(mealID uint) error {
	result := r.db.Delete(&models.DayMeal{}, mealID)
	return result.Error
}

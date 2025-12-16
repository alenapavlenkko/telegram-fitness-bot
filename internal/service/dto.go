package service

// Training DTOs
type CreateTrainingDTO struct {
	Title       string
	Duration    int
	Difficulty  string
	Description string
	CategoryID  *uint
	YouTubeLink string
}

type UpdateTrainingDTO struct {
	Title       string
	Duration    int
	Difficulty  string
	Description string
	CategoryID  *uint
	YouTubeLink string
}

// DTO для недельного меню
type CreateWeeklyMenuDTO struct {
	Name        string
	Description string
}

type AddDayToMenuDTO struct {
	MenuID    uint
	DayNumber int
	DayName   string
}

type AddMealToDayDTO struct {
	DayID       uint
	MealType    string
	MealTime    string
	NutritionID uint
	Notes       string
}

// Остальные существующие DTO...
type CreateNutritionDTO struct {
	Title       string
	Description string
	Calories    int
	Protein     float64
	Carbs       float64
	Fats        float64
	CategoryID  uint
}

type UpdateNutritionDTO struct {
	Title       string
	Description string
	Calories    int
	Protein     float64
	Carbs       float64
	Fats        float64
	CategoryID  uint
}

// Category DTOs
type CreateCategoryDTO struct {
	Name        string
	Description string
	Type        string
}
type UpdateCategoryDTO struct {
	Name        string
	Description string
	Type        string
}

// User DTOs
type CreateUserDTO struct {
	TelegramID int64
	Name       string
	Role       string
}

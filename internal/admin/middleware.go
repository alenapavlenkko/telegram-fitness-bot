package admin

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthMiddleware - базовый пример проверки админ-доступа
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// !!! СЮДА НУЖНО ВСТАВИТЬ ЛОГИКУ АВТОРИЗАЦИИ !!!
		// Например, проверка JWT токена или сессии.

		// Простой пример (небезопасно для продакшена):
		// Проверяем наличие заголовка "X-Admin-Key"
		adminKey := c.GetHeader("X-Admin-Key")
		if adminKey != "your-secret-admin-key" { // Используйте переменную окружения!
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort() // Останавливаем выполнение дальнейших обработчиков
			return
		}

		c.Next() // Продолжаем выполнение запроса
	}
}

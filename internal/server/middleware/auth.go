package middleware

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Auth(repo *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		header := c.Request.Header.Get("Authorization")
		token := strings.Split(header, " ")[1]

		user, err := repo.GetUserByToken(ctx, &token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.ResponseError{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}

		check := user.SessionExpiry.Compare(time.Now())
		if check < 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.ResponseError{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}

		c.Set("username", user.Name)
		c.Next()
	}
}

package middleware

import (
	"Shoka/internal/models"
	"Shoka/internal/repository"
	"Shoka/internal/util"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Auth(repo *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.Background()
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "No authorization header set.",
			})
			return
		}
		token := util.GetAuthTokenFromHeader(header)

		user, err := repo.GetUserByToken(ctx, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}

		check := user.ExpiresAt.Compare(time.Now())
		if check < 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "Unauthorized",
			})
			return
		}

		c.Set("username", user.Name)
		c.Set("userid", user.ID.String())
		c.Set("token", token)
		c.Next()
	}
}

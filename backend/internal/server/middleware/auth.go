// Package middleware contains all the middleware functions for the gin server
package middleware

import (
	"net/http"
	"time"

	"Shoka/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

// TODO: Use env variable's as url for frontend
// TODO: Add JWK caching: https://github.com/lestrrat-go/jwx/blob/develop/v3/docs/04-jwk.md#auto-refreshing-remote-keys

// Auth authenticates a user with the api server
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.Request.Header.Get("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, &models.Response{
				Status:  "error",
				Message: "missing authorization header",
			})
			return
		}

		keyset, err := jwk.Fetch(c.Request.Context(), "http://localhost:3000/api/auth/jwks")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: "missing JWKS",
			})
			return
		}

		token, err := jwt.ParseRequest(c.Request, jwt.WithKeySet(keyset))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, &models.Response{
				Status:  "error",
				Message: "invalid or incorrect jwt",
			})
			return
		}

		currentTime := time.Now()
		expTime, exists := token.Expiration()
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "missing expiration time",
			})
			return
		}

		if currentTime.After(expTime) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "token expired",
			})
			return
		}

		userID, exists := token.Subject()
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, &models.Response{
				Status:  "error",
				Message: "missing user id",
			})
			return
		}

		var username string
		token.Get("username", &username)

		c.Set("username", username)
		c.Set("userid", userID)
		c.Next()
	}
}

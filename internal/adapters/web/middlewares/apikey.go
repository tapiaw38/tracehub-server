package middlewares

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tapiaw38/tracehub-server/internal/domain"
	"github.com/tapiaw38/tracehub-server/internal/platform/auth"
)

type ProjectRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Project, error)
}

// ApiKeyAuth is a middleware that validates API keys
func ApiKeyAuth(projectRepo ProjectRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		apiKey, err := auth.ExtractApiKey(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		if err := auth.ValidateApiKeyFormat(apiKey); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key format"})
			c.Abort()
			return
		}

		// For now, extract project ID from query param
		// In production, you'd validate against api_keys table
		projectID := c.Query("project_id")
		if projectID == "" {
			projectID = c.PostForm("project_id")
		}
		if projectID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "project_id required"})
			c.Abort()
			return
		}

		// Set project ID in context
		c.Set("project_id", projectID)
		c.Next()
	}
}

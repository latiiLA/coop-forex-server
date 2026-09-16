package middleware

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
)

// ServeProtectedUpload serves files from the upload directory for authenticated users only.
// Rejects path traversal attempts outside the configured upload root.
func ServeProtectedUpload(c *gin.Context) {
	relPath := c.Param("filepath")
	relPath = strings.TrimPrefix(relPath, "/")
	if relPath == "" || strings.Contains(relPath, "..") {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Invalid file path", Error: "invalid path"})
		return
	}

	uploadRoot, err := filepath.Abs(configs.FileUploadPath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, response.Status{Message: "Internal server error", Error: "upload path error"})
		return
	}

	absPath, err := filepath.Abs(filepath.Join(uploadRoot, relPath))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Invalid file path", Error: "invalid path"})
		return
	}

	if !strings.HasPrefix(absPath, uploadRoot+string(filepath.Separator)) && absPath != uploadRoot {
		c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Access denied", Error: "path outside upload root"})
		return
	}

	c.File(absPath)
}

package middleware

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/delivery/http/response"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
)

func hasPermissionList(perms []string, target string) bool {
	for _, p := range perms {
		if p == target {
			return true
		}
	}
	return false
}

func sameOrgUser(user *model.User, request *model.Request) bool {
	if user == nil || request == nil || user.Profile == nil {
		return false
	}
	if request.CreatedBy == user.ID {
		return true
	}
	sameBranch := user.Profile.BranchID != nil &&
		request.BranchID != nil &&
		*user.Profile.BranchID == *request.BranchID
	sameDepartment := user.Profile.DepartmentID != nil &&
		request.DepartmentID != nil &&
		*user.Profile.DepartmentID == *request.DepartmentID
	return sameBranch || sameDepartment
}

// NewServeProtectedUpload serves upload files only when the caller may see the parent request.
func NewServeProtectedUpload(
	fileRepo model.FileRepository,
	requestRepo model.RequestRepository,
	userRepo model.UserRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		relPath := c.Param("filepath")
		relPath = strings.TrimPrefix(relPath, "/")
		if relPath == "" || strings.Contains(relPath, "..") {
			c.AbortWithStatusJSON(http.StatusBadRequest, response.Status{Message: "Invalid file path", Error: "invalid path"})
			return
		}

		fileName := filepath.Base(relPath)
		fileMeta, err := fileRepo.FindByName(c.Request.Context(), fileName)
		if err != nil || fileMeta == nil {
			c.AbortWithStatusJSON(http.StatusNotFound, response.Status{Message: "File not found", Error: common.ErrFileNotFound.Error()})
			return
		}

		userID, err := utils.GetUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: err.Error()})
			return
		}

		user, err := userRepo.FindByID(c.Request.Context(), userID)
		if err != nil || user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, response.Status{Message: common.MessUnauthorized, Error: "user not found"})
			return
		}

		parent, err := requestRepo.FindByAttachmentFileID(c.Request.Context(), fileMeta.ID)
		if err != nil || parent == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Access denied", Error: common.ErrFileAccessDenied.Error()})
			return
		}

		perms := utils.GetPermissions(c)
		allowed := false
		if hasPermissionList(perms, "request:view") {
			allowed = true
		} else if sameOrgUser(user, parent) {
			allowed = true
		} else if parent.CreatedBy == userID {
			allowed = true
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, response.Status{Message: "Access denied", Error: common.ErrFileAccessDenied.Error()})
			return
		}

		uploadRoot, err := filepath.Abs(configs.FileUploadPath)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, response.Status{Message: "Internal server error", Error: "upload path error"})
			return
		}

		absPath, err := filepath.Abs(filepath.Join(uploadRoot, fileName))
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
}

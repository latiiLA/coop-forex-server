package usecase

import (
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const maxUploadBytes int64 = 5 << 20 // 5MB

var allowedUploadExt = map[string]bool{
	".pdf":  true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

var allowedUploadMime = map[string]bool{
	"application/pdf": true,
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"application/octet-stream": true, // browsers sometimes send this; ext still enforced
}

type FileUsecase interface {
	AddFile(ctx context.Context, file *multipart.FileHeader, prefix string) (*primitive.ObjectID, error)
	GetFileByID(ctx context.Context, file_id primitive.ObjectID) (*model.File, error)
	GetFileByName(ctx context.Context, name string) (*model.File, error)
}

type fileUsecase struct {
	fileRepository model.FileRepository
	contextTimeout time.Duration
}

func NewFileUsecase(fileRepository model.FileRepository, timeout time.Duration) FileUsecase {
	return &fileUsecase{
		fileRepository: fileRepository,
		contextTimeout: timeout,
	}
}

func (fu *fileUsecase) AddFile(ctx context.Context, file *multipart.FileHeader, prefix string) (*primitive.ObjectID, error) {
	if file == nil {
		return nil, common.ErrInvalidUpload
	}
	if file.Size <= 0 || file.Size > maxUploadBytes {
		return nil, common.ErrInvalidUpload
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadExt[ext] {
		return nil, common.ErrInvalidUpload
	}

	mimeType := strings.ToLower(strings.TrimSpace(file.Header.Get("Content-Type")))
	if mimeType != "" && !allowedUploadMime[mimeType] {
		return nil, common.ErrInvalidUpload
	}

	originalName := utils.SanitizeFilename(strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)))
	fid := uuid.NewString()
	uniqueFilename := prefix + "_" + originalName + "_" + fid + ext
	uploadPath := configs.FileUploadPath

	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		_ = os.MkdirAll(uploadPath, os.ModePerm)
	}

	fullPath := filepath.Join(uploadPath, uniqueFilename)

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	written, err := io.Copy(dst, io.LimitReader(src, maxUploadBytes+1))
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}
	if written > maxUploadBytes {
		_ = os.Remove(fullPath)
		return nil, common.ErrInvalidUpload
	}

	if mimeType == "" || mimeType == "application/octet-stream" {
		switch ext {
		case ".pdf":
			mimeType = "application/pdf"
		case ".png":
			mimeType = "image/png"
		default:
			mimeType = "image/jpeg"
		}
	}

	newFile := model.File{
		ID:        primitive.NewObjectID(),
		Name:      uniqueFilename,
		URL:       fullPath,
		Fid:       fid,
		Size:      written,
		MimeType:  mimeType,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fileID, err := fu.fileRepository.Create(ctx, &newFile)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}

	return fileID, nil
}

func (fu *fileUsecase) GetFileByID(ctx context.Context, file_id primitive.ObjectID) (*model.File, error) {
	ctx, cancel := context.WithTimeout(ctx, fu.contextTimeout)
	defer cancel()
	return fu.fileRepository.FindByID(ctx, file_id)
}

func (fu *fileUsecase) GetFileByName(ctx context.Context, name string) (*model.File, error) {
	ctx, cancel := context.WithTimeout(ctx, fu.contextTimeout)
	defer cancel()
	return fu.fileRepository.FindByName(ctx, name)
}

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
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FileUsecase interface {
	AddFile(ctx context.Context, file *multipart.FileHeader, prefix string) (*primitive.ObjectID, error)
	GetFileByID(ctx context.Context, file_id primitive.ObjectID) (*model.File, error)
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
	ext := filepath.Ext(file.Filename)
	originalName := utils.SanitizeFilename(strings.TrimSuffix(file.Filename, ext))

	fid := uuid.NewString()
	uniqueFilename := prefix + "_" + originalName + "_" + fid + ext
	uploadPath := configs.FileUploadPath

	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		os.MkdirAll(uploadPath, os.ModePerm)
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

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	newFile := model.File{
		ID:        primitive.NewObjectID(),
		Name:      uniqueFilename,
		URL:       fullPath,
		Fid:       fid,
		Size:      file.Size,
		MimeType:  file.Header.Get("Content-Type"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	fileID, err := fu.fileRepository.Create(ctx, &newFile)
	if err != nil {
		return nil, err
	}

	return fileID, nil
}

func (ru *fileUsecase) GetFileByID(ctx context.Context, file_id primitive.ObjectID) (*model.File, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.fileRepository.FindByID(ctx, file_id)
}

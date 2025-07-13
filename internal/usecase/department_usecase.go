package usecase

import (
	"context"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DepartmentUsecase interface {
	AddDepartment(ctx context.Context, userID primitive.ObjectID, department *model.Department) error
	GetDepartmentByID(ctx context.Context, department_id primitive.ObjectID) (*model.Department, error)
	GetDepartmentBySubprocessID(ctx context.Context, subprocess_id primitive.ObjectID) (*[]model.Department, error)
	GetAllDepartments(ctx context.Context) ([]model.Department, error)
}

type departmentUsecase struct {
	departmentRepository model.DepartmentRepository
	contextTimeout       time.Duration
}

func NewDepartmentUsecase(departmentRepository model.DepartmentRepository, timeout time.Duration) DepartmentUsecase {
	return &departmentUsecase{
		departmentRepository: departmentRepository,
		contextTimeout:       timeout,
	}
}

func (du *departmentUsecase) AddDepartment(ctx context.Context, userID primitive.ObjectID, department *model.Department) error {
	return nil
}

func (du *departmentUsecase) GetDepartmentByID(ctx context.Context, department_id primitive.ObjectID) (*model.Department, error) {
	return nil, nil
}

func (du *departmentUsecase) GetAllDepartments(ctx context.Context) ([]model.Department, error) {
	ctx, cancel := context.WithTimeout(ctx, du.contextTimeout)
	defer cancel()
	return du.departmentRepository.FindAll(ctx)
}

func (su *departmentUsecase) GetDepartmentBySubprocessID(ctx context.Context, subprocess_id primitive.ObjectID) (*[]model.Department, error) {
	ctx, cancel := context.WithTimeout(ctx, su.contextTimeout)
	defer cancel()
	return su.departmentRepository.FindBySubprocessID(ctx, subprocess_id)
}

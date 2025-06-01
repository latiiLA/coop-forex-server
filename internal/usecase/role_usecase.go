package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleUsecase interface {
	AddRole(ctx context.Context, userID primitive.ObjectID, role *model.Role) error
	GetRoleByID(ctx context.Context, role_id primitive.ObjectID) (model.Role, error)
	GetAllRoles(ctx context.Context) ([]model.Role, error)
}

type roleUsecase struct{
	roleRepository model.RoleRepository
	contextTimeout time.Duration
}

func NewRoleUsecase(roleRepository model.RoleRepository, timeout time.Duration) RoleUsecase{
	return &roleUsecase{
		roleRepository: roleRepository,
		contextTimeout: timeout,
	}
}

func (ru *roleUsecase) AddRole(ctx context.Context, userID primitive.ObjectID, role *model.Role) error{
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	if role.Type == "superadmin"{
		return errors.New("role not allowed")
	}

	exists, err := ru.roleRepository.ExistsRoleByName(ctx, role.Type)
	if err != nil{
		return err
	}
	if exists{
		return errors.New("role already exists")
	}

	role.ID = primitive.NewObjectID()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	role.CreatedBy = userID
	role.IsDeleted = false

	return ru.roleRepository.Create(ctx, role)
}

func (ru *roleUsecase) GetRoleByID(ctx context.Context, role_id primitive.ObjectID) (model.Role, error){
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.roleRepository.FindByID(ctx, role_id)
}

func (ru *roleUsecase) GetAllRoles(ctx context.Context) ([]model.Role, error){
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.roleRepository.FindAll(ctx)
}
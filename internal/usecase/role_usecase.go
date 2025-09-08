package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RoleUsecase interface {
	AddRole(ctx context.Context, authUserID primitive.ObjectID, role *model.Role) error
	GetRoleByID(ctx context.Context, roleID primitive.ObjectID) (*model.Role, error)
	GetAllRoles(ctx context.Context) ([]model.Role, error)
	UpdateRole(ctx context.Context, authUserID primitive.ObjectID, roleID primitive.ObjectID, role model.UpdateRoleDTO) error
	DeleteRole(ctx context.Context, authUserID primitive.ObjectID, roleID primitive.ObjectID) error
	GetDeletedRoles(ctx context.Context) ([]model.Role, error)
}

type roleUsecase struct {
	roleRepository model.RoleRepository
	contextTimeout time.Duration
}

func NewRoleUsecase(roleRepository model.RoleRepository, timeout time.Duration) RoleUsecase {
	return &roleUsecase{
		roleRepository: roleRepository,
		contextTimeout: timeout,
	}
}

func (ru *roleUsecase) AddRole(ctx context.Context, authUserID primitive.ObjectID, role *model.Role) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	if role.Name == "SUPERADMIN" {
		return errors.New("role not allowed")
	}

	existingRoleName, err := ru.roleRepository.FindRoleByName(ctx, strings.ToUpper(role.Name))
	if err != nil {
		return err
	}
	if existingRoleName != nil {
		return errors.New("role already exists")
	}

	role.ID = primitive.NewObjectID()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	role.CreatedBy = authUserID
	role.IsDeleted = false

	return ru.roleRepository.Create(ctx, role)
}

func (ru *roleUsecase) GetRoleByID(ctx context.Context, role_id primitive.ObjectID) (*model.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.roleRepository.FindByID(ctx, role_id)
}

func (ru *roleUsecase) GetAllRoles(ctx context.Context) ([]model.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.roleRepository.FindAll(ctx)
}

func (ru *roleUsecase) UpdateRole(ctx context.Context, authUserID primitive.ObjectID, roleID primitive.ObjectID, roleUpdated model.UpdateRoleDTO) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	role, err := ru.roleRepository.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to find role by ID: %w", err)
	}

	if role == nil {
		return errors.New("role not found")
	}

	if roleUpdated.Name != "" {
		existingRoleName, err := ru.roleRepository.FindRoleByName(ctx, strings.ToUpper(roleUpdated.Name))
		if err != nil {
			return nil
		}
		if existingRoleName != nil && role.Name != existingRoleName.Name {
			return errors.New("role with this name already exists")
		}

		role.Name = strings.ToUpper(roleUpdated.Name)
	}

	now := time.Now().UTC()
	if role.Permissions != nil {
		role.Permissions = roleUpdated.Permissions
	}
	if role.Name != "" {
		role.Name = roleUpdated.Name
	}

	role.UpdatedAt = now
	role.UpdatedBy = &authUserID

	return ru.roleRepository.Update(ctx, roleID, role)
}

func (ru *roleUsecase) DeleteRole(ctx context.Context, authUserID primitive.ObjectID, roleID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()

	role, err := ru.roleRepository.FindByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to find role by ID: %w", err)
	}

	now := time.Now().UTC()
	role.DeletedAt = &now
	role.DeletedBy = &authUserID
	role.IsDeleted = true

	return ru.roleRepository.Delete(ctx, roleID, role)
}

func (ru *roleUsecase) GetDeletedRoles(ctx context.Context) ([]model.Role, error) {
	ctx, cancel := context.WithTimeout(ctx, ru.contextTimeout)
	defer cancel()
	return ru.roleRepository.FindDeleted(ctx)
}

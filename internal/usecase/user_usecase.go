package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(c context.Context, user *model.User) error
	Login(c context.Context, userReq model.LoginRequestDTO) (*model.LoginResponseDTO, error)
	GetUserByID(c context.Context, userID primitive.ObjectID) (*model.UserResponseDTO, error)
}

type userUsecase struct {
	userRepository    model.UserRepository
	roleRepository    model.RoleRepository
	profileRepository model.ProfileRepository
	contextTimeout    time.Duration
}

func NewUserUsecase(userRepository model.UserRepository, roleRepository model.RoleRepository, profileRepository model.ProfileRepository, timeout time.Duration) UserUsecase {
	return &userUsecase{
		userRepository:    userRepository,
		roleRepository:    roleRepository,
		profileRepository: profileRepository,
		contextTimeout:    timeout,
	}
}

func (uc *userUsecase) Register(c context.Context, user *model.User) error {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	existingUser, err := uc.userRepository.FindByUsername(ctx, user.Username)
	if err == nil && existingUser.Username != "" {
		return errors.New("username already exists")
	}

	encriptedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(encriptedPassword)
	user.ID = primitive.NewObjectID()

	return uc.userRepository.Create(ctx, user)
}

func (uc *userUsecase) Login(c context.Context, userReq model.LoginRequestDTO) (*model.LoginResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	existingUser, err := uc.userRepository.FindByUsername(ctx, userReq.Username)
	if err != nil {
		fmt.Println("Invalid username or user")
		return nil, errors.New("invalid username or password")
	}

	err = infrastructure.CheckPasswordHash(existingUser.Password, userReq.Password)
	if err != nil {
		fmt.Println("Invalid password:", err)
		return nil, errors.New("invalid username or password")
	}

	role, err := uc.roleRepository.FindByID(ctx, existingUser.RoleID)
	if err != nil {
		return nil, errors.New("role not found")
	}

	accessToken, err := infrastructure.GenerateToken(existingUser.ID, role.Type)
	if err != nil {
		return nil, err
	}

	profile, err := uc.profileRepository.FindByID(ctx, existingUser.ProfileID)
	if err != nil {
		return nil, err
	}

	response := model.LoginResponseDTO{
		ID:         existingUser.ID,
		FirstName:  profile.FirstName,
		MiddleName: profile.FirstName,
		Username:   existingUser.Username,
		Role:       role.Type,
		Token:      accessToken,
	}

	return &response, nil
}

func (uc *userUsecase) GetUserByID(ctx context.Context, user_id primitive.ObjectID) (*model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindByID(ctx, user_id)
}

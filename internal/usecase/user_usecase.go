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
	Login(c context.Context, userReq model.LoginRequestDTO) (model.LoginResponseDTO, string, error)
}

type userUsecase struct {
	userRepository model.UserRepository
	roleRepository model.RoleRepository
	contextTimeout time.Duration
}

func NewUserUsecase(userRepository model.UserRepository, roleRepository model.RoleRepository, timeout time.Duration) UserUsecase{
	return &userUsecase{
		userRepository: userRepository,
		roleRepository: roleRepository,
		contextTimeout: timeout,
	}
}

func (uc *userUsecase) Register(c context.Context, user *model.User) error{
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	existingUser, err := uc.userRepository.FindByUsername(ctx, user.Username)
	if err == nil && existingUser.Username != ""{
		return errors.New("username already exists")
	}

	encriptedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil{
		return err
	}

	user.Password = string(encriptedPassword)
	user.ID = primitive.NewObjectID()

	return uc.userRepository.Create(ctx, user)
}

func (uc *userUsecase) Login(c context.Context, userReq model.LoginRequestDTO)(model.LoginResponseDTO, string, error){
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	existingUser, err := uc.userRepository.FindByUsername(ctx, userReq.Username)
	if err != nil{
		fmt.Println("Invalid username or user")
		return model.LoginResponseDTO{}, "", errors.New("invalid username or password")
	}

	err = infrastructure.CheckPasswordHash(existingUser.Password, userReq.Password)
	if err != nil {
		fmt.Println("Invalid password:", err)
		return model.LoginResponseDTO{}, "", errors.New("invalid username or password")
	}

	role, err := uc.roleRepository.FindByID(ctx, existingUser.RoleID)
	if err != nil{
		return model.LoginResponseDTO{}, "", errors.New("role not found")
	}

	accessToken, err := infrastructure.GenerateToken(existingUser.ID, role.Type)
	if err != nil{
		return model.LoginResponseDTO{}, "", err
	}

	response := model.LoginResponseDTO{
		ID: existingUser.ID,
		// FirstName: existingUser.FirstName,
		// MiddleName: existingUser.MiddleName,
		Username: existingUser.Username,
		Role: role.Type,
		Token: accessToken,
	}
	

	return response, accessToken, nil
}

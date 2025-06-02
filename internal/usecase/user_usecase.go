package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(c context.Context, authUserID primitive.ObjectID, registerReq *model.RegisterRequestDTO) error
	Login(c context.Context, userReq model.LoginRequestDTO) (*model.LoginResponseDTO, error)
	GetUserByID(c context.Context, userID primitive.ObjectID) (*model.UserResponseDTO, error)
	UpdateUserByID(c context.Context, userID primitive.ObjectID, authUserID primitive.ObjectID, user *model.UpdateRequestDTO) (*model.UserResponseDTO, error)
	GetAllUsers(c context.Context) (*[]model.UserResponseDTO, error)
}

type userUsecase struct {
	userRepository    model.UserRepository
	roleRepository    model.RoleRepository
	profileRepository model.ProfileRepository
	contextTimeout    time.Duration
	client            *mongo.Client
}

func NewUserUsecase(userRepository model.UserRepository, roleRepository model.RoleRepository, profileRepository model.ProfileRepository, timeout time.Duration, client *mongo.Client) UserUsecase {
	return &userUsecase{
		userRepository:    userRepository,
		roleRepository:    roleRepository,
		profileRepository: profileRepository,
		contextTimeout:    timeout,
		client:            client,
	}
}

func (uc *userUsecase) Register(c context.Context, authUserID primitive.ObjectID, registerReq *model.RegisterRequestDTO) error {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	if registerReq.BranchID == nil && registerReq.DepartmentID == nil {
		return errors.New("either BranchID or DepartmentID must be provided")
	}

	existingUser, err := uc.userRepository.FindByUsername(ctx, registerReq.Username)
	if err == nil && existingUser.Username != "" {
		return errors.New("username already exists")
	}

	encriptedPassword, err := bcrypt.GenerateFromPassword([]byte(registerReq.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// start MongoDB session
	session, err := uc.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	// Run everything in the transaction
	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		// populate profile
		profile := &model.Profile{
			ID:           primitive.NewObjectID(),
			FirstName:    registerReq.FirstName,
			MiddleName:   registerReq.MiddleName,
			LastName:     registerReq.LastName,
			Email:        registerReq.Email,
			DepartmentID: registerReq.DepartmentID,
			BranchID:     registerReq.BranchID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = uc.profileRepository.Create(ctx, profile)
		if err != nil {
			return err
		}

		user := &model.User{
			ID:        primitive.NewObjectID(),
			ProfileID: profile.ID,
			RoleID:    registerReq.Role,
			Username:  registerReq.Username,
			Password:  string(encriptedPassword),
			Status:    "New",
			CreatedBy: authUserID,
			IsDeleted: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := uc.userRepository.Create(sessCtx, user); err != nil {
			session.AbortTransaction(sessCtx)
			return err
		}

		return session.CommitTransaction(sessCtx)

	})
	return err
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

	fmt.Println("profile", profile)

	response := model.LoginResponseDTO{
		ID:         existingUser.ID,
		FirstName:  profile.FirstName,
		MiddleName: profile.LastName,
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

func (uc *userUsecase) UpdateUserByID(ctx context.Context, user_id primitive.ObjectID, authUserID primitive.ObjectID, user *model.UpdateRequestDTO) (*model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	// populate profile
	profile := &model.Profile{
		FirstName:    user.FirstName,
		MiddleName:   user.MiddleName,
		LastName:     user.LastName,
		Email:        user.Email,
		DepartmentID: user.DepartmentID,
		BranchID:     user.BranchID,
		UpdatedAt:    time.Now(),
	}

	err := uc.profileRepository.Create(ctx, profile)
	if err != nil {
		return nil, err
	}

	updatedUser := &model.User{
		RoleID:    user.Role,
		Username:  user.Username,
		UpdatedBy: &authUserID,
		UpdatedAt: time.Now(),
	}

	return uc.userRepository.Update(ctx, user_id, updatedUser)
}

func (uc *userUsecase) GetAllUsers(ctx context.Context) (*[]model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindAll(ctx)
}

package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jinzhu/copier"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecase interface {
	Register(c context.Context, authUserID primitive.ObjectID, registerReq *model.RegisterRequestDTO) error
	Login(c context.Context, userReq model.LoginRequestDTO) (*model.LoginResponseDTO, error)
	GetUserByID(c context.Context, userID primitive.ObjectID) (*model.User, error)
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
			session.AbortTransaction(sessCtx)
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

	accessToken, err := infrastructure.GenerateToken(existingUser.ID, role.Name)
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
		Role:       role.Name,
		Token:      accessToken,
	}

	return &response, nil
}

func (uc *userUsecase) GetUserByID(ctx context.Context, user_id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindByID(ctx, user_id)
}

func (uc *userUsecase) UpdateUserByID(ctx context.Context, user_id primitive.ObjectID, authUserID primitive.ObjectID, user *model.UpdateRequestDTO) (*model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	// start MongoDB session
	session, err := uc.client.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	responseUser := model.UserResponseDTO{}

	// Run everything in the transaction
	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return err
		}

		existingUser, err := uc.userRepository.FindByID(sessCtx, user_id)
		if err != nil {
			return err
		}

		existingProfile, err := uc.profileRepository.FindByID(sessCtx, existingUser.ProfileID)
		if err != nil {
			return err
		}

		// Populate profile
		if user.FirstName != "" {
			existingProfile.FirstName = user.FirstName
		}
		if user.MiddleName != "" {
			existingProfile.MiddleName = user.MiddleName
		}
		if user.LastName != "" {
			existingProfile.LastName = user.LastName
		}
		if user.Email != "" {
			existingProfile.Email = user.Email
		}
		if user.BranchID != &primitive.NilObjectID {
			existingProfile.BranchID = user.BranchID
		}
		if user.DepartmentID != &primitive.NilObjectID {
			existingProfile.DepartmentID = user.DepartmentID
		}
		existingProfile.UpdatedAt = time.Now()

		// Populate user
		if user.Role != primitive.NilObjectID {
			existingUser.RoleID = user.Role
		}
		if user.Username != "" {
			existingUser.UpdatedBy = &authUserID
		}
		existingUser.UpdatedBy = &authUserID
		existingUser.UpdatedAt = time.Now()

		// Database update for both profile and user
		updateProfile, err := uc.profileRepository.Update(sessCtx, existingUser.ProfileID, existingProfile)
		if err != nil {
			session.AbortTransaction(sessCtx)
			return err
		}

		updatedUser, err := uc.userRepository.Update(sessCtx, user_id, existingUser)
		if err != nil {
			session.AbortTransaction(sessCtx)
			return err
		}

		copier.Copy(&responseUser, updatedUser)
		copier.Copy(&responseUser, updateProfile)
		return session.CommitTransaction(sessCtx)
	})

	if err != nil {
		return nil, err
	}
	return &responseUser, nil
}

func (uc *userUsecase) GetAllUsers(ctx context.Context) (*[]model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindAll(ctx)
}

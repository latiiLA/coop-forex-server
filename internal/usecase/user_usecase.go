package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jinzhu/copier"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserUsecase interface {
	Register(c context.Context, authUserID primitive.ObjectID, registerReq *model.RegisterUsecaseRequestDTO) error
	Login(c context.Context, userReq model.LoginRequestDTO, ip string) (*model.LoginResponseDTO, error)
	Logout(c context.Context, userID primitive.ObjectID, accessToken string, refreshToken string, clientIP string) error
	GetUserByID(c context.Context, userID primitive.ObjectID) (*model.User, error)
	UpdateUserByID(c context.Context, userID primitive.ObjectID, authUserID primitive.ObjectID, user *model.UpdateUserRequestDTO) (*model.UserResponseDTO, error)
	GetAllUsers(c context.Context) (*[]model.UserResponseDTO, error)
	RefreshToken(c context.Context, refresh model.RefreshTokenDTO, clientIP string) (string, string, error)
	SuspendUser(c context.Context, authUserID, targetUserID primitive.ObjectID) error
	UnsuspendUser(c context.Context, authUserID, targetUserID primitive.ObjectID) error
	GetUserStatusSummary(c context.Context) (*model.UserStatusSummary, error)
	GetEligibleApprovers(c context.Context) (*[]model.UserResponseDTO, error)
	GetEligibleBranchAuthorizers(c context.Context, branchID primitive.ObjectID) (*[]model.UserResponseDTO, error)
}

type userUsecase struct {
	userRepository     model.UserRepository
	roleRepository     model.RoleRepository
	profileRepository  model.ProfileRepository
	tokenBlacklistRepo model.TokenBlacklistRepository
	activityLog        model.ActivityLogUsecase
	contextTimeout     time.Duration
	client             *mongo.Client
}

func NewUserUsecase(
	userRepository model.UserRepository,
	roleRepository model.RoleRepository,
	profileRepository model.ProfileRepository,
	tokenBlacklistRepo model.TokenBlacklistRepository,
	activityLog model.ActivityLogUsecase,
	timeout time.Duration,
	client *mongo.Client,
) UserUsecase {
	return &userUsecase{
		userRepository:     userRepository,
		roleRepository:     roleRepository,
		profileRepository:  profileRepository,
		tokenBlacklistRepo: tokenBlacklistRepo,
		activityLog:        activityLog,
		contextTimeout:     timeout,
		client:             client,
	}
}

func (uc *userUsecase) Register(c context.Context, authUserID primitive.ObjectID, registerReq *model.RegisterUsecaseRequestDTO) error {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	if registerReq.BranchID == nil && registerReq.DepartmentID == nil {
		return common.ErrBranchOrDepartmentNotFound
	}

	existingUser, err := uc.userRepository.FindByUsername(ctx, registerReq.Username)
	if err == nil && existingUser.Username != "" {
		return common.ErrUsernameAlreadyExists
	}

	var createdUserID primitive.ObjectID

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
			DisplayName:  registerReq.DisplayName,
			Email:        registerReq.Email,
			DepartmentID: registerReq.DepartmentID,
			BranchID:     registerReq.BranchID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		err = uc.profileRepository.Create(sessCtx, profile)
		if err != nil {
			session.AbortTransaction(sessCtx)
			return err
		}

		user := &model.User{
			ID:          primitive.NewObjectID(),
			ProfileID:   profile.ID,
			RoleID:      registerReq.Role,
			Username:    registerReq.Username,
			Status:      model.StatusNew,
			CreatedBy:   authUserID,
			Permissions: registerReq.Permissions,
			IsDeleted:   false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := uc.userRepository.Create(sessCtx, user); err != nil {
			session.AbortTransaction(sessCtx)
			return err
		}

		createdUserID = user.ID
		return session.CommitTransaction(sessCtx)
	})

	if err != nil {
		return err
	}

	entityID := createdUserID
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionUserCreate,
		EntityType:  model.EntityUser,
		EntityID:    &entityID,
		Description: "User created",
		ToState:     string(model.StatusNew),
	})

	return nil
}

func (uc *userUsecase) Login(c context.Context, userReq model.LoginRequestDTO, ip string) (*model.LoginResponseDTO, error) {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	existingUser, err := uc.userRepository.FindByUsername(ctx, userReq.Username)
	if err != nil {
		fmt.Println("Invalid username or user", err)
		return nil, common.ErrInvalidCredentials
	}

	if existingUser.Status != "active" && existingUser.Status != "new" {
		return nil, common.ErrUserAccessRevoked
	}

	err = infrastructure.CheckPasswordHash(existingUser.Password, userReq.Password)
	if err != nil {
		fmt.Println("Invalid password:", err)
		return nil, common.ErrInvalidCredentials
	}

	var perms []string
	if existingUser.Permissions != nil {
		perms = existingUser.Permissions
	}

	effectivePerms := utils.MergePermissions(existingUser.Role.Permissions, perms)

	var branchID primitive.ObjectID
	if existingUser.Profile.BranchID != nil {
		branchID = *existingUser.Profile.BranchID
	} else {
		branchID = primitive.NilObjectID // fallback
	}

	var departmentID primitive.ObjectID
	if existingUser.Profile.DepartmentID != nil {
		departmentID = *existingUser.Profile.DepartmentID
	} else {
		departmentID = primitive.NilObjectID // fallback
	}

	accessToken, err := infrastructure.GenerateToken(existingUser.ID, existingUser.Role.Name, branchID, departmentID, effectivePerms, ip)
	if err != nil {
		return nil, err
	}

	refeshToken, err := infrastructure.GenerateRefreshToken(existingUser.ID, ip)
	if err != nil {
		return nil, err
	}

	response := model.LoginResponseDTO{
		ID:           existingUser.ID,
		FirstName:    existingUser.Profile.FirstName,
		MiddleName:   existingUser.Profile.MiddleName,
		Username:     existingUser.Username,
		Email:        existingUser.Profile.Email,
		Role:         existingUser.Role.Name,
		Permissions:  effectivePerms,
		Signature:    existingUser.Signature,
		Token:        accessToken,
		RefreshToken: refeshToken,
	}

	actorID := existingUser.ID
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     actorID,
		Action:      model.ActionLogin,
		EntityType:  model.EntityAuth,
		EntityID:    &actorID,
		Description: "User logged in",
		Metadata: map[string]any{
			"ip": ip,
		},
	})

	return &response, nil
}

func (uc *userUsecase) Logout(c context.Context, userID primitive.ObjectID, accessToken string, refreshToken string, clientIP string) error {
	ctx, cancel := context.WithTimeout(c, uc.contextTimeout)
	defer cancel()

	if accessToken != "" && uc.tokenBlacklistRepo != nil {
		expTime, err := parseJWTExpiry(accessToken)
		if err != nil {
			return err
		}
		if err := uc.tokenBlacklistRepo.BlacklistToken(ctx, accessToken, expTime, userID, clientIP); err != nil {
			return err
		}
	}

	if refreshToken != "" && uc.tokenBlacklistRepo != nil {
		expTime, err := parseJWTExpiry(refreshToken)
		if err == nil {
			_ = uc.tokenBlacklistRepo.BlacklistToken(ctx, refreshToken, expTime, userID, clientIP)
		}
	}

	entityID := userID
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     userID,
		Action:      model.ActionLogout,
		EntityType:  model.EntityAuth,
		EntityID:    &entityID,
		Description: "User logged out",
		Metadata: map[string]any{
			"ip": clientIP,
		},
	})

	return nil
}

func parseJWTExpiry(tokenString string) (time.Time, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid token claims")
	}
	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("invalid expiration in token")
	}
	return time.Unix(int64(expUnix), 0), nil
}

func (uc *userUsecase) GetUserByID(ctx context.Context, user_id primitive.ObjectID) (*model.User, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindByID(ctx, user_id)
}

func (uc *userUsecase) UpdateUserByID(ctx context.Context, user_id primitive.ObjectID, authUserID primitive.ObjectID, user *model.UpdateUserRequestDTO) (*model.UserResponseDTO, error) {
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

		logrus.Println(existingUser.Profile.ID, user.Username, "existing user")

		existingProfile, err := uc.profileRepository.FindByID(sessCtx, existingUser.Profile.ID)
		if err != nil {
			logrus.Println(err)
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
		if user.BranchID != &primitive.NilObjectID {
			existingProfile.DepartmentID = nil
			existingProfile.BranchID = user.BranchID
		}
		if user.DepartmentID != &primitive.NilObjectID {
			existingProfile.BranchID = nil
			existingProfile.DepartmentID = user.DepartmentID
		}

		// Populate user
		if user.Role != primitive.NilObjectID {
			existingUser.RoleID = user.Role
		}
		if user.Username != "" {
			existingUserByUsername, err := uc.userRepository.FindByUsername(sessCtx, user.Username)
			if err != nil && err != mongo.ErrNoDocuments {
				return err
			}
			if existingUserByUsername != nil && existingUserByUsername.ID != user_id {
				return errors.New("username already exists")
			}
			existingUser.Username = user.Username
		}
		if user.Password != "" {
			encrytedPassword, err := infrastructure.HashPassword(user.Password)
			if err != nil {
				return err
			}
			existingUser.Password = encrytedPassword
		}
		existingUser.UpdatedBy = &authUserID
		existingUser.UpdatedAt = time.Now()

		// Database update for both profile and user
		updateProfile, err := uc.profileRepository.Update(sessCtx, existingUser.Profile.ID, existingProfile)
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

	entityID := user_id
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionUserUpdate,
		EntityType:  model.EntityUser,
		EntityID:    &entityID,
		Description: "User updated",
	})

	return &responseUser, nil
}

func (uc *userUsecase) GetAllUsers(ctx context.Context) (*[]model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindAll(ctx)
}

func (uc *userUsecase) SuspendUser(ctx context.Context, authUserID, targetUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	if authUserID == targetUserID {
		return common.ErrSelfSuspension
	}

	existing, err := uc.userRepository.FindByID(ctx, targetUserID)
	if err != nil || existing == nil {
		return common.ErrUserNotFound
	}

	fromStatus := string(existing.Status)
	if _, err := uc.userRepository.UpdateStatus(ctx, targetUserID, model.StatusSuspended, authUserID); err != nil {
		return err
	}

	entityID := targetUserID
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionUserSuspend,
		EntityType:  model.EntityUser,
		EntityID:    &entityID,
		Description: "User suspended",
		FromState:   fromStatus,
		ToState:     string(model.StatusSuspended),
	})
	return nil
}

func (uc *userUsecase) UnsuspendUser(ctx context.Context, authUserID, targetUserID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()

	existing, err := uc.userRepository.FindByID(ctx, targetUserID)
	if err != nil || existing == nil {
		return common.ErrUserNotFound
	}

	if existing.Status != model.StatusSuspended {
		return common.ErrInvalidUserStatus
	}

	fromStatus := string(existing.Status)
	if _, err := uc.userRepository.UpdateStatus(ctx, targetUserID, model.StatusActive, authUserID); err != nil {
		return err
	}

	entityID := targetUserID
	LogActivity(uc.activityLog, ctx, model.ActivityLogInput{
		ActorID:     authUserID,
		Action:      model.ActionUserUnsuspend,
		EntityType:  model.EntityUser,
		EntityID:    &entityID,
		Description: "User unsuspended",
		FromState:   fromStatus,
		ToState:     string(model.StatusActive),
	})
	return nil
}

func (uc *userUsecase) GetUserStatusSummary(ctx context.Context) (*model.UserStatusSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.CountByStatus(ctx)
}

func (uc *userUsecase) GetEligibleApprovers(ctx context.Context) (*[]model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindEligibleApprovers(ctx)
}

func (uc *userUsecase) GetEligibleBranchAuthorizers(ctx context.Context, branchID primitive.ObjectID) (*[]model.UserResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, uc.contextTimeout)
	defer cancel()
	return uc.userRepository.FindEligibleBranchAuthorizers(ctx, branchID)
}

func (a *userUsecase) RefreshToken(ctx context.Context, refreshToken model.RefreshTokenDTO, clientIP string) (string, string, error) {
	// 1. Validate the refresh token
	claims, err := infrastructure.ValidateRefreshToken(refreshToken.RefreshToken, clientIP)
	if err != nil {
		return "", "", err
	}

	// 2. Check if token is blacklisted
	isBlacklisted, err := a.tokenBlacklistRepo.IsBlacklisted(ctx, refreshToken.RefreshToken)
	if err != nil {
		return "", "", err
	}
	if isBlacklisted {
		return "", "", fmt.Errorf("Jwt is already black listed")
	}

	// 4. Retrieve user to generate new tokens and blacklist
	userIDHex, ok := claims["userID"].(string)
	if !ok {
		return "", "", fmt.Errorf("userID missing or invalid in token")
	}

	userObjID, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		return "", "", err
	}

	// 3. Blacklist the current refresh token
	expUnix, ok := claims["exp"].(float64)
	if !ok {
		return "", "", fmt.Errorf("invalid expiration in token")
	}
	expTime := time.Unix(int64(expUnix), 0)

	if err := a.tokenBlacklistRepo.BlacklistToken(ctx, refreshToken.RefreshToken, expTime, userObjID, clientIP); err != nil {
		return "", "", fmt.Errorf("failed to blacklist token")
	}

	user, err := a.userRepository.FindByID(ctx, userObjID)
	if err != nil {
		return "", "", err
	}
	if user == nil {
		return "", "", fmt.Errorf("user not found")
	}

	// 5. Retrieve user role
	role, err := a.roleRepository.FindByID(ctx, user.Role.ID)
	if err != nil {
		return "", "", err
	}
	if role == nil {
		return "", "", fmt.Errorf("role not found")
	}

	// 6. Generate new token pair (access + refresh)
	branchID := primitive.NilObjectID
	if user.Profile.BranchID != nil {
		branchID = *user.Profile.BranchID
	}

	departmentID := primitive.NilObjectID
	if user.Profile.DepartmentID != nil {
		departmentID = *user.Profile.DepartmentID
	}

	userPerms := []string{}
	if user.Permissions != nil {
		userPerms = user.Permissions
	}
	// Must include role permissions (same as Login) or refreshed tokens lose request:send/etc.
	permissions := utils.MergePermissions(role.Permissions, userPerms)

	newAccessToken, err := infrastructure.GenerateToken(user.ID, role.Name, branchID, departmentID, permissions, clientIP)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := infrastructure.GenerateRefreshToken(user.ID, clientIP)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

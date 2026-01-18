package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthUsecase interface {
	Authenticate(ctx context.Context, username, password, ip string) (*model.LoginResponseDTO, error)
}

type ldapAuthUsecase struct {
	userRepo     model.UserRepository
	Host         string
	Port         string
	BasedDN      string
	BindUser     string
	BindPassword string
	UserFilter   string // Add this: "uid" for mock, "sAMAccountName" for real AD
	timeout      time.Duration
}

func NewLDAPAuthUsecase(userRepo model.UserRepository, host string, port, baseDN, bindUser, bindPassword, userFilter string, timeout time.Duration) AuthUsecase {
	return &ldapAuthUsecase{
		userRepo:     userRepo,
		Host:         host,
		Port:         port,
		BasedDN:      baseDN,
		BindUser:     bindUser,
		BindPassword: bindPassword,
		UserFilter:   userFilter,
		timeout:      timeout,
	}
}

func (s *ldapAuthUsecase) Authenticate(ctx context.Context, username, password, ip string) (*model.LoginResponseDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	// Step 1 - check local database
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		logrus.Println("Invalid username or user", err)
		return nil, errors.New("invalid username or password")
	}

	if existingUser.Status != "new" && existingUser.Status != "active" {
		logrus.Println("Invalid username or user", err)
		return nil, errors.New("user access has been revoced or user is deleted")
	}

	// Step 2 - check ldap authentications
	log.Printf("DEBUG: LDAP Host = %s, Port = %s\n", s.Host, s.Port)
	log.Printf("DEBUG: Dialing URL = ldap://%s:%s\n", s.Host, s.Port)

	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s", s.Host))
	if err != nil {
		log.Println("Connection failed")
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer l.Close()
	log.Println("Connection successful")

	log.Println(s.BindUser, s.BindPassword)

	// Step 2.1: Search for user DN
	searchRequest := ldap.NewSearchRequest(
		s.BasedDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(%s=%s)", s.UserFilter, username),
		[]string{"dn"},
		nil,
	)

	fmt.Println("search request", searchRequest)

	// Step 2.2: Bind with service account
	err = l.Bind(s.BindUser, s.BindPassword)
	if err != nil {
		log.Println(fmt.Errorf("bind failed: %w", err))
		return nil, fmt.Errorf("bind failed: %w", err)
	}
	log.Println("DEBUG: Initial bind successful")

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Println(fmt.Errorf("LDAP search error: %w", err))
		return nil, fmt.Errorf("LDAP search error: %w", err)
	}
	if len(sr.Entries) == 0 {
		log.Println(fmt.Errorf("user not found"))
		return nil, fmt.Errorf("user not found")
	}

	userDN := sr.Entries[0].DN
	log.Printf("DEBUG: Found user DN = %s\n", userDN)

	// Step 2.3: Try binding as the user with the provided password
	err = l.Bind(userDN, password)
	if err != nil {
		log.Println(fmt.Errorf("user authentication failed: %w", err))
		return nil, fmt.Errorf("authentication failed: invalid username or password")
	}
	log.Println("✅ User authentication successful")

	// Step 3 Prepare response and reply
	var perms []string
	if existingUser.Permissions != nil {
		perms = *existingUser.Permissions
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

	// // Update last login
	// var now = time.Now()
	// existingUser.LastLogin = &now

	// if _, err := uc.userRepository.Update(ctx, existingUser.ID, existingUser); err != nil {
	// 	return nil, fmt.Errorf("user login failed %s", err)
	// }

	response := model.LoginResponseDTO{
		ID:           existingUser.ID,
		FirstName:    existingUser.Profile.FirstName,
		MiddleName:   existingUser.Profile.MiddleName,
		Username:     existingUser.Username,
		Email:        existingUser.Profile.Email,
		Role:         existingUser.Role.Name,
		Permissions:  effectivePerms,
		Token:        accessToken,
		RefreshToken: refeshToken,
	}

	return &response, nil
}

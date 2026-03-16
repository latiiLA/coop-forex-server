package usecase

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-ldap/ldap/v3"
	"github.com/latiiLA/coop-forex-server/internal/common"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure"
	"github.com/latiiLA/coop-forex-server/internal/infrastructure/utils"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthUsecase interface {
	Authenticate(ctx context.Context, username, password, ip string) (*model.LoginResponseDTO, error)
	GetUserDetails(ctx context.Context, username string) (*model.User, error)
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

	// check local database
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		logrus.Println("invalid username or user", err)
		if err == mongo.ErrNoDocuments {
			return nil, common.ErrUserNotFound
		}

		return nil, common.ErrInvalidCredentials
	}

	if existingUser.Status != model.StatusNew && existingUser.Status != model.StatusActive {
		logrus.Println("user status is active", err)
		return nil, common.ErrUserAccessRevoked
	}

	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s", s.Host))
	if err != nil {
		log.Println("LDAP: Connection failed")
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer l.Close()

	// Search for user DN
	searchRequest := ldap.NewSearchRequest(
		s.BasedDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(%s=%s)", s.UserFilter, username),
		[]string{"dn"},
		nil,
	)

	// Bind with service account
	err = l.Bind(s.BindUser, s.BindPassword)
	if err != nil {
		log.Println(fmt.Errorf("bind failed: %w", err))
		return nil, fmt.Errorf("bind failed: %w", err)
	}
	log.Println("LDAP: Initial bind successful")

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Println(fmt.Errorf("LDAP search error: %w", err))
		return nil, fmt.Errorf("LDAP search error: %w", err)
	}
	if len(sr.Entries) == 0 {
		log.Println(fmt.Errorf("user not found"))
		return nil, common.ErrADUserNotFound
	}

	userDN := sr.Entries[0].DN
	log.Printf("LDAP: Found user DN = %s\n", userDN)

	// Try binding as the user with the provided password
	err = l.Bind(userDN, password)
	if err != nil {
		log.Println(fmt.Errorf("user authentication failed: %w", err))
		return nil, common.ErrInvalidCredentials
	}
	log.Println("✅ User authentication successful")

	// Prepare response and reply
	var perms []string
	if existingUser.Role != nil && existingUser.Permissions != nil {
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

	refreshToken, err := infrastructure.GenerateRefreshToken(existingUser.ID, ip)
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
		Token:        accessToken,
		RefreshToken: refreshToken,
	}

	// Update last login
	var now = time.Now()
	if existingUser.Status == model.StatusNew {
		existingUser.Status = model.StatusActive
	}
	existingUser.LastLogin = &now
	existingUser.Updater = nil
	existingUser.Creator = nil
	existingUser.Profile = nil
	existingUser.Role = nil

	if _, err := s.userRepo.Update(ctx, existingUser.ID, existingUser); err != nil {
		return nil, fmt.Errorf("user login failed %s", err)
	}

	return &response, nil
}

// Inside your LDAP service
func (s *ldapAuthUsecase) GetUserDetails(ctx context.Context, username string) (*model.User, error) {
	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s:%s", s.Host, s.Port))
	if err != nil {
		return nil, err
	}
	defer l.Close()

	// Bind with Service Account
	err = l.Bind(s.BindUser, s.BindPassword)
	if err != nil {
		return nil, err
	}

	// Search for the user
	searchRequest := ldap.NewSearchRequest(
		s.BasedDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(%s=%s)", s.UserFilter, username), // Use sAMAccountName for real AD
		[]string{"dn", "givenName", "name", "sn", "mail", "userAccountControl"},
		// []string{"*"}, // used for debug to fetch all the data
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		return nil, common.ErrADUserNotFound
	}

	entry := sr.Entries[0]

	// log.Println("===== LDAP ATTRIBUTES DUMP =====")

	// for _, attr := range entry.Attributes {
	// 	log.Printf("Attribute: %s\n", attr.Name)
	// 	for i, val := range attr.Values {
	// 		log.Printf("   Value[%d]: %s\n", i, val)
	// 	}
	// }

	// log.Println("================================")

	profile := model.Profile{
		DisplayName: entry.GetAttributeValue("name"),
		FirstName:   entry.GetAttributeValue("givenName"),
		MiddleName:  entry.GetAttributeValue("sn"),
		Email:       entry.GetAttributeValue("mail"),
	}
	return &model.User{
		Profile: &profile,
	}, nil
}

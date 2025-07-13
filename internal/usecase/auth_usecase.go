package usecase

import (
	"fmt"
	"log"

	"github.com/go-ldap/ldap/v3"
)

type AuthUsecase interface {
	Authenticate(username, password string) error
}

type ldapAuthUsecase struct {
	Host         string
	Port         string
	BasedDN      string
	BindUser     string
	BindPassword string
}

func NewLDAPAuthUsecase(host string, port, baseDN, bindUser, bindPassword string) AuthUsecase {
	return &ldapAuthUsecase{
		Host:         host,
		Port:         port,
		BasedDN:      baseDN,
		BindUser:     bindUser,
		BindPassword: bindPassword,
	}
}

func (s *ldapAuthUsecase) Authenticate(username, password string) error {
	log.Printf("DEBUG: LDAP Host = %s, Port = %s\n", s.Host, s.Port)
	log.Printf("DEBUG: Dialing URL = ldap://%s:%s\n", s.Host, s.Port)

	l, err := ldap.DialURL(fmt.Sprintf("ldap://%s", s.Host))
	if err != nil {
		log.Println("Connection failed")
		return fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer l.Close()
	log.Println("Connection successful")

	log.Println(s.BindUser, s.BindPassword)

	// Step 2: Search for user DN
	searchRequest := ldap.NewSearchRequest(
		s.BasedDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"dn"},
		nil,
	)

	fmt.Println("search request", searchRequest)

	// Step 1: Bind with service account
	err = l.Bind(s.BindUser, s.BindPassword)
	if err != nil {
		log.Println(fmt.Errorf("bind failed: %w", err))
		return fmt.Errorf("bind failed: %w", err)
	}
	log.Println("DEBUG: Initial bind successful")

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Println(fmt.Errorf("LDAP search error: %w", err))
		return fmt.Errorf("LDAP search error: %w", err)
	}
	if len(sr.Entries) == 0 {
		log.Println(fmt.Errorf("user not found"))
		return fmt.Errorf("user not found")
	}

	userDN := sr.Entries[0].DN
	log.Printf("DEBUG: Found user DN = %s\n", userDN)

	// Step 3: Try binding as the user with the provided password
	err = l.Bind(userDN, password)
	if err != nil {
		log.Println(fmt.Errorf("user authentication failed: %w", err))
		return fmt.Errorf("authentication failed: invalid username or password")
	}
	log.Println("✅ User authentication successful")

	return nil
}

package model

type ldapAuthService struct {
	Host         string
	Port         int
	BaseDN       string
	BindUser     string
	BindPassword string
}

type LDAPAuthService interface {
	Authenticate(username, password string) error
}

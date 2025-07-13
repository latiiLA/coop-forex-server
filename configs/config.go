package configs

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

var (
	JwtSecret        string
	MongoURL         string
	Timeout          time.Duration
	DisableMigration string
	FileUploadPath   string

	// Ldap configs
	LDAPHost         string
	LDAPPort         string
	LDAPBaseDN       string
	LDAPBindUser     string
	LDAPBindPassword string
)

func LoadConfig() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found or couldn't load it, relying on environment variables")
	}

	MongoURL = os.Getenv("MONGO_URL")
	if MongoURL == "" {
		log.Fatal("Mongo url is required but not set")
	}

	DisableMigration = os.Getenv("DISABLE_MIGRATION")
	if DisableMigration == "" {
		log.Fatal("DisableMigration status is required but not set")
	}

	JwtSecret = os.Getenv("JWT_SECRET")
	if JwtSecret == "" {
		log.Fatal("JWT_SECRET is required but not set")
	}

	timeoutStr := os.Getenv("APP_TIMEOUT")
	if timeoutStr == "" {
		log.Fatalf("APP_TIMEOUT is required but not set")
	}

	Timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		log.Fatalf("Invalid APP_TIMEOUT format: %v", err)
	}

	FileUploadPath = os.Getenv("FILE_UPLOAD_PATH")
	if FileUploadPath == "" {
		log.Fatal("FILE_UPLOAD_PATH is required but not set")
	}

	LDAPHost = os.Getenv("LDAP_HOST")
	if LDAPHost == "" {
		log.Fatalf("LDAP Host is required but not set")
	}

	LDAPPort = os.Getenv("LDAP_PORT")
	if LDAPPort == "" {
		log.Fatalf("LDAP Port is required but not set")
	}

	LDAPBaseDN = os.Getenv("LDAP_BASE_DN")
	if LDAPBaseDN == "" {
		log.Fatalf("LDAP base DN is required but not set")
	}

	LDAPBindUser = os.Getenv("LDAP_BIND_USER")
	if LDAPBindUser == "" {
		log.Fatalf("LDAP bind user is required but not set")
	}

	LDAPBindPassword = os.Getenv("LDAP_BIND_PASSWORD")
	if LDAPBindPassword == "" {
		log.Fatalf("LDAP bind password is required but not set")
	}

	fmt.Println(LDAPHost, LDAPPort, LDAPBaseDN, LDAPBindUser, LDAPBindPassword)
}

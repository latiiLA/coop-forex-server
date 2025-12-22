package configs

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	JwtSecret        string
	MongoURL         string
	Timeout          time.Duration
	DisableMigration string
	FileUploadPath   string
	LogLevel         string

	// Mail env
	MailServer   string
	MailUsername string
	MailPassword string
	MailPort     string
	MailSender   string
	MailReciever string

	// Ldap configs
	LDAPHost         string
	LDAPPort         string
	LDAPBaseDN       string
	LDAPBindUser     string
	LDAPBindPassword string

	// ssl
	CertFile string
	KeyFile  string

	AllowedOrigins []string
)

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or couldn't load it, relying on environment variables", err)
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

	LogLevel = os.Getenv("LOG_LEVEL")
	if LogLevel == "" {
		log.Fatal("LOG_LEVEL is required but not set")
	}

	// MAIL server

	MailServer = os.Getenv("MAIL_SERVER")
	if MailServer == "" {
		log.Fatal("MAIL_SERVER is required but not set")
	}

	MailUsername = os.Getenv("MAIL_USERNAME")
	if MailUsername == "" {
		log.Fatal("MAIL_USERNAME is required but not set")
	}

	MailUsername = os.Getenv("MAIL_USERNAME")
	if MailUsername == "" {
		log.Fatal("MAIL_USERNAME is required but not set")
	}

	MailPassword = os.Getenv("MAIL_PASSWORD")
	if MailPassword == "" {
		log.Fatal("MAIL_PASSWORD is required but not set")
	}

	MailPort = os.Getenv("MAIL_PORT")
	if MailPort == "" {
		log.Fatal("MAIL_PORT is required but not set")
	}

	MailReciever = os.Getenv("MAIL_RECIEVER")
	if MailReciever == "" {
		log.Fatal("MAIL_RECIEVER is required but not set")
	}

	MailSender = os.Getenv("MAIL_SENDER")
	if MailSender == "" {
		log.Fatal("MAIL_SENDER is required but not set")
	}

	// ssl
	CertFile = os.Getenv("CERT_FILE")
	if CertFile == "" {
		log.Fatal("CERT_FILE is required but not set")
	}

	KeyFile = os.Getenv("KEY_FILE")
	if KeyFile == "" {
		log.Fatal("KEY_FILE is required but not set")
	}

	// LDAP env

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

	// Read allowed origins from environment and split into slice
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv != "" {
		AllowedOrigins = strings.Split(originsEnv, ",")
	}
}

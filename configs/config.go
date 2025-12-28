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

	MailRequestCreatedTo  []string
	MailRequestCreatedCc  []string
	MailRequestCreatedBcc []string

	MailRequestSentTo  []string
	MailRequestSentCc  []string
	MailRequestSentBcc []string

	MailRequestAuthorizedTo  []string
	MailRequestAuthorizedCc  []string
	MailRequestAuthorizedBcc []string

	MailRequestValidatedTo  []string
	MailRequestValidatedCc  []string
	MailRequestValidatedBcc []string

	MailRequestRejectedTo  []string
	MailRequestRejectedCc  []string
	MailRequestRejectedBcc []string

	MailRequestApprovedTo  []string
	MailRequestApprovedCc  []string
	MailRequestApprovedBcc []string

	MailRequestAcceptedTo  []string
	MailRequestAcceptedCc  []string
	MailRequestAcceptedBcc []string

	MailRequestDeclinedTo  []string
	MailRequestDeclinedCc  []string
	MailRequestDeclinedBcc []string

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

	MailPassword = os.Getenv("MAIL_PASSWORD")
	if MailPassword == "" {
		log.Fatal("MAIL_PASSWORD is required but not set")
	}

	MailPort = os.Getenv("MAIL_PORT")
	if MailPort == "" {
		log.Fatal("MAIL_PORT is required but not set")
	}

	// ---------------------------------SEND EMAIL VARIABLES---------------------
	MailRequestCreatedTo = LoadEmailsFromEnv("MAIL_REQUEST_CREATED_TO")
	if len(MailRequestCreatedTo) == 0 {
		log.Fatal("MAIL_REQUEST_CREATED_TO is required but not set")
	}
	MailRequestCreatedCc = LoadEmailsFromEnv("MAIL_REQUEST_CREATED_CC")
	if len(MailRequestCreatedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_CREATED_CC is not set")
	}
	MailRequestCreatedBcc = LoadEmailsFromEnv("MAIL_REQUEST_CREATED_BCC")
	if len(MailRequestCreatedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_CREATED_BCC is not set")
	}

	MailRequestSentTo = LoadEmailsFromEnv("MAIL_REQUEST_SENT_TO")
	if len(MailRequestSentTo) == 0 {
		log.Fatal("MAIL_REQUEST_SENT_TO is required but not set")
	}
	MailRequestSentCc = LoadEmailsFromEnv("MAIL_REQUEST_SENT_CC")
	if len(MailRequestSentCc) == 0 {
		log.Print("Info: MAIL_REQUEST_SENT_CC is not set")
	}
	MailRequestSentBcc = LoadEmailsFromEnv("MAIL_REQUEST_SENT_BCC")
	if len(MailRequestSentBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_SENT_BCC is not set")
	}

	MailRequestAuthorizedTo = LoadEmailsFromEnv("MAIL_REQUEST_AUTHORIZED_TO")
	if len(MailRequestAuthorizedTo) == 0 {
		log.Fatal("MAIL_REQUEST_AUTHORIZED_TO is required but not set")
	}
	MailRequestAuthorizedCc = LoadEmailsFromEnv("MAIL_REQUEST_AUTHORIZED_CC")
	if len(MailRequestAuthorizedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_AUTHORIZED_CC is not set")
	}
	MailRequestAuthorizedBcc = LoadEmailsFromEnv("MAIL_REQUEST_AUTHORIZED_BCC")
	if len(MailRequestAuthorizedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_AUTHORIZED_BCC is not set")
	}

	MailRequestValidatedTo = LoadEmailsFromEnv("MAIL_REQUEST_VALIDATED_TO")
	if len(MailRequestValidatedTo) == 0 {
		log.Fatal("MAIL_REQUEST_VALIDATED_TO is required but not set")
	}
	MailRequestValidatedCc = LoadEmailsFromEnv("MAIL_REQUEST_VALIDATED_CC")
	if len(MailRequestValidatedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_VALIDATED_CC is not set")
	}
	MailRequestValidatedBcc = LoadEmailsFromEnv("MAIL_REQUEST_VALIDATED_BCC")
	if len(MailRequestValidatedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_VALIDATED_BCC is not set")
	}

	MailRequestRejectedTo = LoadEmailsFromEnv("MAIL_REQUEST_REJECTED_TO")
	if len(MailRequestRejectedTo) == 0 {
		log.Fatal("MAIL_REQUEST_REJECTED_TO is required but not set")
	}
	MailRequestRejectedCc = LoadEmailsFromEnv("MAIL_REQUEST_REJECTED_CC")
	if len(MailRequestRejectedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_REJECTED_CC is not set")
	}
	MailRequestRejectedBcc = LoadEmailsFromEnv("MAIL_REQUEST_REJECTED_BCC")
	if len(MailRequestRejectedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_REJECTED_BCC is not set")
	}

	MailRequestApprovedTo = LoadEmailsFromEnv("MAIL_REQUEST_APPROVED_TO")
	if len(MailRequestApprovedTo) == 0 {
		log.Fatal("MAIL_REQUEST_APPROVED_TO is required but not set")
	}
	MailRequestApprovedCc = LoadEmailsFromEnv("MAIL_REQUEST_APPROVED_CC")
	if len(MailRequestApprovedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_APPROVED_CC is not set")
	}
	MailRequestApprovedBcc = LoadEmailsFromEnv("MAIL_REQUEST_APPROVED_BCC")
	if len(MailRequestApprovedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_APPROVED_BCC is not set")
	}

	MailRequestAcceptedTo = LoadEmailsFromEnv("MAIL_REQUEST_ACCEPTED_TO")
	if len(MailRequestAcceptedTo) == 0 {
		log.Fatal("MAIL_REQUEST_ACCEPTED_TO is required but not set")
	}
	MailRequestAcceptedCc = LoadEmailsFromEnv("MAIL_REQUEST_ACCEPTED_CC")
	if len(MailRequestAcceptedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_ACCEPTED_CC is not set")
	}
	MailRequestAcceptedBcc = LoadEmailsFromEnv("MAIL_REQUEST_ACCEPTED_BCC")
	if len(MailRequestAcceptedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_ACCEPTED_BCC is not set")
	}

	MailRequestDeclinedTo = LoadEmailsFromEnv("MAIL_REQUEST_DECLINED_TO")
	if len(MailRequestDeclinedTo) == 0 {
		log.Fatal("MAIL_REQUEST_DECLINED_TO is required but not set")
	}
	MailRequestDeclinedCc = LoadEmailsFromEnv("MAIL_REQUEST_DECLINED_CC")
	if len(MailRequestDeclinedCc) == 0 {
		log.Print("Info: MAIL_REQUEST_DECLINED_CC is not set")
	}
	MailRequestDeclinedBcc = LoadEmailsFromEnv("MAIL_REQUEST_DECLINED_BCC")
	if len(MailRequestDeclinedBcc) == 0 {
		log.Print("Info: MAIL_REQUEST_DECLINED_BCC is not set")
	}

	// --------------------------------------------------------------------------------

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

func LoadEmailsFromEnv(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return nil
	}

	parts := strings.Split(val, ",")
	emails := make([]string, 0, len(parts))

	for _, p := range parts {
		e := strings.TrimSpace(p)
		if e != "" {
			emails = append(emails, e)
		}
	}

	return emails
}

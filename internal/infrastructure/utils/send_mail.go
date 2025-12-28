package utils

import (
	"fmt"
	"os"
	"strconv"

	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"gopkg.in/gomail.v2"

	log "github.com/sirupsen/logrus"
)

// SendEmail sends an email with the provided subject and body to the given recipient
func SendEmail(to, cc, bcc []string, subject string, body string, request model.Request) error {
	// Load sender info from environment variables
	from := configs.MailUsername
	password := configs.MailPassword
	smtpHost := configs.MailServer
	smtpPort, err := strconv.Atoi(configs.MailPort)
	if err != nil {
		log.Printf("Invalid SMTP port: %v\n", err)
		return err
	}

	branchName := "N/A"
	departmentName := "N/A"

	if request.Branch != nil {
		branchName = request.Branch.Name
	}
	if request.Department != nil {
		departmentName = request.Department.Name
	}

	htmlBody := fmt.Sprintf(`
	<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 700px; margin: auto; border: 1px solid #ddd; border-radius: 8px; overflow: hidden;">

		<!-- Header -->
		<div style="background-color: #0693e3; padding: 20px; color: #fff; text-align: center;">
			<div style="display: inline-flex; align-items: center; justify-content: center;">
				<img src="cid:coop_logo" alt="Coop Logo" style="max-height: 33px; margin-right: 10px;" />
				<h1 style="margin: 0; font-size: 2.5em; line-height: 1;">Forex </h1>
			</div>
		</div>

		<!-- Body -->
		<div style="padding: 20px;">
			<p>%s</p>

			<ul style="padding-left: 20px;">
				<li><strong>Request Code:</strong> %s</li>
				<li><strong>Branch:</strong> %s</li>
				<li><strong>Department:</strong> %s</li>
			</ul>

			<p>Please take the necessary actions.</p>
		</div>

		<!-- Footer -->
		<div style="background-color: #f1f1f1; padding: 15px; text-align: center; font-size: 0.85em; color: #666;">
			<p style="margin: 0;"> &copy; 2025 Cooperative Bank of Oromia. All rights reserved.</p>
			<p style="margin: 0;">This is an automated message. Please do not reply.</p>
		</div>
	</div>
	`, body, request.RequestCode, departmentName, branchName)

	// Compose the email
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}

	m.SetHeader("Subject", subject, fmt.Sprintf("<%s@coop-forex>", request.RequestCode))
	m.SetBody("text/html", htmlBody)

	if _, err := os.Stat("assets/coop.gif"); err != nil {
		log.Printf("Image file not found: %v", err)
		return err
	}

	m.Embed("assets/coop.gif", gomail.SetHeader(map[string][]string{
		"Content-ID": {"coop_logo"},
	}))

	// Setup dialer
	d := gomail.NewDialer(smtpHost, smtpPort, from, password)

	s, err := d.Dial()
	if err != nil {
		log.Printf("Failed to dial SMTP server: %v", err)
		return err
	}
	defer s.Close()

	// Send the email
	if err := gomail.Send(s, m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully to", to, cc, bcc)
	return nil
}

// SendEmail sends an email with the provided subject and body to the given recipient
func SendAcknowledgementEmail(to, cc, bcc []string, subject string, body string, request model.Request) error {
	// Load sender info from environment variables
	from := configs.MailUsername
	password := configs.MailPassword
	smtpHost := configs.MailServer
	smtpPort, err := strconv.Atoi(configs.MailPort)
	if err != nil {
		log.Printf("Invalid SMTP port: %v\n", err)
		return err
	}

	branchName := "N/A"
	departmentName := "N/A"

	if request.Branch != nil {
		branchName = request.Branch.Name
	}
	if request.Department != nil {
		departmentName = request.Department.Name
	}

	htmlBody := fmt.Sprintf(`
	<div style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 700px; margin: auto; border: 1px solid #ddd; border-radius: 8px; overflow: hidden;">

		<!-- Header -->
		<div style="background-color: #0693e3; padding: 20px; color: #fff; text-align: center;">
			<div style="display: inline-flex; align-items: center; justify-content: center;">
				<img src="cid:coop_logo" alt="Coop Logo" style="max-height: 33px; margin-right: 10px;" />
				<h1 style="margin: 0; font-size: 2.5em; line-height: 1;">Forex </h1>
			</div>
		</div>

		<!-- Body -->
		<div style="padding: 20px;">
			<p>%s</p>

			<ul style="padding-left: 20px;">
				<li><strong>Request Code:</strong> %s</li>
				<li><strong>Branch:</strong> %s</li>
				<li><strong>Department:</strong> %s</li>
			</ul>

			<p>Please wait for response from relavant team.</p>
		</div>

		<!-- Footer -->
		<div style="background-color: #f1f1f1; padding: 15px; text-align: center; font-size: 0.85em; color: #666;">
			<p style="margin: 0;"> &copy; 2025 Cooperative Bank of Oromia. All rights reserved.</p>
			<p style="margin: 0;">This is an automated message. Please do not reply.</p>
		</div>
	</div>
	`, body, request.RequestCode, departmentName, branchName)

	// Compose the email
	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to...)
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}

	m.SetHeader("Subject", subject, fmt.Sprintf("<%s@coop-forex>", request.RequestCode))
	m.SetBody("text/html", htmlBody)

	if _, err := os.Stat("assets/coop.gif"); err != nil {
		log.Printf("Image file not found: %v", err)
		return err
	}

	m.Embed("assets/coop.gif", gomail.SetHeader(map[string][]string{
		"Content-ID": {"coop_logo"},
	}))

	// Setup dialer
	d := gomail.NewDialer(smtpHost, smtpPort, from, password)

	// Send the email
	s, err := d.Dial()
	if err != nil {
		log.Printf("Failed to dial SMTP server: %v", err)
		return err
	}
	defer s.Close()

	if err := gomail.Send(s, m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully to", to)
	return nil
}

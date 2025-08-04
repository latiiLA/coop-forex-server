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
func SendEmail(to1, to2 string, subject string, body string, request model.Request) error {
	// Load sender info from environment variables
	from := configs.MailSender
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
	m.SetHeader("To", to1)
	m.SetHeader("Cc", to2)
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
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully to", to1, to2)
	return nil
}

// SendEmail sends an email with the provided subject and body to the given recipient
func SendAcknowledgementEmail(to string, subject string, body string, request model.Request) error {
	// Load sender info from environment variables
	from := configs.MailSender
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
	m.SetHeader("To", to)
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
	if err := d.DialAndSend(m); err != nil {
		log.Printf("Failed to send email: %v", err)
		return err
	}

	log.Println("Email sent successfully to", to)
	return nil
}

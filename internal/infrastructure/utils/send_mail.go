package utils

import (
	"crypto/tls"
	"fmt"
	"html"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

const (
	mailFromAddress = "forexhub@coopbankoromiasc.com"
	mailFromName    = "Forex Hub"
)

// RequestMailEvent identifies the lifecycle notification.
type RequestMailEvent string

const (
	MailEventSent       RequestMailEvent = "sent"
	MailEventAuthorized RequestMailEvent = "authorized"
	MailEventValidated  RequestMailEvent = "validated"
	MailEventApproved   RequestMailEvent = "approved"
	MailEventRejected   RequestMailEvent = "rejected"
	MailEventAccepted   RequestMailEvent = "accepted"
	MailEventDeclined   RequestMailEvent = "declined"
	MailEventAssigned   RequestMailEvent = "assigned"
)

type requestMailMeta struct {
	Label   string
	Accent  string
	Subject string
	Cta     string
}

func mailMeta(event RequestMailEvent) requestMailMeta {
	switch event {
	case MailEventSent:
		return requestMailMeta{
			Label: "Submitted", Accent: "#0369a1",
			Subject: "FCY request submitted",
			Cta:     "The request is waiting for branch authorization.",
		}
	case MailEventAuthorized:
		return requestMailMeta{
			Label: "Authorized", Accent: "#0f766e",
			Subject: "FCY request authorized",
			Cta:     "The request is ready for head-office validation.",
		}
	case MailEventValidated:
		return requestMailMeta{
			Label: "Validated", Accent: "#1d4ed8",
			Subject: "FCY request validated",
			Cta:     "The request is ready for approval.",
		}
	case MailEventApproved:
		return requestMailMeta{
			Label: "Approved", Accent: "#15803d",
			Subject: "FCY request approved",
			Cta:     "Branch processing can accept or decline the approved amounts.",
		}
	case MailEventRejected:
		return requestMailMeta{
			Label: "Rejected", Accent: "#b91c1c",
			Subject: "FCY request rejected",
			Cta:     "The request may be edited and resubmitted as a draft.",
		}
	case MailEventAccepted:
		return requestMailMeta{
			Label: "Accepted", Accent: "#047857",
			Subject: "FCY request accepted",
			Cta:     "Processing is complete for this request.",
		}
	case MailEventDeclined:
		return requestMailMeta{
			Label: "Declined", Accent: "#c2410c",
			Subject: "FCY request declined",
			Cta:     "The approved request was declined at processing.",
		}
	case MailEventAssigned:
		return requestMailMeta{
			Label: "Assigned", Accent: "#6d28d9",
			Subject: "FCY request assigned to you",
			Cta:     "Please open Forex Hub and take the next action.",
		}
	default:
		return requestMailMeta{
			Label: "Update", Accent: "#0693e3",
			Subject: "FCY request update",
			Cta:     "Please take the necessary actions in Forex Hub.",
		}
	}
}

// UniqueEmails lowercases, trims, and deduplicates addresses (skips empty).
func UniqueEmails(emails ...string) []string {
	seen := make(map[string]struct{}, len(emails))
	out := make([]string, 0, len(emails))
	for _, raw := range emails {
		email := strings.ToLower(strings.TrimSpace(raw))
		if email == "" || !strings.Contains(email, "@") {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, email)
	}
	return out
}

func appendUnique(dst []string, more ...string) []string {
	return UniqueEmails(append(dst, more...)...)
}

// NotifyRequestActivity sends a branded lifecycle email. No-op when To is empty.
func NotifyRequestActivity(to, cc, bcc []string, event RequestMailEvent, request model.Request, summary string, details []string) error {
	to = UniqueEmails(to...)
	cc = UniqueEmails(cc...)
	bcc = UniqueEmails(bcc...)
	if len(to) == 0 && len(cc) == 0 && len(bcc) == 0 {
		logrus.WithField("event", event).Info("Skipping request email: no recipients")
		return nil
	}
	// gomail requires at least one To; promote Cc/Bcc if needed.
	if len(to) == 0 {
		if len(cc) > 0 {
			to, cc = cc[:1], cc[1:]
		} else {
			to, bcc = bcc[:1], bcc[1:]
		}
	}

	meta := mailMeta(event)
	subject := fmt.Sprintf("%s · %s", meta.Subject, request.RequestCode)
	htmlBody := buildRequestMailHTML(meta, request, summary, details)

	return sendHTMLMail(to, cc, bcc, subject, htmlBody, request.RequestCode)
}

// SendEmail keeps the legacy signature used by older call sites.
func SendEmail(to, cc, bcc []string, subject string, body string, request model.Request) error {
	details := []string{}
	if strings.TrimSpace(body) != "" {
		details = append(details, body)
	}
	return NotifyRequestActivity(to, cc, bcc, "", request, subject, details)
}

// SendAcknowledgementEmail is kept for compatibility; prefers the shared template.
func SendAcknowledgementEmail(to, cc, bcc []string, subject string, body string, request model.Request) error {
	return NotifyRequestActivity(to, cc, bcc, MailEventSent, request, body, nil)
}

func buildRequestMailHTML(meta requestMailMeta, request model.Request, summary string, details []string) string {
	branchName := "N/A"
	departmentName := "N/A"
	if request.Branch != nil && request.Branch.Name != "" {
		branchName = request.Branch.Name
	}
	if request.Department != nil && request.Department.Name != "" {
		departmentName = request.Department.Name
	}

	applicant := strings.TrimSpace(request.ApplicantName)
	if applicant == "" {
		applicant = "—"
	}

	extra := ""
	for _, line := range details {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		extra += fmt.Sprintf(`<li>%s</li>`, html.EscapeString(line))
	}
	if extra != "" {
		extra = `<ul style="padding-left:20px;margin:12px 0;">` + extra + `</ul>`
	}

	summaryHTML := ""
	if strings.TrimSpace(summary) != "" {
		summaryHTML = fmt.Sprintf(`<p style="margin:0 0 16px;">%s</p>`, html.EscapeString(summary))
	}

	year := time.Now().Year()
	return fmt.Sprintf(`
	<div style="font-family:Segoe UI,Arial,sans-serif;line-height:1.55;color:#1f2937;max-width:640px;margin:auto;border:1px solid #e5e7eb;border-radius:10px;overflow:hidden;background:#ffffff;">
		<div style="background:%s;padding:18px 20px;color:#fff;">
			<table width="100%%" cellpadding="0" cellspacing="0"><tr>
				<td style="vertical-align:middle;">
					<img src="cid:coop_logo" alt="Coop" height="28" style="display:block;height:28px;width:auto;"/>
				</td>
				<td style="text-align:right;vertical-align:middle;">
					<span style="display:inline-block;background:rgba(255,255,255,0.18);padding:4px 10px;border-radius:999px;font-size:12px;letter-spacing:0.04em;text-transform:uppercase;">%s</span>
				</td>
			</tr></table>
			<h1 style="margin:14px 0 0;font-size:22px;font-weight:600;">Forex Hub</h1>
		</div>
		<div style="padding:22px 20px;">
			%s
			<table style="width:100%%;border-collapse:collapse;font-size:14px;">
				<tr><td style="padding:8px 0;color:#6b7280;width:38%%;">Request code</td><td style="padding:8px 0;font-weight:600;">%s</td></tr>
				<tr><td style="padding:8px 0;color:#6b7280;">Applicant</td><td style="padding:8px 0;">%s</td></tr>
				<tr><td style="padding:8px 0;color:#6b7280;">Branch</td><td style="padding:8px 0;">%s</td></tr>
				<tr><td style="padding:8px 0;color:#6b7280;">Department</td><td style="padding:8px 0;">%s</td></tr>
				<tr><td style="padding:8px 0;color:#6b7280;">Status</td><td style="padding:8px 0;"><strong>%s</strong></td></tr>
			</table>
			%s
			<p style="margin:16px 0 0;color:#374151;">%s</p>
		</div>
		<div style="background:#f9fafb;padding:14px 20px;text-align:center;font-size:12px;color:#6b7280;border-top:1px solid #e5e7eb;">
			<p style="margin:0;">&copy; %d Cooperative Bank of Oromia. All rights reserved.</p>
			<p style="margin:4px 0 0;">Automated message from Forex Hub — please do not reply.</p>
		</div>
	</div>`,
		meta.Accent,
		html.EscapeString(meta.Label),
		summaryHTML,
		html.EscapeString(request.RequestCode),
		html.EscapeString(applicant),
		html.EscapeString(branchName),
		html.EscapeString(departmentName),
		html.EscapeString(string(request.RequestStatus)),
		extra,
		html.EscapeString(meta.Cta),
		year,
	)
}

func sendHTMLMail(to, cc, bcc []string, subject, htmlBody, requestCode string) error {
	from := configs.MailUsername
	password := configs.MailPassword
	smtpHost := configs.MailServer
	smtpPort, err := strconv.Atoi(configs.MailPort)
	if err != nil {
		logrus.WithError(err).Error("Invalid SMTP port")
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", mailFromName, mailFromAddress))
	m.SetHeader("Reply-To", mailFromAddress)
	m.SetHeader("To", to...)
	if len(cc) > 0 {
		m.SetHeader("Cc", cc...)
	}
	if len(bcc) > 0 {
		m.SetHeader("Bcc", bcc...)
	}
	m.SetHeader("Subject", subject)
	if requestCode != "" {
		m.SetHeader("X-Forex-Request-Code", requestCode)
	}
	m.SetBody("text/html", htmlBody)

	if _, err := os.Stat("assets/coop.gif"); err == nil {
		m.Embed("assets/coop.gif", gomail.SetHeader(map[string][]string{
			"Content-ID": {"coop_logo"},
		}))
	} else {
		logrus.WithError(err).Warn("coop logo missing; sending email without embed")
	}

	d := gomail.NewDialer(smtpHost, smtpPort, from, password)
	d.TLSConfig = &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: configs.MailTLSInsecure, //nolint:gosec // explicit non-prod opt-in via MAIL_TLS_INSECURE
	}

	if err := d.DialAndSend(m); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"to": to, "cc": cc, "subject": subject,
		}).Error("Failed to send request activity email")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"to": to, "cc": cc, "subject": subject,
	}).Info("Request activity email sent")
	return nil
}

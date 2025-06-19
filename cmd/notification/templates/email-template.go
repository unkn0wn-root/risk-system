// Package templates provides email template management and rendering functionality.
// supports both file-based and embedded templates with dynamic data substitution.
package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
)

// EmailTemplate represents an email template with subject and body content.
type EmailTemplate struct {
	Subject  string
	HTMLBody string
	TextBody string
}

// EmailTemplateData contains the dynamic data used to populate email templates.
type EmailTemplateData struct {
	UserID      string
	FirstName   string
	LastName    string
	Email       string
	CompanyName string
	SupportURL  string
	LoginURL    string
	Reason      string
	RiskLevel   string
	Flags       []string
}

// EmailTemplateManager handles email template loading, caching, and rendering.
// supports both file-based templates and embedded fallback templates.
type EmailTemplateManager struct {
	templates map[string]*template.Template
	baseData  EmailTemplateData
}

// NewEmailTemplateManager creates a new template manager with the specified template directory.
func NewEmailTemplateManager(templateDir string) *EmailTemplateManager {
	manager := &EmailTemplateManager{
		templates: make(map[string]*template.Template),
		baseData: EmailTemplateData{
			CompanyName: "User Risk Management System",
			SupportURL:  "https://support.unkn0wnroot.com",
			LoginURL:    "https://app.unkn0wnroot.com/login",
		},
	}

	manager.loadTemplates(templateDir)
	return manager
}

// loadTemplates loads email templates from files or falls back to embedded templates.
func (m *EmailTemplateManager) loadTemplates(templateDir string) {
	templates := map[string]string{
		"welcome":        "welcome.html",
		"risk_alert":     "risk_alert.html",
		"password_reset": "password_reset.html",
		"login_alert":    "login_alert.html",
	}

	for name, filename := range templates {
		path := filepath.Join(templateDir, filename)
		tmpl, err := template.ParseFiles(path)
		if err != nil {
			// Fallback to embedded templates if files not found
			m.templates[name] = m.getEmbeddedTemplate(name)
		} else {
			m.templates[name] = tmpl
		}
	}
}

// RenderTemplate renders an email template with the provided data.
func (m *EmailTemplateManager) RenderTemplate(templateName string, data EmailTemplateData) (string, string, error) {
	data.CompanyName = m.baseData.CompanyName
	data.SupportURL = m.baseData.SupportURL
	data.LoginURL = m.baseData.LoginURL

	tmpl, exists := m.templates[templateName]
	if !exists {
		return "", "", fmt.Errorf("template not found: %s", templateName)
	}

	var htmlBuf bytes.Buffer

	// Render HTML
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		return "", "", fmt.Errorf("failed to render HTML template: %w", err)
	}

	return m.getSubject(templateName, data), htmlBuf.String(), nil
}

// getSubject generates the email subject line based on template type and data.
func (m *EmailTemplateManager) getSubject(templateName string, data EmailTemplateData) string {
	subjects := map[string]string{
		"welcome":        fmt.Sprintf("Welcome to %s, %s!", data.CompanyName, data.FirstName),
		"risk_alert":     fmt.Sprintf("🚨 Security Alert - %s Risk Detected", data.RiskLevel),
		"password_reset": "Password Reset Request",
		"login_alert":    "🔐 New Login to Your Account",
	}

	if subject, exists := subjects[templateName]; exists {
		return subject
	}
	return "Notification from " + data.CompanyName
}

// getEmbeddedTemplate returns hardcoded HTML templates as fallbacks.
// These are used when template files are not available in the filesystem.
func (m *EmailTemplateManager) getEmbeddedTemplate(name string) *template.Template {
	templates := map[string]string{
		"welcome": `
<!DOCTYPE html>
<html>
<head><title>Welcome</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
	<div style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 20px; text-align: center;">
		<h1 style="color: white; margin: 0;">Welcome to {{.CompanyName}}!</h1>
	</div>
	<div style="padding: 30px;">
		<h2>Hi {{.FirstName}},</h2>
		<p>Welcome aboard! We're excited to have you as part of our community.</p>
		<p>Your account has been successfully created with the email: <strong>{{.Email}}</strong></p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="{{.LoginURL}}" style="background: #667eea; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Get Started</a>
		</div>
		<p>If you have any questions, feel free to contact our support team.</p>
		<p>Best regards,<br>The {{.CompanyName}} Team</p>
	</div>
	<div style="background: #f8f9fa; padding: 20px; text-align: center; font-size: 12px; color: #666;">
		<p>Need help? Visit our <a href="{{.SupportURL}}">Support Center</a></p>
	</div>
</body>
</html>`,

		"risk_alert": `
<!DOCTYPE html>
<html>
<head><title>Security Alert</title></head>
<body style="font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto;">
	<div style="background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%); padding: 20px; text-align: center;">
		<h1 style="color: white; margin: 0;">🚨 Security Alert</h1>
	</div>
	<div style="padding: 30px;">
		<h2>Security Risk Detected</h2>
		<div style="background: #fff3cd; border: 1px solid #ffeaa7; padding: 15px; border-radius: 5px; margin: 20px 0;">
			<p><strong>Risk Level:</strong> {{.RiskLevel}}</p>
			<p><strong>Reason:</strong> {{.Reason}}</p>
			{{if .Flags}}<p><strong>Flags:</strong> {{range .Flags}}{{.}} {{end}}</p>{{end}}
		</div>
		<p>We've detected suspicious activity associated with your account. Please review your account immediately.</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="{{.LoginURL}}" style="background: #ee5a24; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Review Account</a>
		</div>
		<p>If you believe this is an error, please contact our security team immediately.</p>
		<p>Security Team<br>{{.CompanyName}}</p>
	</div>
</body>
</html>`,
	}

	if tmplStr, exists := templates[name]; exists {
		tmpl, _ := template.New(name).Parse(tmplStr)
		return tmpl
	}
	return template.New(name)
}

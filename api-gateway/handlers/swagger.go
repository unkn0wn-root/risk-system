package handlers

import (
	"encoding/json"
	"net/http"
)

type SwaggerHandler struct{}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

type OpenAPISpec struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers"`
	Paths      map[string]interface{} `json:"paths"`
	Components OpenAPIComponents      `json:"components"`
}

type OpenAPIInfo struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Contact     OpenAPIContact `json:"contact"`
}

type OpenAPIContact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type OpenAPIServer struct {
	URL         string `json:"url"`
	Description string `json:"description"`
}

type OpenAPIComponents struct {
	SecuritySchemes map[string]OpenAPISecurityScheme `json:"securitySchemes"`
	Schemas         map[string]interface{}           `json:"schemas"`
}

type OpenAPISecurityScheme struct {
	Type         string `json:"type"`
	Scheme       string `json:"scheme"`
	BearerFormat string `json:"bearerFormat,omitempty"`
	Description  string `json:"description"`
}

func (h *SwaggerHandler) GetOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := OpenAPISpec{
		OpenAPI: "3.0.3",
		Info: OpenAPIInfo{
			Title:       "User Risk Management System API",
			Description: "An API for user risk assessment and management",
			Version:     "2.0.0",
			Contact: OpenAPIContact{
				Name:  "Risk Management Super Team",
				Email: "support@mysupperfakecompany.com",
			},
		},
		Servers: []OpenAPIServer{
			{
				URL:         "/api/v1",
				Description: "Production API Server",
			},
		},
		Paths: map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Health"},
					"summary":     "Health check endpoint",
					"description": "Returns the health status of the API service",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Service is healthy",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{
												"type":    "string",
												"example": "healthy",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"/auth/login": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "User login",
					"description": "Authenticate user and return JWT token",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"email", "password"},
									"properties": map[string]interface{}{
										"email": map[string]interface{}{
											"type":    "string",
											"format":  "email",
											"example": "user@example.com",
										},
										"password": map[string]interface{}{
											"type":    "string",
											"example": "password123",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Successful authentication",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/AuthResponse",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Invalid credentials",
						},
					},
				},
			},
			"/auth/register": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "User registration",
					"description": "Register a new user account",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/UserRegistration",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "User created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Invalid input data",
						},
					},
				},
			},
			"/auth/refresh": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Authentication"},
					"summary":     "Refresh JWT token",
					"description": "Refresh an expired JWT token",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":     "object",
									"required": []string{"refresh_token"},
									"properties": map[string]interface{}{
										"refresh_token": map[string]interface{}{
											"type": "string",
										},
									},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Token refreshed successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/AuthResponse",
									},
								},
							},
						},
					},
				},
			},
			"/profile": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"User Profile"},
					"summary":     "Get user profile",
					"description": "Get the authenticated user's profile information",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User profile retrieved successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
						"401": map[string]interface{}{
							"description": "Unauthorized",
						},
					},
				},
			},
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"User Management"},
					"summary":     "List users (Admin only)",
					"description": "Get a list of all users - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of users",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/User",
										},
									},
								},
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{"User Management"},
					"summary":     "Create user (Admin only)",
					"description": "Create a new user - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/UserRegistration",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "User created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
			},
			"/users/{id}": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"User Management"},
					"summary":     "Get user by ID",
					"description": "Get user details by ID - users can access their own data, admins can access any",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "User ID",
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User details",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "User not found",
						},
					},
				},
				"put": map[string]interface{}{
					"tags":        []string{"User Management"},
					"summary":     "Update user",
					"description": "Update user information - users can update their own data, admins can update any",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "User ID",
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/UserUpdate",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "User updated successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/User",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "User not found",
						},
					},
				},
			},
			"/risk/check": map[string]interface{}{
				"post": map[string]interface{}{
					"tags":        []string{"Risk Assessment"},
					"summary":     "Check user risk",
					"description": "Perform risk assessment for authenticated user",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/RiskCheckRequest",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Risk assessment completed",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RiskCheckResponse",
									},
								},
							},
						},
					},
				},
			},
			"/risk/rules": map[string]interface{}{
				"get": map[string]interface{}{
					"tags":        []string{"Risk Management"},
					"summary":     "List risk rules (Admin only)",
					"description": "Get all risk rules - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of risk rules",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "array",
										"items": map[string]interface{}{
											"$ref": "#/components/schemas/RiskRule",
										},
									},
								},
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
				"post": map[string]interface{}{
					"tags":        []string{"Risk Management"},
					"summary":     "Create risk rule (Admin only)",
					"description": "Create a new risk rule - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/RiskRuleCreate",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"201": map[string]interface{}{
							"description": "Risk rule created successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RiskRule",
									},
								},
							},
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
			},
			"/risk/rules/{id}": map[string]interface{}{
				"put": map[string]interface{}{
					"tags":        []string{"Risk Management"},
					"summary":     "Update risk rule (Admin only)",
					"description": "Update an existing risk rule - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Risk rule ID",
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/RiskRuleUpdate",
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Risk rule updated successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"$ref": "#/components/schemas/RiskRule",
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Risk rule not found",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
				"delete": map[string]interface{}{
					"tags":        []string{"Risk Management"},
					"summary":     "Delete risk rule (Admin only)",
					"description": "Delete a risk rule - requires admin role",
					"security": []map[string]interface{}{
						{"bearerAuth": []string{}},
					},
					"parameters": []map[string]interface{}{
						{
							"name":        "id",
							"in":          "path",
							"required":    true,
							"description": "Risk rule ID",
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"204": map[string]interface{}{
							"description": "Risk rule deleted successfully",
						},
						"404": map[string]interface{}{
							"description": "Risk rule not found",
						},
						"403": map[string]interface{}{
							"description": "Forbidden - Admin role required",
						},
					},
				},
			},
		},
		Components: OpenAPIComponents{
			SecuritySchemes: map[string]OpenAPISecurityScheme{
				"bearerAuth": {
					Type:         "http",
					Scheme:       "bearer",
					BearerFormat: "JWT",
					Description:  "JWT Authorization header using the Bearer scheme",
				},
			},
			Schemas: map[string]interface{}{
				"User": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "string",
						},
						"email": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
						"first_name": map[string]interface{}{
							"type": "string",
						},
						"last_name": map[string]interface{}{
							"type": "string",
						},
						"role": map[string]interface{}{
							"type": "string",
							"enum": []string{"user", "admin"},
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
					},
				},
				"UserRegistration": map[string]interface{}{
					"type":     "object",
					"required": []string{"email", "password", "first_name", "last_name"},
					"properties": map[string]interface{}{
						"email": map[string]interface{}{
							"type":   "string",
							"format": "email",
						},
						"password": map[string]interface{}{
							"type":      "string",
							"minLength": 8,
						},
						"first_name": map[string]interface{}{
							"type": "string",
						},
						"last_name": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"UserUpdate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"first_name": map[string]interface{}{
							"type": "string",
						},
						"last_name": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"AuthResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"access_token": map[string]interface{}{
							"type": "string",
						},
						"refresh_token": map[string]interface{}{
							"type": "string",
						},
						"expires_in": map[string]interface{}{
							"type": "integer",
						},
						"user": map[string]interface{}{
							"$ref": "#/components/schemas/User",
						},
					},
				},
				"RiskCheckRequest": map[string]interface{}{
					"type":     "object",
					"required": []string{"transaction_amount", "transaction_type"},
					"properties": map[string]interface{}{
						"transaction_amount": map[string]interface{}{
							"type": "number",
						},
						"transaction_type": map[string]interface{}{
							"type": "string",
						},
						"merchant_category": map[string]interface{}{
							"type": "string",
						},
						"location": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"RiskCheckResponse": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"risk_score": map[string]interface{}{
							"type": "number",
						},
						"risk_level": map[string]interface{}{
							"type": "string",
							"enum": []string{"low", "medium", "high", "critical"},
						},
						"decision": map[string]interface{}{
							"type": "string",
							"enum": []string{"approve", "review", "decline"},
						},
						"reasons": map[string]interface{}{
							"type": "array",
							"items": map[string]interface{}{
								"type": "string",
							},
						},
					},
				},
				"RiskRule": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id": map[string]interface{}{
							"type": "string",
						},
						"name": map[string]interface{}{
							"type": "string",
						},
						"description": map[string]interface{}{
							"type": "string",
						},
						"rule_type": map[string]interface{}{
							"type": "string",
						},
						"threshold": map[string]interface{}{
							"type": "number",
						},
						"action": map[string]interface{}{
							"type": "string",
						},
						"is_active": map[string]interface{}{
							"type": "boolean",
						},
						"created_at": map[string]interface{}{
							"type":   "string",
							"format": "date-time",
						},
					},
				},
				"RiskRuleCreate": map[string]interface{}{
					"type":     "object",
					"required": []string{"name", "rule_type", "threshold", "action"},
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"description": map[string]interface{}{
							"type": "string",
						},
						"rule_type": map[string]interface{}{
							"type": "string",
						},
						"threshold": map[string]interface{}{
							"type": "number",
						},
						"action": map[string]interface{}{
							"type": "string",
						},
					},
				},
				"RiskRuleUpdate": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type": "string",
						},
						"description": map[string]interface{}{
							"type": "string",
						},
						"threshold": map[string]interface{}{
							"type": "number",
						},
						"action": map[string]interface{}{
							"type": "string",
						},
						"is_active": map[string]interface{}{
							"type": "boolean",
						},
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spec)
}

func (h *SwaggerHandler) GetSwaggerUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>User Risk Management System API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/docs/openapi.json',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                persistAuthorization: true,
                displayRequestDuration: true,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
        };
    </script>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

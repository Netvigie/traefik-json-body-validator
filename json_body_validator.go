package traefik_json_body_validator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

// Config holds the plugin configuration
type Config struct {
	Rules []ValidationRule `json:"rules,omitempty"`
	Response ErrorResponse `json:"response,omitempty"`
}

// ValidationRule defines a single validation rule
type ValidationRule struct {
	Field    string `json:"field"`
	Pattern  string `json:"pattern,omitempty"`
	Required bool   `json:"required"`
	MinLength int   `json:"minLength,omitempty"`
	MaxLength int   `json:"maxLength,omitempty"`
}

// ErrorResponse defines the error response structure
type ErrorResponse struct {
	Status  int    `json:"status,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreateConfig creates the default plugin configuration
func CreateConfig() *Config {
	return &Config{
		Response: ErrorResponse{
			Status:  400,
			Code:    "INVALID_REQUEST",
			Message: "Invalid request body",
		},
	}
}

// JSONBodyValidator is the plugin struct
type JSONBodyValidator struct {
	next   http.Handler
	config *Config
	name   string
	rules  map[string]*compiledRule
}

type compiledRule struct {
	rule    ValidationRule
	regex   *regexp.Regexp
}

// New creates a new JSONBodyValidator plugin
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.Rules) == 0 {
		return nil, fmt.Errorf("at least one validation rule is required")
	}

	validator := &JSONBodyValidator{
		next:   next,
		config: config,
		name:   name,
		rules:  make(map[string]*compiledRule),
	}

	// Compile regex patterns
	for _, rule := range config.Rules {
		cr := &compiledRule{rule: rule}
		
		if rule.Pattern != "" {
			regex, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regex pattern for field %s: %w", rule.Field, err)
			}
			cr.regex = regex
		}
		
		validator.rules[rule.Field] = cr
	}

	return validator, nil
}

func (v *JSONBodyValidator) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Only process requests with body
	if req.Body == nil {
		v.sendError(rw, "Request body is required")
		return
	}

	// Read body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		v.sendError(rw, "Failed to read request body")
		return
	}
	req.Body.Close()

	// Parse JSON
	var jsonBody map[string]interface{}
	if err := json.Unmarshal(body, &jsonBody); err != nil {
		v.sendError(rw, "Invalid JSON format")
		return
	}

	// Validate rules
	for fieldName, compiledRule := range v.rules {
		value, exists := jsonBody[fieldName]

		// Check if field is required
		if compiledRule.rule.Required && !exists {
			v.sendError(rw, fmt.Sprintf("Field '%s' is required", fieldName))
			return
		}

		if !exists {
			continue
		}

		// Convert value to string
		strValue := fmt.Sprintf("%v", value)

		// Check empty value for required fields
		if compiledRule.rule.Required && strValue == "" {
			v.sendError(rw, fmt.Sprintf("Field '%s' cannot be empty", fieldName))
			return
		}

		// Check min length
		if compiledRule.rule.MinLength > 0 && len(strValue) < compiledRule.rule.MinLength {
			v.sendError(rw, fmt.Sprintf("Field '%s' must be at least %d characters", fieldName, compiledRule.rule.MinLength))
			return
		}

		// Check max length
		if compiledRule.rule.MaxLength > 0 && len(strValue) > compiledRule.rule.MaxLength {
			v.sendError(rw, fmt.Sprintf("Field '%s' must not exceed %d characters", fieldName, compiledRule.rule.MaxLength))
			return
		}

		// Check regex pattern
		if compiledRule.regex != nil && !compiledRule.regex.MatchString(strValue) {
			v.sendError(rw, fmt.Sprintf("Field '%s' does not match required pattern", fieldName))
			return
		}
	}

	// Restore body for next handler
	req.Body = io.NopCloser(bytes.NewBuffer(body))

	// All validations passed
	v.next.ServeHTTP(rw, req)
}

func (v *JSONBodyValidator) sendError(rw http.ResponseWriter, message string) {
	status := v.config.Response.Status
	if status == 0 {
		status = 400
	}

	response := map[string]interface{}{
		"error": message,
	}

	if v.config.Response.Code != "" {
		response["code"] = v.config.Response.Code
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(status)
	json.NewEncoder(rw).Encode(response)
}

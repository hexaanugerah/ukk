package gallery

import (
	"net/url"
	"strings"

	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func init() {
	validate.RegisterValidation("file", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		return val != "" && len(val) <= 500
	})
}

func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
	}
	return out
}

func ValidateFileRequest(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: fe.Field() + " is invalid"})
	}
	return out
}

func isHTTPImageURL(value string) bool {
	parsed, err := url.ParseRequestURI(strings.TrimSpace(value))
	return err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https")
}

func isValidImageFileRef(value string) bool {
	if value == "" || value == "." || value == ".." || strings.TrimSpace(value) != value || len(value) > 200 {
		return false
	}
	for _, ch := range value {
		switch ch {
		case '/', '\\', ':', '|', '?', '*':
			return false
		}
	}
	return true
}

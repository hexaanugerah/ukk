package article

import (
	"github.com/KawahPutih/backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) []response.FieldError {
	err := validate.Struct(s)
	if err == nil {
		return nil
	}
	var out []response.FieldError
	for _, fe := range err.(validator.ValidationErrors) {
		out = append(out, response.FieldError{Field: fe.Field(), Message: friendlyMessage(fe)})
	}
	return out
}

func friendlyMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "url":
		return fe.Field() + " must be a valid URL"
	default:
		return fe.Field() + " is invalid"
	}
}

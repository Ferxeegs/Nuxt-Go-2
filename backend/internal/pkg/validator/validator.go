// Package validator translates Gin/go-playground validation errors into the
// standardized field->message map, and sanitizes rich-text HTML input to
// prevent stored XSS.
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/microcosm-cc/bluemonday"
)

var sanitizer = bluemonday.UGCPolicy()

// SanitizeHTML strips dangerous tags/attributes from user-supplied HTML
// (e.g. article content) while keeping common formatting tags allowed by
// bluemonday's UGC policy.
func SanitizeHTML(input string) string {
	return sanitizer.Sanitize(input)
}

// TranslateErrors converts validator.ValidationErrors into a flat
// field -> human-readable message map used by response.Error.
func TranslateErrors(err error) map[string]string {
	out := map[string]string{}

	verrs, ok := err.(validator.ValidationErrors)
	if !ok {
		out["error"] = err.Error()
		return out
	}

	for _, fe := range verrs {
		field := strings.ToLower(fe.Field())
		out[field] = message(fe)
	}
	return out
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters/value", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters/value", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}

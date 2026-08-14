package response

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"codebasego/internal/common"
)

// Body is the standard API response envelope.
type Body struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Body{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Body{Success: true, Data: data})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Body{Success: false, Error: message})
}

func SuccessWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Body{Success: true, Data: data, Meta: meta})
}

func InternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, Body{Success: false, Error: "internal server error"})
}

// HandleError automatically detects *common.AppError types and writes the corresponding HTTP response.
func HandleError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	var appErr *common.AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.Code, Body{Success: false, Error: appErr.Message})
		return
	}

	c.JSON(http.StatusInternalServerError, Body{Success: false, Error: "internal server error"})
}

// ValidationError formats binding/validation errors into human-readable field messages.
func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string, len(ve))
		for _, fe := range ve {
			details[toSnakeCase(fe.Field())] = formatValidationMsg(fe)
		}
		c.JSON(http.StatusBadRequest, Body{
			Success: false,
			Error:   "validation failed",
			Meta:    details,
		})
		return
	}

	c.JSON(http.StatusBadRequest, Body{Success: false, Error: err.Error()})
}

func formatValidationMsg(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	default:
		return fe.Error()
	}
}

func toSnakeCase(str string) string {
	var b strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}


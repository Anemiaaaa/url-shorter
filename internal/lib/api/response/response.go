package response

import (
	"fmt"
	"github.com/go-playground/validator"
	"strings"
)

type Response struct {
	Status string `json:"status"` // "ok", "error"
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK    = "ok"
	StatusError = "error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field URL is a required field"))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field URL is not a valid URL"))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("Field '%s' is invalid: %s", err.Field(), err.ActualTag()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}

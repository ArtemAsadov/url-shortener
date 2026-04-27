package response

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOk  = "Ok"
	StatusErr = "Error"
)

func Ok() Response {
	return Response{
		Status: StatusOk,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field"), err.Field())
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not a valid URL"), err.Field())
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid"), err.Field())
		}
	}

	return Error(strings.Join(errMsgs, ", "))
}

func Error(msg string) Response {
	return Response{
		Status: StatusErr,
		Error:  msg,
	}
}

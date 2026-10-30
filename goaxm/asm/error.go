package asm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/micromdm/nanoaxm/client"
)

// ASMErrorResponseError is an error wraps an ASM ErrorResponse JSON struct.
type ASMErrorResponseError struct {
	*ErrorResponse
}

// NewASMErrorResponseError creates the error from the ASM ErrorResponse JSON struct.
func NewASMErrorResponseError(errorResponse *ErrorResponse) *ASMErrorResponseError {
	return &ASMErrorResponseError{ErrorResponse: errorResponse}
}

// NewASMErrorResponseErrorFromReader decodes the ASM ErrorResponse JSON from jsonReader.
// A new [ASMErrorResponseError] is returned or an error decoding the JSON.
func NewASMErrorResponseErrorFromReader(jsonReader io.Reader) error {
	if jsonReader == nil {
		return errors.New("nil reader")
	}
	asmErr := new(ErrorResponse)
	err := json.NewDecoder(jsonReader).Decode(asmErr)
	if err != nil {
		return fmt.Errorf("decoding ASM error response: %w", err)
	}
	return NewASMErrorResponseError(asmErr)
}

// Error generates an error string based on the embedded ErrorResponse JSON struct.
func (e *ASMErrorResponseError) Error() string {
	if e == nil {
		return "nil *ASMErrorResponseError"
	}
	if e.ErrorResponse == nil {
		return "nil *ErrorResponse"
	}
	if len(e.Errors) < 1 {
		return "empty ASM errors"
	}

	var errStrs []string
	for i, e := range e.Errors {
		errStr := fmt.Sprintf("error %d: %s: %s: %s", i+1, e.Detail, e.Code, e.Status)
		if e.ID != nil {
			errStr += ": " + *e.ID
		}
		errStrs = append(errStrs, errStr)
	}

	return "ASM error response: " + strings.Join(errStrs, ", ")
}

// defaultErrStatuses is the default set of HTTP statuses that will generate HTTP errors.
var defaultErrStatuses = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusTooManyRequests,
}

// ASMError searches errStatuses for the response code in r and returns a
// [ASMErrorResponseError] if found, otherwise an HTTP error.
func ASMError(r *http.Response, errStatuses []int) error {
	if errStatuses == nil {
		errStatuses = defaultErrStatuses
	}
	foundErrStatus := false
	for _, e := range errStatuses {
		if r.StatusCode == e {
			foundErrStatus = true
			break
		}
	}
	if foundErrStatus {
		return NewASMErrorResponseErrorFromReader(r.Body)
	}
	return client.NewHTTPError(r)
}

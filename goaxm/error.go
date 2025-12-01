package goaxm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/micromdm/nanoaxm/goaxm/abm"
)

// ABMErrorResponseError is an error wraps an ABM ErrorResponse JSON struct.
// See https://developer.apple.com/documentation/applebusinessmanagerapi/errorresponse
type ABMErrorResponseError struct {
	*abm.ErrorResponseJson
}

// NewABMErrorResponseError creates the error from the ABM ErrorResponse JSON struct.
func NewABMErrorResponseError(errorResponse *abm.ErrorResponseJson) *ABMErrorResponseError {
	return &ABMErrorResponseError{ErrorResponseJson: errorResponse}
}

// NewABMErrorResponseErrorFromReader decodes the ABM ErrorResponse JSON from jsonReader.
// A new [ABMErrorResponseError] is returned or an error decoding the JSON.
func NewABMErrorResponseErrorFromReader(jsonReader io.Reader) error {
	if jsonReader == nil {
		return errors.New("nil reader")
	}
	abmErr := new(abm.ErrorResponseJson)
	err := json.NewDecoder(jsonReader).Decode(abmErr)
	if err != nil {
		return fmt.Errorf("decoding ABM error response: %w", err)
	}
	return NewABMErrorResponseError(abmErr)
}

// Error generates an error string based on the embedded ErrorResponse JSON struct.
func (e *ABMErrorResponseError) Error() string {
	if e == nil {
		return "nil *ABMErrorResponseError"
	}
	if e.ErrorResponseJson == nil {
		return "nil *ErrorResponseJson"
	}
	if len(e.Errors) < 1 {
		return "empty ABM errors"
	}

	var errStrs []string
	for i, e := range e.Errors {
		errStr := fmt.Sprintf("error %d: %s: %s: %s", i+1, e.Detail, e.Code, e.Status)
		if e.Id != nil {
			errStr += ": " + *e.Id
		}
		errStrs = append(errStrs, errStr)
	}

	return "ABM error response: " + strings.Join(errStrs, ", ")
}

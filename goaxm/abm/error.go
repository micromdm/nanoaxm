package abm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/micromdm/nanoaxm/client"
)

// ABMErrorResponseError is an error wraps an ABM ErrorResponse JSON struct.
// See https://developer.apple.com/documentation/applebusinessmanagerapi/errorresponse
type ABMErrorResponseError struct {
	*ErrorResponse
}

// NewABMErrorResponseError creates the error from the ABM ErrorResponse JSON struct.
func NewABMErrorResponseError(errorResponse *ErrorResponse) *ABMErrorResponseError {
	return &ABMErrorResponseError{ErrorResponse: errorResponse}
}

// NewABMErrorResponseErrorFromReader decodes the ABM ErrorResponse JSON from jsonReader.
// A new [ABMErrorResponseError] is returned or an error decoding the JSON.
func NewABMErrorResponseErrorFromReader(jsonReader io.Reader) error {
	if jsonReader == nil {
		return errors.New("nil reader")
	}
	abmErr := new(ErrorResponse)
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
	if e.ErrorResponse == nil {
		return "nil *ErrorResponse"
	}
	if len(e.Errors) < 1 {
		return "empty ABM errors"
	}

	var errStrs []string
	for i, e := range e.Errors {
		errStr := fmt.Sprintf("error %d: %s: %s: %s", i+1, e.Detail, e.Code, e.Status)
		if e.ID != nil {
			errStr += ": " + *e.ID
		}
		errStrs = append(errStrs, errStr)
	}

	return "ABM error response: " + strings.Join(errStrs, ", ")
}

// defaultErrStatuses is the default set of HTTP statuses that will generate HTTP errors.
var defaultErrStatuses = []int{
	http.StatusBadRequest,
	http.StatusUnauthorized,
	http.StatusForbidden,
	http.StatusTooManyRequests,
}

// ABMError searches errStatuses for the response code in r and returns a
// [ABMErrorResponseError] if found, otherwise an HTTP error.
func ABMError(r *http.Response, errStatuses []int) error {
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
		return NewABMErrorResponseErrorFromReader(r.Body)
	}
	return client.NewHTTPError(r)
}

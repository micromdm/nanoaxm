package goaxm

import (
	"fmt"
	"strings"

	"github.com/micromdm/nanoaxm/goaxm/abm"
)

type ABMErrorResponseError struct {
	abm.ErrorResponseJson
}

func (e *ABMErrorResponseError) Error() string {
	if e == nil {
		return "nil *ABMErrorResponseError"
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

package httputil

import "fmt"

type HTTPError struct {
	Status int
	Err    error
}

func (err *HTTPError) Error() string {
	if err.Err != nil {
		return fmt.Sprintf("status %d, reason %s", err.Status, err.Err.Error())
	}
	return fmt.Sprintf("Status %d", err.Status)
}

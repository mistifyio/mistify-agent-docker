package mdocker

import "fmt"

type (
	// ErrorHTTPCode should be used for errors resulting from an http response
	// code not matching the expected code
	ErrorHTTPCode struct {
		Expected int
		Code     int
		Source   string
	}
)

// Error returns a string error message
func (e ErrorHTTPCode) Error() string {
	return fmt.Sprintf("unexpected http response code: expected %d, received %d, url: %s", e.Expected, e.Code, e.Source)
}

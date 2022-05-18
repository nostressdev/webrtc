package webrtc

import "fmt"

const (
	ErrInvalidState = "InvalidStateError"
)

func makeError(code string, message string) (err error) {
	return fmt.Errorf("%s: %s", code, message)
}

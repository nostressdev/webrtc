package webrtc

import "fmt"

const (
	ErrInvalidState        = "InvalidStateError"
	ErrInvalidModification = "InvalidModificationError"
)

func makeError(code string, message string) (err error) {
	return fmt.Errorf("%s: %s", code, message)
}

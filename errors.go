package emitwave

import "fmt"

// EmitwaveError is returned when the gateway responds with a non-2xx status.
// It carries only the status and the gateway's error body — never the secret.
type EmitwaveError struct {
	Status int
	Body   string
}

func (e *EmitwaveError) Error() string {
	return fmt.Sprintf("emitwave publish failed (HTTP %d): %s", e.Status, e.Body)
}

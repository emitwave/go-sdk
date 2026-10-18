package emitwave

import (
	"strings"
	"testing"
)

func TestEmitwaveErrorFormatsStatusAndBodyWithoutSecret(t *testing.T) {
	err := &EmitwaveError{Status: 403, Body: `{"error":{"code":"INVALID_APP_KEY"}}`}
	msg := err.Error()
	if !strings.Contains(msg, "403") {
		t.Fatalf("want status in message, got %q", msg)
	}
	if !strings.Contains(msg, "INVALID_APP_KEY") {
		t.Fatalf("want body in message, got %q", msg)
	}
	if strings.Contains(msg, "sekret") {
		t.Fatalf("secret must never appear in the error: %q", msg)
	}
}

package emitwave

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/emitwave/protocol"
)

func TestSigningVectorMatchesSharedFixture(t *testing.T) {
	// The shared cross-language vector is vendored into testdata/ so this passes
	// in the standalone emitwave/go-sdk repo; falls back to the monorepo path.
	raw, err := os.ReadFile("testdata/gateway-signing.json")
	if err != nil {
		raw, err = os.ReadFile("../protocol/fixtures/gateway-signing.json")
	}
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var v struct {
		Method, Path, Secret, Signature string
		Query                           map[string]string
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	canonical := protocol.CanonicalString(v.Method, v.Path, protocol.CanonicalQuery(v.Query), v.Query["body_sha256"])
	if got := protocol.Sign(canonical, v.Secret); got != v.Signature {
		t.Fatalf("signature drift: got %s want %s", got, v.Signature)
	}
}

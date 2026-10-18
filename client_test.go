package emitwave

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/emitwave/protocol"
)

func TestTriggerSignsAndPostsCorrectRequest(t *testing.T) {
	var gotMethod, gotPath, gotBody string
	var gotQuery map[string]string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		gotQuery = map[string]string{}
		for k, vs := range r.URL.Query() {
			gotQuery[k] = vs[0]
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"published":true}`))
	}))
	defer srv.Close()

	c, err := New(Config{GatewayURL: srv.URL, AppID: "app_1", Key: "ew_pk_x", Secret: "sekret"})
	if err != nil {
		t.Fatal(err)
	}
	if err := c.Trigger(context.Background(), "news", "article.created", map[string]any{"hello": "world"}); err != nil {
		t.Fatal(err)
	}

	if gotMethod != "POST" {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotPath != "/apps/app_1/events" {
		t.Fatalf("path = %q", gotPath)
	}
	if !strings.Contains(gotBody, `"name":"article.created"`) || !strings.Contains(gotBody, `"channel":"news"`) {
		t.Fatalf("body = %q", gotBody)
	}
	if gotQuery["auth_key"] != "ew_pk_x" || gotQuery["auth_version"] != "1.0" || gotQuery["auth_timestamp"] == "" {
		t.Fatalf("query = %v", gotQuery)
	}
	if gotQuery["body_sha256"] != protocol.BodySHA256([]byte(gotBody)) {
		t.Fatalf("body_sha256 mismatch")
	}
	expected := protocol.Sign(
		protocol.CanonicalString("POST", "/apps/app_1/events", protocol.CanonicalQuery(gotQuery), gotQuery["body_sha256"]),
		"sekret",
	)
	if gotQuery["auth_signature"] != expected {
		t.Fatalf("signature mismatch")
	}
}

func TestTriggerNon2xxReturnsEmitwaveErrorWithoutSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error":{"code":"INVALID_APP_KEY"}}`))
	}))
	defer srv.Close()

	c, _ := New(Config{GatewayURL: srv.URL, AppID: "app_1", Key: "k", Secret: "sekret"})
	err := c.Trigger(context.Background(), "news", "x", nil)
	var ewErr *EmitwaveError
	if err == nil || !errors.As(err, &ewErr) {
		t.Fatalf("want *EmitwaveError, got %v", err)
	}
	if ewErr.Status != 403 {
		t.Fatalf("status = %d", ewErr.Status)
	}
	if !strings.Contains(ewErr.Error(), "INVALID_APP_KEY") {
		t.Fatalf("want body in error")
	}
	if strings.Contains(ewErr.Error(), "sekret") {
		t.Fatalf("secret leaked into error")
	}
}

func TestNewRejectsMissingConfig(t *testing.T) {
	if _, err := New(Config{GatewayURL: "http://x", AppID: "", Key: "k", Secret: "s"}); err == nil {
		t.Fatal("want error for empty AppID")
	}
}

func TestTriggerSurfacesTransportError(t *testing.T) {
	// Point at a closed port; the request must return an error, not nil.
	c, _ := New(Config{GatewayURL: "http://127.0.0.1:1", AppID: "a", Key: "k", Secret: "s"})
	if err := c.Trigger(context.Background(), "news", "x", nil); err == nil {
		t.Fatal("want transport error, got nil")
	}
}

func TestTriggerUnmarshalableDataReturnsError(t *testing.T) {
	c, _ := New(Config{GatewayURL: "http://x", AppID: "a", Key: "k", Secret: "s"})
	// A channel cannot be JSON-marshaled; must error before any HTTP call.
	if err := c.Trigger(context.Background(), "news", "x", make(chan int)); err == nil {
		t.Fatal("want marshal error, got nil")
	}
}

func TestTriggerNilDataSendsEmptyObjectNotNull(t *testing.T) {
	// Cross-SDK consistency: node-sdk defaults missing data to {}. go-sdk must
	// send "data":{} for nil, not "data":null, so a subscriber sees the same
	// payload shape regardless of which server SDK published.
	var gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.WriteHeader(200)
	}))
	defer srv.Close()

	c, _ := New(Config{GatewayURL: srv.URL, AppID: "a", Key: "k", Secret: "s"})
	if err := c.Trigger(context.Background(), "news", "x", nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotBody, `"data":{}`) {
		t.Fatalf(`want "data":{} in body, got %q`, gotBody)
	}
}

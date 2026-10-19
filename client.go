// Package emitwave is a server-side publish SDK for the Emitwave gateway.
package emitwave

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emitwave/protocol"
)

// DefaultGatewayURL is the live Emitwave gateway, used when Config.GatewayURL
// is empty.
const DefaultGatewayURL = "https://api.emitwave.com"

type Config struct {
	GatewayURL string // optional; defaults to DefaultGatewayURL
	AppID      string
	Key        string // public key -> auth_key
	Secret     string
	HTTPClient *http.Client // optional; defaults to a 10s-timeout client
}

type Client struct {
	cfg  Config
	http *http.Client
}

func New(cfg Config) (*Client, error) {
	if cfg.GatewayURL == "" {
		cfg.GatewayURL = DefaultGatewayURL
	}
	for name, v := range map[string]string{
		"AppID":  cfg.AppID,
		"Key":    cfg.Key,
		"Secret": cfg.Secret,
	} {
		if v == "" {
			return nil, fmt.Errorf("emitwave: missing required config %q", name)
		}
	}
	hc := cfg.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{cfg: cfg, http: hc}, nil
}

func (c *Client) Trigger(ctx context.Context, channel, event string, data any) error {
	if data == nil {
		// Match node-sdk's default so the "no data" wire payload is "data":{}
		// across SDKs, not "data":null from one and "data":{} from another.
		data = map[string]any{}
	}
	body, err := json.Marshal(map[string]any{"name": event, "channel": channel, "data": data})
	if err != nil {
		return fmt.Errorf("emitwave: marshal payload: %w", err)
	}
	path := "/apps/" + c.cfg.AppID + "/events"
	q := map[string]string{
		"auth_key":       c.cfg.Key,
		"auth_timestamp": strconv.FormatInt(time.Now().Unix(), 10),
		"auth_version":   "1.0",
		"body_sha256":    protocol.BodySHA256(body),
	}
	q["auth_signature"] = protocol.Sign(
		protocol.CanonicalString("POST", path, protocol.CanonicalQuery(q), q["body_sha256"]),
		c.cfg.Secret,
	)

	uv := url.Values{}
	for k, v := range q {
		uv.Set(k, v)
	}
	full := strings.TrimRight(c.cfg.GatewayURL, "/") + path + "?" + uv.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, full, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("emitwave: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("emitwave: transport: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return &EmitwaveError{Status: resp.StatusCode, Body: string(b)}
	}
	return nil
}

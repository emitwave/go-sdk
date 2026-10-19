# Emitwave Go SDK

Server-side Go SDK for publishing signed events to Emitwave.

## Requirements

- Go 1.26 or newer
- An Emitwave application ID and active server credential

Keep the credential secret on the server.

## Publishing authentication

`Key` and `Secret` are a server credential. They are not the public application
key used by browser clients. The SDK hashes the JSON body and HMAC-signs the
method, path, canonical query, and body hash. The gateway validates that
signature before accepting the event.

Browser connections use a separate public application key and end-user Bearer
token. Authenticated subscriptions receive their own short-lived subscription
tokens.

## Publishing to channel types

`Trigger` publishes to all external channel types:

| Channel | Example | Subscriber authorization |
|---|---|---|
| Public | `news` | Connection token only |
| Private | `private-account.42` | Truthy private callback |
| Presence | `presence-room.lobby` | Callback returning `{ id, info }` |
| Public cache | `cache-prices` | Connection token only |
| Private cache | `private-cache-account.42` | Truthy private callback |
| Presence cache | `presence-cache-room.lobby` | Callback returning `{ id, info }` |
| Private encrypted | `private-encrypted-account.42` | Private callback and shared decryption key |

The gateway scopes each name to the configured application. A publisher's
server credential does not authorize end-user subscriptions.

The Go SDK currently publishes the supplied data unchanged. It does not build
the ciphertext envelope required by private-encrypted channels. Use a
compatible trusted encryption publisher and never send plaintext to an
encrypted channel.

## Installation

```bash
go get github.com/emitwave/go-sdk@latest
```

## Publish an event

```go
package main

import (
	"context"
	"log"
	"os"
	"time"

	emitwave "github.com/emitwave/go-sdk"
)

func main() {
	client, err := emitwave.New(emitwave.Config{
		// GatewayURL defaults to https://api.emitwave.com
		AppID:  os.Getenv("EMITWAVE_APP_ID"),
		Key:    os.Getenv("EMITWAVE_KEY"),
		Secret: os.Getenv("EMITWAVE_SECRET"),
	})
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = client.Trigger(ctx, "orders", "order.created", map[string]any{
		"id":    "ord_123",
		"total": 4900,
	})
	if err != nil {
		log.Fatal(err)
	}
}
```

When `data` is `nil`, the SDK sends an empty JSON object.

## Custom HTTP client

```go
client, err := emitwave.New(emitwave.Config{
	GatewayURL: "https://api.emitwave.com",
	AppID:      appID,
	Key:        key,
	Secret:     secret,
	HTTPClient: &http.Client{Timeout: 3 * time.Second},
})
```

## Error handling

```go
var gatewayError *emitwave.EmitwaveError
if err := client.Trigger(ctx, channel, event, data); err != nil {
	if errors.As(err, &gatewayError) {
		log.Printf("Emitwave rejected the event: HTTP %d: %s", gatewayError.Status, gatewayError.Body)
	} else {
		log.Printf("Emitwave transport error: %v", err)
	}
}
```

The SDK returns `*EmitwaveError` for non-2xx gateway responses and wrapped Go
errors for configuration, serialization, request construction, or transport
failures.

## Development

```bash
go test ./...
```


// Command publisher performs one live signed publish via go-sdk, for the
// Phase 10b DoD. Run: go run ./integration/publisher <channel> <amount>
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	emitwave "github.com/emitwave/go-sdk"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: publisher <channel> <amount>")
		os.Exit(2)
	}
	channel := os.Args[1]
	amount, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad amount: %v\n", err)
		os.Exit(2)
	}
	c, err := emitwave.New(emitwave.Config{
		GatewayURL: os.Getenv("EW_GATEWAY_URL"),
		AppID:      os.Getenv("EW_APP_ID"),
		Key:        os.Getenv("EW_APP_KEY"),
		Secret:     os.Getenv("EW_APP_SECRET"),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}
	if err := c.Trigger(context.Background(), channel, "OrderShipped", map[string]any{"amount": amount}); err != nil {
		fmt.Fprintf(os.Stderr, "trigger: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("published")
}

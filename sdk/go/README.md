# Pulse SDK — Go

SDK Go para integração com o pulse-service (Feature Flags + Event Tracking).

## Instalação

```bash
go get github.com/institutoitinerante/pulse-service/sdk/go
```

## Uso

```go
package main

import (
	"context"
	"fmt"
	"log"

	pulse "github.com/institutoitinerante/pulse-service/sdk/go"
)

func main() {
	client, err := pulse.NewClient("https://pulse.institutoitinerante.com.br")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	// Feature Flags
	enabled, _ := client.IsEnabled(ctx, "dark-mode")
	fmt.Println("dark-mode enabled:", enabled)

	// Event Tracking
	userID := "user-123"
	client.TrackEvent(ctx, pulse.TrackEventRequest{
		EventName:  "page_view",
		UserID:     &userID,
		Properties: map[string]interface{}{"page": "/home"},
	})
}
```

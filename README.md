# go-sumsub

SumSub API client for Go.

## Installation

```bash
go get github.com/thrownew/go-sumsub
```

## Documentation

[https://docs.sumsub.com/docs/overview](https://docs.sumsub.com/docs/overview)
[https://docs.sumsub.com/reference/about-sumsub-api](https://docs.sumsub.com/reference/about-sumsub-api)

## Supported endpoints
- [Health check](https://docs.sumsub.com/reference/review-api-health)
- [Generate SDK access token](https://docs.sumsub.com/reference/generate-access-token)
- [Generate external WebSDK link](https://docs.sumsub.com/reference/generate-websdk-external-link)

Feel free to open an issue or PR if you need more endpoints.

## Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

	sumsub "github.com/thrownew/go-sumsub"
)

func main() {
	cli := sumsub.NewClient(
		"api_token",
		sumsub.NewHMACSigner("api_secret"),
	)
	
	err := cli.Health(context.Background())
	if err != nil {
		panic(fmt.Errorf("health: %w", err))
	}

	resp, err := cli.GenerateAccessTokenSDK(context.Background(), sumsub.GenerateAccessTokenSDKRequest{
		TTL:       time.Minute,
		UserID:    "1000",
		LevelName: "default-level",
	})
	if err != nil {
		panic(fmt.Errorf("generate access token sdk: %w", err))
	}
	fmt.Println("Token: ", resp.Token)
}
```
[![Release](https://img.shields.io/github/release/thrownew/go-sumsub.svg)](https://github.com/thrownew/go-sumsub/releases/latest)
[![License](https://img.shields.io/github/license/thrownew/go-sumsub.svg)](https://raw.githubusercontent.com/thrownew/go-sumsub/master/LICENSE)
[![Godocs](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/thrownew/go-sumsub)

[![Build Status](https://github.com/thrownew/go-sumsub/workflows/CI/badge.svg)](https://github.com/thrownew/go-sumsub/actions)
[![codecov](https://codecov.io/gh/thrownew/go-sumsub/branch/master/graph/badge.svg)](https://codecov.io/gh/thrownew/go-sumsub)
[![Go Report Card](https://goreportcard.com/badge/github.com/thrownew/go-sumsub)](https://goreportcard.com/report/github.com/thrownew/go-sumsub)

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
- [Get applicant review status](https://docs.sumsub.com/reference/get-applicant-review-status)
- [Get applicant data](https://docs.sumsub.com/reference/get-applicant-data)
- [Get applicant data (externalUserId)](https://docs.sumsub.com/reference/get-applicant-data-via-externaluserid)

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
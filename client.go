package sumsub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

var _ error = (*APIError)(nil)

const (
	Host = "api.sumsub.com"
)

type (
	Client struct {
		host   string
		token  string
		signer Signer
		now    NowFunc
		cli    *http.Client
	}

	Signer interface {
		Sign(t time.Time, method, uri string, payload []byte) string
	}

	NowFunc func() time.Time

	options struct {
		Host       string
		HTTPClient *http.Client
		NowFunc    NowFunc
	}

	Opt func(*options)
)

func NewClient(token string, signer Signer, opts ...Opt) *Client {
	o := options{
		Host: Host,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: func() *net.Dialer {
					return &net.Dialer{
						Timeout:   30 * time.Second,
						KeepAlive: 30 * time.Second,
					}
				}().DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
			Timeout: 30 * time.Second,
		},
		NowFunc: time.Now,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return &Client{
		host:   o.Host,
		cli:    o.HTTPClient,
		token:  token,
		signer: signer,
	}
}

func WithHost(host string) Opt {
	return func(opts *options) {
		opts.Host = host
	}
}

func WithHTTPClient(cli *http.Client) Opt {
	return func(opts *options) {
		opts.HTTPClient = cli
	}
}

func WithNowFunc(f NowFunc) Opt {
	return func(opts *options) {
		opts.NowFunc = f
	}
}

type (
	APIError struct {
		Description   string
		Code          int
		CorrelationID string
		ErrorCode     int
		ErrorName     string
	}

	GenerateAccessTokenSDKRequest struct {
		TTL       time.Duration
		UserID    string
		LevelName string
	}

	GenerateAccessTokenSDKResponse struct {
		Token  string
		UserID string
	}

	GenerateExternalWebSDKLinkRequest struct {
		LevelName string
		UserID    string
		TTL       time.Duration
		Lang      string
	}

	GenerateExternalWebSDKLinkResponse struct {
		Link string
	}
)

type (
	respError struct {
		Description   string `json:"description"`
		Code          int    `json:"code"`
		CorrelationID string `json:"correlationId"`
		ErrorCode     int    `json:"errorCode"`
		ErrorName     string `json:"errorName"`
	}

	reqHealth struct {
	}

	respHealth struct {
	}

	reqGenerateAccessTokenSDK struct {
		TTLInSecs int64  `json:"ttlInSecs"`
		UserID    string `json:"userId"`
		LevelName string `json:"levelName"`
	}

	respGenerateAccessTokenSDK struct {
		Token  string `json:"token"`
		UserID string `json:"userId"`
	}

	reqGenerateExternalWebSDKLink struct {
	}

	respGenerateExternalWebSDKLink struct {
		Link string `json:"link"`
	}
)

// GenerateAccessTokenSDK Use this method to generate a new access token for SDK
// https://docs.sumsub.com/reference/generate-access-token
func (c *Client) GenerateAccessTokenSDK(ctx context.Context, req GenerateAccessTokenSDKRequest) (GenerateAccessTokenSDKResponse, error) {
	resp, err := call[reqGenerateAccessTokenSDK, respGenerateAccessTokenSDK](ctx, c,
		http.MethodPost,
		"/resources/accessTokens/sdk",
		reqGenerateAccessTokenSDK{
			TTLInSecs: int64(req.TTL.Seconds()),
			UserID:    req.UserID,
			LevelName: req.LevelName,
		},
	)
	if err != nil {
		return GenerateAccessTokenSDKResponse{}, fmt.Errorf("call: %w", err)
	}

	// response validation
	if resp.UserID != req.UserID {
		return GenerateAccessTokenSDKResponse{}, fmt.Errorf("user id mismatch: `%s` not equal `%s`", resp.UserID, req.UserID)
	}

	//nolint: gosimple
	return GenerateAccessTokenSDKResponse{
		Token:  resp.Token,
		UserID: resp.UserID,
	}, nil
}

// GenerateExternalWebSDKLink Use this method to create an external link to the WebSDK for the specified applicant
// https://docs.sumsub.com/reference/generate-websdk-external-link
func (c *Client) GenerateExternalWebSDKLink(ctx context.Context, req GenerateExternalWebSDKLinkRequest) (GenerateExternalWebSDKLinkResponse, error) {
	resp, err := call[reqGenerateExternalWebSDKLink, respGenerateExternalWebSDKLink](ctx, c,
		http.MethodPost,
		(&url.URL{
			Path: fmt.Sprintf("/resources/sdkIntegrations/levels/%s/websdkLink", url.PathEscape(req.LevelName)),
			RawQuery: url.Values{
				"externalUserId": {req.UserID},
				"ttlInSecs":      {fmt.Sprintf("%d", int64(req.TTL.Seconds()))},
				"lang":           {req.Lang},
			}.Encode(),
		}).String(),
		reqGenerateExternalWebSDKLink{},
	)
	if err != nil {
		return GenerateExternalWebSDKLinkResponse{}, fmt.Errorf("call: %w", err)
	}

	// answer validation
	if _, err = url.Parse(resp.Link); err != nil {
		return GenerateExternalWebSDKLinkResponse{}, fmt.Errorf("invalid link: %w", err)
	}

	//nolint: gosimple
	return GenerateExternalWebSDKLinkResponse{
		Link: resp.Link,
	}, nil
}

// Health Use this method to check the operational status of the API
// https://docs.sumsub.com/reference/review-api-health
func (c *Client) Health(ctx context.Context) error {
	_, err := call[reqHealth, respHealth](ctx, c, http.MethodGet, "/resources/status/api", reqHealth{})
	return err
}

func call[Q, A any](ctx context.Context, cli *Client, method string, uri string, query Q) (A, error) {
	var answer A

	payload, err := json.Marshal(query)
	if err != nil {
		return answer, fmt.Errorf("marshal: %w", err)
	}

	var b io.Reader
	if len(payload) > 0 {
		b = bytes.NewReader(payload)
	}

	now := cli.now()

	req, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("https://%s%s", cli.host, uri), b)
	if err != nil {
		return answer, fmt.Errorf("http: new request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-App-Token", cli.token)
	req.Header.Set("X-App-Access-Ts", fmt.Sprintf("%d", now.Unix()))
	req.Header.Set("X-App-Access-Sig", cli.signer.Sign(now, method, uri, payload))

	resp, err := cli.cli.Do(req)
	if err != nil {
		return answer, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close() //nolint: errcheck

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return answer, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if body != nil && json.Valid(body) {
			var e respError
			if err = json.Unmarshal(body, &e); err == nil {
				return answer, &APIError{
					Description:   e.Description,
					Code:          e.Code,
					CorrelationID: e.CorrelationID,
					ErrorCode:     e.ErrorCode,
					ErrorName:     e.ErrorName,
				}
			}
		}
		return answer, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	if body != nil {
		if !json.Valid(body) {
			return answer, fmt.Errorf("json: not valid")
		}
		if err = json.Unmarshal(body, &answer); err != nil {
			return answer, fmt.Errorf("json: umarshal: %w", err)
		}
	}

	return answer, nil
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s (code: %d, errorCode: %d, correlationId: %s)", e.Description, e.Code, e.ErrorCode, e.CorrelationID)
}

func AsAPIError(err error) (*APIError, bool) {
	var e *APIError
	if !errors.As(err, &e) {
		return nil, false
	}
	return e, true
}

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
		now:    o.NowFunc,
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
		URL string
	}

	ApplicantReviewStatusRequest struct {
		ApplicantID string
	}

	ApplicantReviewStatusResponse struct {
		ReviewID            string
		AttemptID           string
		AttemptCnt          int
		ElapsedSincePending time.Duration
		ElapsedSinceQueued  time.Duration
		Reprocessing        bool
		CreateDate          time.Time
		ReviewDate          time.Time
		ReviewResult        ReviewResult
		ReviewStatus        string
		Priority            int
	}

	ReviewResult struct {
		ModerationComment string
		ClientComment     string
		ReviewAnswer      string
		RejectLabels      []string
		ReviewRejectType  string
	}

	ApplicantDataRequest struct {
		ApplicantID    string
		ExternalUserID string
	}

	ApplicantDataResponse struct {
		ID                string
		CreatedAt         time.Time
		CreatedBy         string
		Key               string
		ClientID          string
		InspectionID      string
		ExternalUserID    string
		FixedInfo         FixedInfo
		Info              Info
		Email             string
		ApplicantPlatform string
		Agreement         Agreement
		Review            Review
		RequiredIDDocs    RequiredIDDocs
		Lang              string
		Type              string
	}

	CreateApplicantRequest struct {
		FixedInfo      FixedInfo
		ExternalUserID string
		Email          *string
		Phone          *string
	}

	CreateApplicantResponse struct {
		// only returning the ID for now, get the complete applicant's data by calling `ApplicantData`
		ID string
	}

	FixedInfo struct {
		FirstName string
		LastName  string
		DOB       *time.Time
	}

	Info struct {
		FirstName   string
		FirstNameEn string
		LastName    string
		LastNameEn  string
		DOB         time.Time
		Country     string
		IDDocs      []IDDoc
	}

	IDDoc struct {
		IDDocType   string
		Country     string
		FirstName   string
		FirstNameEn string
		LastName    string
		LastNameEn  string
		ValidUntil  time.Time
		Number      string
		DOB         time.Time
		MRZLine1    string
		MRZLine2    string
		MRZLine3    string
	}

	Agreement struct {
		CreatedAt  time.Time
		AcceptedAt time.Time
		Source     string
		RecordIDs  []string
	}

	Review struct {
		ReviewID           string
		AttemptID          string
		AttemptCnt         int
		LevelName          string
		LevelAutoCheckMode string
		CreateDate         time.Time
		ReviewStatus       string
		Priority           int
	}

	RequiredIDDocs struct {
		DocSets DocSets
	}

	DocSets []DocSet

	DocSet struct {
		IDDocSetType string
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

	respTime struct {
		time.Time
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
		URL string `json:"url"`
	}

	reqApplicantReviewStatus struct {
	}

	respApplicantReviewStatus struct {
		ReviewID            string   `json:"reviewId"`
		AttemptID           string   `json:"attemptId"`
		AttemptCnt          int      `json:"attemptCnt"`
		ElapsedSincePending int64    `json:"elapsedSincePendingMs"`
		ElapsedSinceQueued  int64    `json:"elapsedSinceQueuedMs"`
		Reprocessing        bool     `json:"reprocessing"`
		CreateDate          respTime `json:"createDate"`
		ReviewDate          respTime `json:"reviewDate"`
		ReviewResult        struct {
			ModerationComment string   `json:"moderationComment"`
			ClientComment     string   `json:"clientComment"`
			ReviewAnswer      string   `json:"reviewAnswer"`
			RejectLabels      []string `json:"rejectLabels"`
			ReviewRejectType  string   `json:"reviewRejectType"`
		} `json:"reviewResult"`
		ReviewStatus string `json:"reviewStatus"`
		Priority     int    `json:"priority"`
	}

	reqApplicantData struct {
	}

	respApplicantData struct {
		ID             string   `json:"id"`
		CreatedAt      respTime `json:"createdAt"`
		CreatedBy      string   `json:"createdBy"`
		Key            string   `json:"key"`
		ClientID       string   `json:"clientId"`
		InspectionID   string   `json:"inspectionId"`
		ExternalUserID string   `json:"externalUserId"`
		FixedInfo      struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		}
		Info struct {
			FirstName   string   `json:"firstName"`
			FirstNameEn string   `json:"firstNameEn"`
			LastName    string   `json:"lastName"`
			LastNameEn  string   `json:"lastNameEn"`
			DOB         respTime `json:"dob"`
			Country     string   `json:"country"`
			IDDocs      []struct {
				IDDocType   string   `json:"idDocType"`
				Country     string   `json:"country"`
				FirstName   string   `json:"firstName"`
				FirstNameEn string   `json:"firstNameEn"`
				LastName    string   `json:"lastName"`
				LastNameEn  string   `json:"lastNameEn"`
				ValidUntil  respTime `json:"validUntil"`
				Number      string   `json:"number"`
				DOB         respTime `json:"dob"`
				MRZLine1    string   `json:"mrzLine1"`
				MRZLine2    string   `json:"mrzLine2"`
				MRZLine3    string   `json:"mrzLine3"`
			}
		}
		Email             string `json:"email"`
		ApplicantPlatform string `json:"applicantPlatform"`
		Agreement         struct {
			CreatedAt  respTime `json:"createdAt"`
			AcceptedAt respTime `json:"acceptedAt"`
			Source     string   `json:"source"`
			RecordIDs  []string `json:"recordIds"`
		} `json:"agreement"`
		RequiredIDDocs struct {
			DocSets []struct {
				IDDocSetType string `json:"idDocSetType"`
			} `json:"docSets"`
		}
		Review struct {
			ReviewID           string   `json:"reviewId"`
			AttemptID          string   `json:"attemptId"`
			AttemptCnt         int      `json:"attemptCnt"`
			LevelName          string   `json:"levelName"`
			LevelAutoCheckMode string   `json:"levelAutoCheckMode"`
			CreateDate         respTime `json:"createDate"`
			ReviewStatus       string   `json:"reviewStatus"`
			Priority           int      `json:"priority"`
		} `json:"review"`
		Lang string `json:"lang"`
		Type string `json:"type"`
	}

	reqCreateApplicant struct {
		FixedInfo struct {
			FirstName string  `json:"firstName"`
			LastName  string  `json:"lastName"`
			Dob       *string `json:"dob"`
		} `json:"fixedInfo"`
		ExternalUserID string  `json:"externalUserId"`
		Email          *string `json:"email"`
		Phone          *string `json:"phone"`
	}

	respCreateApplicant struct {
		ID string `json:"id"`
	}
)

func (t *respTime) UnmarshalJSON(b []byte) error {
	if len(b) < 2 || b[0] != '"' || b[len(b)-1] != '"' {
		return fmt.Errorf("invalid time format")
	}
	str := string(b[1 : len(b)-1])
	var err error
	// try to parse time with different layouts
	for _, layout := range []string{"2006-01-02", "2006-01-02 15:04:05", "2006-01-02 15:04:05-0700"} {
		t.Time, err = time.Parse(layout, str)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("parse time: undefined layout: %s", str)
	}
	return nil
}

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

	//nolint: staticcheck
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
	if _, err = url.Parse(resp.URL); err != nil {
		return GenerateExternalWebSDKLinkResponse{}, fmt.Errorf("invalid link: %w", err)
	}

	//nolint: staticcheck
	return GenerateExternalWebSDKLinkResponse{
		URL: resp.URL,
	}, nil
}

// ApplicantReviewStatus Use this method when utilizing the WebSDK or MobileSDK to get the review status. Both SDKs will show the rejection reasons and associated moderation comments.
// https://docs.sumsub.com/reference/get-applicant-review-status
func (c *Client) ApplicantReviewStatus(ctx context.Context, req ApplicantReviewStatusRequest) (ApplicantReviewStatusResponse, error) {
	resp, err := call[reqApplicantReviewStatus, respApplicantReviewStatus](ctx, c,
		http.MethodGet,
		fmt.Sprintf("/resources/applicants/%s/status", url.PathEscape(req.ApplicantID)),
		reqApplicantReviewStatus{},
	)
	if err != nil {
		return ApplicantReviewStatusResponse{}, fmt.Errorf("call: %w", err)
	}

	return ApplicantReviewStatusResponse{
		ReviewID:            resp.ReviewID,
		AttemptID:           resp.AttemptID,
		AttemptCnt:          resp.AttemptCnt,
		ElapsedSincePending: time.Duration(resp.ElapsedSincePending) * time.Millisecond,
		ElapsedSinceQueued:  time.Duration(resp.ElapsedSinceQueued) * time.Millisecond,
		Reprocessing:        resp.Reprocessing,
		CreateDate:          resp.CreateDate.Time,
		ReviewDate:          resp.ReviewDate.Time,
		ReviewResult: ReviewResult{
			ModerationComment: resp.ReviewResult.ModerationComment,
			ClientComment:     resp.ReviewResult.ClientComment,
			ReviewAnswer:      resp.ReviewResult.ReviewAnswer,
			RejectLabels:      resp.ReviewResult.RejectLabels,
			ReviewRejectType:  resp.ReviewResult.ReviewRejectType,
		},
		ReviewStatus: resp.ReviewStatus,
		Priority:     resp.Priority,
	}, nil
}

// ApplicantData Use this method in cases the applicant ID is not yet known to you. For example, when the WebSDK has created an applicant for you and we called your webhook endpoint.
// https://docs.sumsub.com/reference/get-applicant-data
// https://docs.sumsub.com/reference/get-applicant-data-via-externaluserid
func (c *Client) ApplicantData(ctx context.Context, req ApplicantDataRequest) (ApplicantDataResponse, error) {
	var uri string

	switch {
	case len(req.ApplicantID) > 0:
		// https://docs.sumsub.com/reference/get-applicant-data
		uri = fmt.Sprintf("/resources/applicants/%s/one", url.PathEscape(req.ApplicantID))
	case len(req.ExternalUserID) > 0:
		// https://docs.sumsub.com/reference/get-applicant-data-via-externaluserid
		uri = fmt.Sprintf("/resources/applicants/-;externalUserId=%s/one", url.PathEscape(req.ExternalUserID))
	default:
		return ApplicantDataResponse{}, errors.New("applicant id or external user id required")
	}

	resp, err := call[reqApplicantData, respApplicantData](ctx, c,
		http.MethodGet,
		uri,
		reqApplicantData{},
	)
	if err != nil {
		return ApplicantDataResponse{}, fmt.Errorf("call: %w", err)
	}

	return ApplicantDataResponse{
		ID:             resp.ID,
		CreatedAt:      resp.CreatedAt.Time,
		CreatedBy:      resp.CreatedBy,
		Key:            resp.Key,
		ClientID:       resp.ClientID,
		InspectionID:   resp.InspectionID,
		ExternalUserID: resp.ExternalUserID,
		FixedInfo: FixedInfo{
			FirstName: resp.FixedInfo.FirstName,
			LastName:  resp.FixedInfo.LastName,
		},
		Info: Info{
			FirstName:   resp.Info.FirstName,
			FirstNameEn: resp.Info.FirstNameEn,
			LastName:    resp.Info.LastName,
			LastNameEn:  resp.Info.LastNameEn,
			DOB:         resp.Info.DOB.Time,
			Country:     resp.Info.Country,
			IDDocs: func() []IDDoc {
				var docs []IDDoc
				for _, d := range resp.Info.IDDocs {
					docs = append(docs, IDDoc{
						IDDocType:   d.IDDocType,
						Country:     d.Country,
						FirstName:   d.FirstName,
						FirstNameEn: d.FirstNameEn,
						LastName:    d.LastName,
						LastNameEn:  d.LastNameEn,
						ValidUntil:  d.ValidUntil.Time,
						Number:      d.Number,
						DOB:         d.DOB.Time,
						MRZLine1:    d.MRZLine1,
						MRZLine2:    d.MRZLine2,
						MRZLine3:    d.MRZLine3,
					})
				}
				return docs
			}(),
		},
		Email:             resp.Email,
		ApplicantPlatform: resp.ApplicantPlatform,
		Agreement: struct {
			CreatedAt  time.Time
			AcceptedAt time.Time
			Source     string
			RecordIDs  []string
		}{
			CreatedAt:  resp.Agreement.CreatedAt.Time,
			AcceptedAt: resp.Agreement.AcceptedAt.Time,
			Source:     resp.Agreement.Source,
			RecordIDs:  resp.Agreement.RecordIDs,
		},
		RequiredIDDocs: RequiredIDDocs{
			DocSets: func() DocSets {
				var sets DocSets
				for _, s := range resp.RequiredIDDocs.DocSets {
					sets = append(sets, DocSet{
						IDDocSetType: s.IDDocSetType,
					})
				}
				return sets
			}(),
		},
		Review: Review{
			ReviewID:           resp.Review.ReviewID,
			AttemptID:          resp.Review.AttemptID,
			AttemptCnt:         resp.Review.AttemptCnt,
			LevelName:          resp.Review.LevelName,
			LevelAutoCheckMode: resp.Review.LevelAutoCheckMode,
			CreateDate:         resp.Review.CreateDate.Time,
			ReviewStatus:       resp.Review.ReviewStatus,
			Priority:           resp.Review.Priority,
		},
		Lang: resp.Lang,
		Type: resp.Type,
	}, nil
}

// CreateApplicant Use this method to create an applicant on sumsub via API.
// https://docs.sumsub.com/reference/create-applicant
func (c *Client) CreateApplicant(ctx context.Context, req CreateApplicantRequest) (CreateApplicantResponse, error) {
	var dateString string

	if req.FixedInfo.DOB != nil {
		dateString = req.FixedInfo.DOB.Format("2006-01-02")
	}

	resp, err := call[reqCreateApplicant, respCreateApplicant](ctx, c,
		http.MethodPost,
		"/resources/applicants",
		reqCreateApplicant{
			FixedInfo: struct {
				FirstName string  `json:"firstName"`
				LastName  string  `json:"lastName"`
				Dob       *string `json:"dob"`
			}{
				FirstName: req.FixedInfo.FirstName,
				LastName:  req.FixedInfo.LastName,
				Dob:       &dateString,
			},
			ExternalUserID: req.ExternalUserID,
			Email:          req.Email,
			Phone:          req.Phone,
		},
	)

	if err != nil {
		return CreateApplicantResponse{}, fmt.Errorf("call: %w", err)
	}

	return CreateApplicantResponse{
		ID: resp.ID,
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

package sumsub

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransportModels(t *testing.T) {
	cases := []struct {
		msg   json.RawMessage
		model any
	}{
		{
			msg: json.RawMessage(`{
  "description": "Error analyzing file, unsupported format or corrupted file",
  "code": 400,
  "correlationId": "req-5fd59b09-7f5e-41cd-a86b-38a4e6d57e08",
  "errorCode": 1004,
  "errorName": "corrupted-file"
}`),
			model: respError{},
		},
		{
			msg: json.RawMessage(`{
  "token": "_act-b8ebfb63-5f24-4b89-9c08-000000000000",
  "userId": "johndoeID"
}`),
			model: respGenerateAccessTokenSDK{},
		},
		{
			msg: json.RawMessage(`{
    "url": "https://api.sumsub.com/idensic/l/#/lPDnIKwzmxPfDohk"
}`),
			model: respGenerateExternalWebSDKLink{},
		},
		{
			msg: json.RawMessage(`{
  "reviewId": "anGLu",
  "attemptId": "MAnCa",
  "attemptCnt": 0,
  "elapsedSincePendingMs": 27,
  "elapsedSinceQueuedMs": 27,
  "reprocessing": false,
  "levelAutoCheckMode": null,
  "createDate": "2024-03-18 06:46:20+0000",
  "reviewDate": "2024-03-18 06:46:20+0000",
  "reviewResult": {
    "moderationComment": "We could not verify your profile. If you have any questions, please contact the Company where you try to verify your profile",
    "clientComment": "User was misled/forced to create this account by a third party",
    "reviewAnswer": "RED",
    "rejectLabels": [
      "FRAUDULENT_PATTERNS"
    ],
    "reviewRejectType": "FINAL"
  },
  "reviewStatus": "init",
  "priority": 0
}`),
			model: respApplicantReviewStatus{},
		},
		{
			msg: json.RawMessage(`{
  "id": "5b594ade0a975a36c9349e66",
  "createdAt": "2020-06-24 05:05:14",
  "clientId": "ClientName",
  "inspectionId": "5b594ade0a975a36c9379e67",
  "externalUserId": "SomeExternalUserId",
  "fixedInfo": {
    "firstName": "Chris",
    "lastName": "Smith"
  },
  "info": {
    "firstName": "CHRISTIAN",
    "firstNameEn": "CHRISTIAN",
    "lastName": "SMITH",
    "lastNameEn": "SMITH",
    "dob": "1989-07-16",
    "country": "DEU",
    "idDocs": [
      {
        "idDocType": "ID_CARD",
        "country": "DEU",
        "firstName": "CHRISTIAN",
        "firstNameEn": "CHRISTIAN",
        "lastName": "SMITH",
        "lastNameEn": "SMITH",
        "validUntil": "2028-09-04",
        "number": "LGXX359T8",
        "dob": "1989-07-16",
        "mrzLine1": "IDD<<LGXX359T88<<<<<<<<<<<<<<<",
        "mrzLine2": "8907167<2809045D<<<<<<<<<<<<<8",
        "mrzLine3": "SMITH<<CHRISTIAN<<<<<<<<<<<<<<"
      }
    ]
  },
  "agreement": {
    "createdAt": "2020-06-24 04:18:40",
    "source": "WebSDK",
    "targets": [
      "By clicking Next, I accept [the Terms and Conditions](https://www.sumsub.com/consent-to-personal-data-processing/)",
      "I agree to the processing of my personal data, as described in [the Consent to Personal Data Processing](https://sumsub.com/consent-to-personal-data-processing/)"
    ]
  },
  "email": "christman1@gmail.com",
  "applicantPlatform": "Android",
  "requiredIdDocs": {
    "docSets": [
      {
        "idDocSetType": "IDENTITY",
        "types": [
          "PASSPORT",
          "ID_CARD"
        ]
      },
      {
        "idDocSetType": "SELFIE",
        "types": [
          "SELFIE"
        ]
      }
    ]
  },
  "review": {
    "elapsedSincePendingMs": 115879,
    "elapsedSinceQueuedMs": 95785,
    "reprocessing": true,
    "levelName": "basic-kyc",
    "createDate": "2020-06-24 05:11:02+0000",
    "reviewDate": "2020-06-24 05:12:58+0000",
    "reviewResult": {
      "reviewAnswer": "GREEN"
    },
    "reviewStatus": "completed"
  },
  "lang": "de",
  "type": "individual"
}`),
			model: respApplicantData{},
		},
	}
	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d (%T)", i, c.model), func(t *testing.T) {
			require.True(t, json.Valid(c.msg), "invalid json")
			err := json.Unmarshal(c.msg, &c.model)
			require.NoError(t, err)
			msg, err := json.Marshal(c.model)
			require.NoError(t, err)
			require.JSONEq(t, string(c.msg), string(msg))
		})
	}
}

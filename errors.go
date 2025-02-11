package sumsub

const (
	ErrCodeDuplicateDocument                = 1000 // Duplicate document (image, video) was uploaded. Exact equality is taken into account.
	ErrCodeTooManyDocuments                 = 1001 // Applicant contains too many documents. Adding new is not allowed.
	ErrCodeFileTooBig                       = 1002 // Uploaded file is too big (more than 64MB).
	ErrCodeEmptyFile                        = 1003 // Uploaded file is empty (0 bytes).
	ErrCodeCorruptedFile                    = 1004 // File is corrupted or of incorrect format (e.g. PDF file is uploaded as JPEG).
	ErrCodeUnsupportedFileFormat            = 1005 // Unsupported file format (e.g. a TIFF image).
	ErrCodeNoUploadVerificationInProgress   = 1006 // Applicant is being checked. Adding new data is not allowed.
	ErrCodeIncorrectFileSize                = 1007 // The file size must meet the file upload requirements specified in the global settings.
	ErrCodeApplicantMarkedAsDeleted         = 1008 // Applicant is marked as deleted/inactive. No action is allowed to change the status.
	ErrCodeApplicantWithFinalReject         = 1009 // Applicant is rejected with the FINAL rejection type. Adding new data/files is not allowed.
	ErrCodeDocTypeNotInReqDocs              = 1010 // Attempt to upload a document outside of the applicant level set or set of required documents.
	ErrCodeEncryptedFile                    = 1011 // Attempt to open an encrypted file.
	ErrCodeApplicantAlreadyInTheState       = 3000 // Attempt to change the status of the applicant against the logic — the applicant is already in the required state.
	ErrCodeAppTokenInvalidFormat            = 4000 // Invalid format of the X-App-Token value.
	ErrCodeAppTokenNotFound                 = 4001 // App token does not exist (e.g. test env. token used on production).
	ErrCodeAppTokenPrivatePartMismatch      = 4002 // Private part of the token (after dot) does not match public part.
	ErrCodeAppTokenSignatureMismatch        = 4003 // Signature encoded value does not match the request content.
	ErrCodeAppTokenRequestExpired           = 4004 // X-App-Access-Ts does not match the number of seconds since Unix Epoch in UTC.
	ErrCodeAppTokenInvalidValue             = 4005 // Invalid authentication header values were provided.
	ErrCodeAppTokenNotAllAuthParamsProvided = 4006 // Not all required authorization headers were provided.
	ErrCodeAppTokenInvalidParams            = 4007 // Invalid authentication parameters were provided.
	ErrCodeApplicantAlreadyBlacklisted      = 5000 // Attempt to blocklist the applicant that is already blocklisted.
	ErrCodeApplicantAlreadyWhitelisted      = 5001 // Attempt to whitelist the applicant that is already whitelisted.
)

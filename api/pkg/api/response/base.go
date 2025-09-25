package response

type (
	errCode     string
	successCode string
	M           map[string]any
)

var ErrCodes = struct {
	InternalServerError,
	NotFound,
	UnsupportedMediaType,
	InvalidPathValue,
	MalformedJSONBody,
	AuthTokenNotFound,
	InvalidAuthToken,
	InsufficientAuthRole,
	InactiveUser,
	BodyValidation,
	QueryValidation,
	UserAlreadyExists,
	ExternalServiceError,
	BadRequest,
	SessionLimitReached errCode
}{
	InternalServerError:  "internal_error",
	NotFound:             "not_found",
	UnsupportedMediaType: "unsupported_media_type",
	InvalidPathValue:     "invalid_path_value",
	MalformedJSONBody:    "malformed_json_body",
	AuthTokenNotFound:    "auth_token_not_found",
	InvalidAuthToken:     "invalid_auth_token",
	InsufficientAuthRole: "insufficient_auth_role",
	InactiveUser:         "inactive_user",
	BodyValidation:       "body_validation",
	QueryValidation:      "query_validation",
	UserAlreadyExists:    "user_already_exists",
	ExternalServiceError: "external_service_error",
	BadRequest:           "bad_request",
	SessionLimitReached:  "session_limit_reached",
}

var SuccessCodes = struct {
	Ok,
	Created,
	Deleted successCode
}{
	Ok:      "ok",
	Created: "created",
	Deleted: "deleted",
}

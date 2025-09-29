package response

type (
	errCode     string
	successCode string
	M           map[string]any
)

var ErrCodes = struct {
	InternalServerError,
	NotFound,
	BadRequest,
	UnsupportedMediaType,
	MalformedJSONBody,
	BodyValidation,
	FormDataTooLarge errCode
}{
	InternalServerError:  "internal_error",
	NotFound:             "not_found",
	BadRequest:           "bad_request",
	UnsupportedMediaType: "unsupported_media_type",
	MalformedJSONBody:    "malformed_json_body",
	BodyValidation:       "body_validation",
	FormDataTooLarge:     "form_data_too_large",
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

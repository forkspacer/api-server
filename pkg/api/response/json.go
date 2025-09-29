package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/forkspacer/api-server/pkg/utils"
)

type JSONErrorResponse struct {
	Code errCode `json:"code"`
	Data any     `json:"data"`
}

func NewJSONError(code errCode, data any) *JSONErrorResponse {
	return &JSONErrorResponse{Code: code, Data: data}
}

type JSONSuccessResponse struct {
	Code successCode `json:"code"`
	Data any         `json:"data"`
}

func NewJSONSuccess(code successCode, data any) *JSONSuccessResponse {
	return &JSONSuccessResponse{Code: code, Data: data}
}

type Response struct {
	Success *JSONSuccessResponse `json:"success"`
	Error   *JSONErrorResponse   `json:"error"`
}

func JSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

func JSONSuccess(w http.ResponseWriter, statusCode int, successResponse *JSONSuccessResponse) {
	JSON(w, statusCode, Response{Success: successResponse})
}

func JSONCreated(w http.ResponseWriter) {
	JSONSuccess(w, 201, NewJSONSuccess(SuccessCodes.Created, nil))
}

func JSONDeleted(w http.ResponseWriter) {
	JSON(w, 204, nil)
}

func JSONError(w http.ResponseWriter, statusCode int, errorResponse *JSONErrorResponse) {
	JSON(w, statusCode, Response{Error: errorResponse})
}

func JSONBadRequest(w http.ResponseWriter, data any) {
	JSONError(w, 400, NewJSONError(ErrCodes.BadRequest, data))
}

func JSONMalformedJSONBody(w http.ResponseWriter) {
	JSONError(w, 400, NewJSONError(ErrCodes.MalformedJSONBody, nil))
}

func JSONBodyValidationError(w http.ResponseWriter, errs map[string]string) {
	JSONError(w, 400, NewJSONError(ErrCodes.BodyValidation, errs))
}

func JSONNotFound(w http.ResponseWriter) {
	JSONError(w, 404, NewJSONError(ErrCodes.NotFound, "Not Found"))
}

func JSONFormDataTooLarge(w http.ResponseWriter, limit *int64) {
	if limit == nil {
		JSONError(w, 413,
			NewJSONError(ErrCodes.FormDataTooLarge, "Form data too large"),
		)
		return
	}

	JSONError(w, 413,
		NewJSONError(
			ErrCodes.FormDataTooLarge,
			fmt.Sprintf("Form data too large (limit: %s bytes)", utils.FormatBytes(*limit)),
		),
	)
}

func JSONUnsopportedMediaType(w http.ResponseWriter, expectedMediaType string) {
	JSONError(w, 415,
		NewJSONError(
			ErrCodes.UnsupportedMediaType,
			fmt.Sprintf("Unsupported Media Type (expected: %s)", expectedMediaType),
		),
	)
}

func JSONInternal(w http.ResponseWriter) {
	JSONError(w, 500, NewJSONError(ErrCodes.InternalServerError, "Internal Server Error"))
}

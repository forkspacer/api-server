package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/types"
	"github.com/go-playground/validator/v10"
)

func JSONBodyValidate(w http.ResponseWriter, r *http.Request, structData any) error {
	contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")
	if contentType != "application/json" {
		response.JSONUnsopportedMediaType(w, "application/json")
		return fmt.Errorf("unsupported media type")
	}

	if err := json.NewDecoder(r.Body).Decode(structData); err != nil {
		response.JSONMalformedJSONBody(w)
		return fmt.Errorf("failed to decode request body: %w", err)
	}

	if err := Validate.StructCtx(r.Context(), structData); err != nil {
		switch errs := err.(type) {
		case validator.ValidationErrors:
			response.JSONBodyValidationError(w, validationErrors2Map(errs))
			return fmt.Errorf("request body validation failed: %w", errs)
		default:
			response.JSONInternal(w)
			return types.ErrInternal
		}
	}

	return nil
}

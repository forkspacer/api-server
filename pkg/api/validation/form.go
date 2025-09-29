package validation

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/types"
	"github.com/go-playground/validator/v10"
)

func FormDataBodyValidate(w http.ResponseWriter, r *http.Request, structData any) error {
	contentType, _, _ := strings.Cut(r.Header.Get("Content-Type"), ";")
	if contentType != "multipart/form-data" {
		response.JSONUnsopportedMediaType(w, "multipart/form-data")
		return fmt.Errorf("unsupported media type")
	}

	if err := Validate.StructCtx(r.Context(), structData); err != nil {
		switch errs := err.(type) {
		case validator.ValidationErrors:
			response.JSONBodyValidationError(w, errs.Translate(GetTranslation("en")))
			return fmt.Errorf("request body validation failed: %w", errs)
		default:
			response.JSONInternal(w)
			return types.ErrInternal
		}
	}

	return nil
}

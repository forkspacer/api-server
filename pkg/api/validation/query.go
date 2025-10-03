package validation

import (
	"context"
	"fmt"
	"net/http"

	"github.com/forkspacer/api-server/pkg/api/response"
	"github.com/forkspacer/api-server/pkg/types"
	"github.com/go-playground/validator/v10"
)

func URLParamsValidate(ctx context.Context, w http.ResponseWriter, structData any) error {
	if err := Validate.StructCtx(ctx, structData); err != nil {
		switch errs := err.(type) {
		case validator.ValidationErrors:
			response.JSONQueryValidationError(w, errs.Translate(GetTranslation("en")))
			return fmt.Errorf("request query validation failed: %w", errs)
		default:
			response.JSONInternal(w)
			return types.ErrInternal
		}
	}

	return nil
}

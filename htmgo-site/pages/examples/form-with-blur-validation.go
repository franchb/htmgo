package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func FormWithBlurValidation(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &FormWithBlurValidationSnippet)
	return Index(ctx)
}

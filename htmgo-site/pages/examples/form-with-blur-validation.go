package examples

import (
	"github.com/franchb/htmgo/framework/h"
)

func FormWithBlurValidation(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &FormWithBlurValidationSnippet)
	return Index(ctx)
}

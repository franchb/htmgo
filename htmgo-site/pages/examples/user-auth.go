package examples

import (
	"github.com/franchb/htmgo/framework/v2/h"
)

func UserAuthExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &UserAuthSnippet)
	return Index(ctx)
}

package examples

import (
	"github.com/franchb/htmgo/framework/h"
)

func UserAuthExample(ctx *h.RequestContext) *h.Page {
	SetSnippet(ctx, &UserAuthSnippet)
	return Index(ctx)
}

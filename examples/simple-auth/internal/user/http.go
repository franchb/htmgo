package user

import (
	"github.com/franchb/htmgo/framework/v2/h"
	"simpleauth/internal/db"
)

func GetUserOrRedirect(ctx *h.RequestContext) (db.User, bool) {
	user, err := GetUserFromSession(ctx)

	if err != nil {
		_ = ctx.Redirect("/login", 302)
		return db.User{}, false
	}

	return user, true
}

package pages

import "github.com/franchb/htmgo/framework/h"

func DiscordPage(ctx *h.RequestContext) *h.Page {
	_ = ctx.Redirect("https://discord.com/invite/nwQY4h6DtJ", 302)
	return h.EmptyPage()
}

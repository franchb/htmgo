// Package ax provides Go helpers for Alpine.js directives, mirroring the shape
// of framework/hx (htmx helpers). All constants are Alpine 3 attribute names;
// the Ren-returning builders live in builder.go.
package ax

type Attribute = string

const (
	DataAttr       Attribute = "x-data"
	InitAttr       Attribute = "x-init"
	ShowAttr       Attribute = "x-show"
	BindAttr       Attribute = "x-bind"
	OnAttr         Attribute = "x-on"
	TextAttr       Attribute = "x-text"
	HtmlAttr       Attribute = "x-html"
	ModelAttr      Attribute = "x-model"
	ModelableAttr  Attribute = "x-modelable"
	CloakAttr      Attribute = "x-cloak"
	RefAttr        Attribute = "x-ref"
	IgnoreAttr     Attribute = "x-ignore"
	TeleportAttr   Attribute = "x-teleport"
	EffectAttr     Attribute = "x-effect"
	IfAttr         Attribute = "x-if"
	ForAttr        Attribute = "x-for"
	IdAttr         Attribute = "x-id"
	TransitionAttr Attribute = "x-transition"
)

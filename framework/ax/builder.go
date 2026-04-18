package ax

import "github.com/franchb/htmgo/framework/h"

// Simple single-arg directives. Each wraps h.Attribute and returns h.Ren.

func Data(expr string) h.Ren      { return h.Attribute(DataAttr, expr) }
func Init(expr string) h.Ren      { return h.Attribute(InitAttr, expr) }
func Show(expr string) h.Ren      { return h.Attribute(ShowAttr, expr) }
func Text(expr string) h.Ren      { return h.Attribute(TextAttr, expr) }
func Html(expr string) h.Ren      { return h.Attribute(HtmlAttr, expr) }
func Model(expr string) h.Ren     { return h.Attribute(ModelAttr, expr) }
func Effect(expr string) h.Ren    { return h.Attribute(EffectAttr, expr) }
func Modelable(expr string) h.Ren { return h.Attribute(ModelableAttr, expr) }
func If(expr string) h.Ren        { return h.Attribute(IfAttr, expr) }
func For(expr string) h.Ren       { return h.Attribute(ForAttr, expr) }
func Id(scopes string) h.Ren      { return h.Attribute(IdAttr, scopes) }
func Ref(name string) h.Ren       { return h.Attribute(RefAttr, name) }
func Teleport(selector string) h.Ren {
	return h.Attribute(TeleportAttr, selector)
}

// No-arg directives. Alpine accepts an empty string as the attribute value.

func Cloak() h.Ren  { return h.Attribute(CloakAttr, "") }
func Ignore() h.Ren { return h.Attribute(IgnoreAttr, "") }

// Transition emits a bare x-transition="". Richer variants
// (x-transition:enter, .opacity, .duration.500ms) are out of scope;
// use h.Attribute("x-transition:enter", ...) directly when needed.

func Transition() h.Ren { return h.Attribute(TransitionAttr, "") }

// x-bind:* family. Bind is the generic form; the shortcut helpers cover the
// attributes that appear most often in practice.

func Bind(attr, expr string) h.Ren {
	return h.Attribute(BindAttr+":"+attr, expr)
}

func BindClass(expr string) h.Ren    { return Bind("class", expr) }
func BindStyle(expr string) h.Ren    { return Bind("style", expr) }
func BindHref(expr string) h.Ren     { return Bind("href", expr) }
func BindValue(expr string) h.Ren    { return Bind("value", expr) }
func BindDisabled(expr string) h.Ren { return Bind("disabled", expr) }
func BindChecked(expr string) h.Ren  { return Bind("checked", expr) }
func BindId(expr string) h.Ren       { return Bind("id", expr) }

package h

import (
	"github.com/franchb/htmgo/framework/hx"
	"html"
	"html/template"
	"strings"
)

type CustomElement = string

var (
	CachedNodeTag        CustomElement = "htmgo_cache_node"
	CachedNodeByKeyEntry CustomElement = "htmgo_cached_node_by_key_entry"
)

/*
*
void tags are tags that cannot have children
*/
var voidTags = map[string]bool{
	"area":   true,
	"base":   true,
	"br":     true,
	"col":    true,
	"embed":  true,
	"hr":     true,
	"img":    true,
	"input":  true,
	"link":   true,
	"meta":   true,
	"source": true,
	"track":  true,
	"wbr":    true,
}

type ScriptEntry struct {
	Body    string
	ChildOf *Element
}

type RenderContext struct {
	builder        *strings.Builder
	scripts        []ScriptEntry
	currentElement *Element
}

func (ctx *RenderContext) AddScript(funcName string, body string) {
	var sb strings.Builder
	sb.WriteString("\n\t<script id=\"")
	sb.WriteString(funcName)
	sb.WriteString("\">\n\t\tfunction ")
	sb.WriteString(funcName)
	sb.WriteString("(self, event) {\n\t\t\t\tlet e = event;\n\t\t\t\t")
	sb.WriteString(body)
	sb.WriteString("\n\t\t}\n\t</script>")

	ctx.scripts = append(ctx.scripts, ScriptEntry{
		Body:    sb.String(),
		ChildOf: ctx.currentElement,
	})
}

func (node *Element) Render(context *RenderContext) {
	if node == nil {
		return
	}

	context.currentElement = node

	if node.tag == CachedNodeTag {
		meta := node.meta.(*CachedNode)
		meta.Render(context)
		return
	}

	if node.tag == CachedNodeByKeyEntry {
		meta := node.meta.(*ByKeyEntry)
		meta.Render(context)
		return
	}

	// some elements may not have a tag, such as a Fragment
	if node.tag != "" {
		context.builder.WriteString("<")
		context.builder.WriteString(node.tag)
		node.attributes.Each(func(key string, value string) {
			context.builder.WriteString(" ")
			context.builder.WriteString(key)
			if value != "" {
				context.builder.WriteString(`="`)
				context.builder.WriteString(html.EscapeString(value))
				context.builder.WriteString(`"`)
			}
		})
	}

	// Single-pass: render attribute-like children (within the tag)
	renderChildAttrs(context, node.children)

	// close the tag
	if node.tag != "" {
		if voidTags[node.tag] {
			context.builder.WriteString("/")
		}
		context.builder.WriteString(">")
	}

	// void elements do not have children
	if !voidTags[node.tag] {
		// render the content children, handling ChildList inline
		renderChildContent(context, node.children)
	}

	if node.tag != "" {
		renderScripts(context, node)
		if !voidTags[node.tag] {
			context.builder.WriteString("</")
			context.builder.WriteString(node.tag)
			context.builder.WriteString(">")
		}
	}
}

// renderChildAttrs renders only attribute-like children from the list,
// recursing into ChildList without flattening.
func renderChildAttrs(context *RenderContext, children []Ren) {
	for _, child := range children {
		switch c := child.(type) {
		case *AttributeMapOrdered:
			c.Render(context)
		case *AttributeR:
			c.Render(context)
		case *LifeCycle:
			c.Render(context)
		case *ChildList:
			renderChildAttrs(context, c.Children)
		}
	}
}

// renderChildContent renders non-attribute children from the list,
// recursing into ChildList without flattening.
func renderChildContent(context *RenderContext, children []Ren) {
	for _, child := range children {
		switch child.(type) {
		case *AttributeMapOrdered:
			continue
		case *AttributeR:
			continue
		case *LifeCycle:
			continue
		case *ChildList:
			renderChildContent(context, child.(*ChildList).Children)
		default:
			child.Render(context)
		}
	}
}

func renderScripts(context *RenderContext, parent *Element) {
	if len(context.scripts) == 0 {
		return
	}
	n := 0
	for _, script := range context.scripts {
		if script.ChildOf == parent {
			context.builder.WriteString(script.Body)
		} else {
			context.scripts[n] = script
			n++
		}
	}
	// Zero out the tail to release stale Body/ChildOf references,
	// preventing memory retention when RenderContext is pooled.
	for i := n; i < len(context.scripts); i++ {
		context.scripts[i] = ScriptEntry{}
	}
	context.scripts = context.scripts[:n]
}

func (a *AttributeR) Render(context *RenderContext) {
	context.builder.WriteString(" ")
	context.builder.WriteString(a.Name)
	if a.Value != "" {
		context.builder.WriteString(`=`)
		context.builder.WriteString(`"`)
		context.builder.WriteString(html.EscapeString(a.Value))
		context.builder.WriteString(`"`)
	}
}

func (t *TextContent) Render(context *RenderContext) {
	context.builder.WriteString(template.HTMLEscapeString(t.Content))
}

func (r *RawContent) Render(context *RenderContext) {
	context.builder.WriteString(r.Content)
}

func (c *ChildList) Render(context *RenderContext) {
	for _, child := range c.Children {
		child.Render(context)
	}
}

func (j SimpleJsCommand) Render(context *RenderContext) {
	context.builder.WriteString(j.Command)
}

func (j ComplexJsCommand) Render(context *RenderContext) {
	context.builder.WriteString(j.Command)
}

func (p *Partial) Render(context *RenderContext) {
	p.Root.Render(context)
}

func (m *AttributeMapOrdered) Render(context *RenderContext) {
	m.Each(func(key string, value string) {
		context.builder.WriteString(" ")
		context.builder.WriteString(key)
		if value != "" {
			context.builder.WriteString(`="`)
			context.builder.WriteString(html.EscapeString(value))
			context.builder.WriteString(`"`)
		}
	})
}

func (l *LifeCycle) fromAttributeMap(event string, key string, value string, context *RenderContext) {

	if key == hx.GetAttr || key == hx.PatchAttr || key == hx.PostAttr {
		HxTriggerString(hx.ToHtmxTriggerName(event)).Render(context)
	}

	Attribute(key, value).Render(context)
}

func (l *LifeCycle) Render(context *RenderContext) {
	type eventEntry struct {
		event string
		value string
	}

	entries := make([]eventEntry, 0, len(l.handlers))

	for event, commands := range l.handlers {
		var sb strings.Builder

		for _, command := range commands {
			switch c := command.(type) {
			case SimpleJsCommand:
				sb.WriteString("var self=this;var e=event;")
				sb.WriteString(c.Command)
				sb.WriteString(";")
			case ComplexJsCommand:
				context.AddScript(c.TempFuncName, c.Command)
				sb.WriteString(c.TempFuncName)
				sb.WriteString("(this, event);")
			case *AttributeMapOrdered:
				c.Each(func(key string, value string) {
					l.fromAttributeMap(event, key, value, context)
				})
			case *AttributeR:
				l.fromAttributeMap(event, c.Name, c.Value, context)
			}
		}

		if sb.Len() > 0 {
			entries = append(entries, eventEntry{event: event, value: sb.String()})
		}
	}

	if len(entries) == 0 {
		return
	}

	for _, e := range entries {
		Attribute(e.event, e.value).Render(context)
	}
}

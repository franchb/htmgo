package h

import (
	"strings"
	"sync"
)

type Ren interface {
	Render(context *RenderContext)
}

type RenderOptions struct {
	doctype bool
}

type RenderOpt func(context *RenderContext, opt *RenderOptions)

var withDocTypeOpt RenderOpt = func(context *RenderContext, opt *RenderOptions) {
	opt.doctype = true
}

func WithDocType() RenderOpt {
	return withDocTypeOpt
}

var renderContextPool = sync.Pool{
	New: func() any {
		return &RenderContext{
			builder: &strings.Builder{},
		}
	},
}

// Render renders the given node recursively, and returns the resulting string.
func Render(node Ren, opts ...RenderOpt) string {
	context := renderContextPool.Get().(*RenderContext)
	context.builder.Reset()
	if context.scripts != nil {
		context.scripts = context.scripts[:0]
	}
	context.currentElement = nil

	options := &RenderOptions{}

	for _, opt := range opts {
		opt(context, options)
	}

	if options.doctype {
		context.builder.WriteString("<!DOCTYPE html>")
	}

	node.Render(context)
	result := context.builder.String()
	// Don't return oversized builders to the pool; a single large render
	// should not permanently inflate memory for all subsequent renders.
	if context.builder.Cap() <= 64*1024 {
		renderContextPool.Put(context)
	}
	return result
}

package h

import (
	"testing"

	"github.com/franchb/htmgo/framework/v2/hx"
	"github.com/stretchr/testify/assert"
)

func TestLifeCycle_OnEvent_htmx4Colon(t *testing.T) {
	t.Parallel()
	cases := []struct {
		event hx.Event
		want  string // the resulting key in l.handlers
	}{
		{hx.AfterSwapEvent, "hx-on::after:swap"},
		{hx.BeforeRequestEvent, "hx-on::before:request"},
		{hx.ConfigRequestEvent, "hx-on::config:request"},
		{hx.ErrorEvent, "hx-on::error"},
		{hx.ClickEvent, "onclick"}, // DOM event unchanged
	}
	for _, tc := range cases {
		l := NewLifeCycle().OnEvent(tc.event, SimpleJsCommand{Command: "noop"})
		_, ok := l.handlers[hx.Event(tc.want)]
		assert.True(t, ok, "event %q should map to handler key %q; got keys %v", tc.event, tc.want, l.handlers)
	}
}

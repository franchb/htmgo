package ax

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttributeConstants(t *testing.T) {
	t.Parallel()
	type c struct {
		got  Attribute
		want string
	}
	cases := []c{
		{DataAttr, "x-data"},
		{InitAttr, "x-init"},
		{ShowAttr, "x-show"},
		{BindAttr, "x-bind"},
		{OnAttr, "x-on"},
		{TextAttr, "x-text"},
		{HtmlAttr, "x-html"},
		{ModelAttr, "x-model"},
		{ModelableAttr, "x-modelable"},
		{CloakAttr, "x-cloak"},
		{RefAttr, "x-ref"},
		{IgnoreAttr, "x-ignore"},
		{TeleportAttr, "x-teleport"},
		{EffectAttr, "x-effect"},
		{IfAttr, "x-if"},
		{ForAttr, "x-for"},
		{IdAttr, "x-id"},
		{TransitionAttr, "x-transition"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, string(tc.got))
	}
}

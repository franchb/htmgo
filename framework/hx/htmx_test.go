package hx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringTrigger(t *testing.T) {
	trigger := "click once, htmx:click throttle:5, load delay:10"
	tgr := NewStringTrigger(trigger)
	assert.Equal(t, len(tgr.events), 3)
	assert.Equal(t, tgr.events[0].event, "click")
	assert.Equal(t, tgr.events[0].modifiers[0].Modifier(), "once")
	assert.Equal(t, tgr.events[1].event, "click")
	assert.Equal(t, tgr.events[1].modifiers[0].Modifier(), "throttle:5")
	assert.Equal(t, tgr.events[2].event, "load")
	assert.Equal(t, tgr.events[2].modifiers[0].Modifier(), "delay:10")
	assert.Equal(t, "click once, click throttle:5, load delay:10", tgr.ToString())
}

func TestEventConstants_htmx4(t *testing.T) {
	t.Parallel()
	// Use slice of pairs so that aliases sharing the same string value
	// (e.g. AfterOnLoadEvent / AfterProcessNodeEvent → "htmx:after:init")
	// can each be tested without duplicate-key compile errors.
	type eventCase struct {
		event Event
		want  string
	}
	cases := []eventCase{
		{AbortEvent, "htmx:abort"},
		{AfterOnLoadEvent, "htmx:after:init"},
		{AfterProcessNodeEvent, "htmx:after:init"},
		{AfterRequestEvent, "htmx:after:request"},
		{AfterSettleEvent, "htmx:after:swap"},
		{AfterSwapEvent, "htmx:after:swap"},
		{BeforeCleanupElementEvent, "htmx:before:cleanup"},
		{BeforeOnLoadEvent, "htmx:before:init"},
		{BeforeProcessNodeEvent, "htmx:before:process"},
		{BeforeRequestEvent, "htmx:before:request"},
		{BeforeSendEvent, "htmx:before:request"},
		{BeforeSwapEvent, "htmx:before:swap"},
		{ConfigRequestEvent, "htmx:config:request"},
		{BeforeHistorySaveEvent, "htmx:before:history:update"},
		{HistoryRestoreEvent, "htmx:before:history:restore"},
		{HistoryCacheMissEvent, "htmx:before:history:restore"},
		{PushedIntoHistoryEvent, "htmx:after:history:push"},
		{ErrorEvent, "htmx:error"},
		{ConfirmEvent, "htmx:confirm"},
		{PromptEvent, "htmx:prompt"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, string(tc.event))
	}
}

func TestHeaderConstants_htmx4(t *testing.T) {
	t.Parallel()
	type headerCase struct {
		got  Header
		want string
	}
	cases := []headerCase{
		{BoostedHeader, "HX-Boosted"},
		{RequestHeader, "HX-Request"},
		{TargetIdHeader, "HX-Target"},
		{TriggerIdHeader, "HX-Trigger"},
		{LocationHeader, "HX-Location"},
		{PushUrlHeader, "HX-Push-Url"},
		{RedirectHeader, "HX-Redirect"},
		{RefreshHeader, "HX-Refresh"},
		{ReplaceUrlHeader, "HX-Replace-Url"},
		{CurrentUrlHeader, "HX-Current-Url"},
		{ReswapHeader, "HX-Reswap"},
		{RetargetHeader, "HX-Retarget"},
		{ReselectHeader, "HX-Reselect"},
		{TriggerHeader, "HX-Trigger"},
		{SourceHeader, "HX-Source"},
		{RequestTypeHeader, "HX-Request-Type"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, string(c.got))
	}
}

func TestAttributeConstants_htmx4(t *testing.T) {
	t.Parallel()
	cases := []struct {
		got  Attribute
		want string
	}{
		{GetAttr, "hx-get"},
		{PostAttr, "hx-post"},
		{PutAttr, "hx-put"},
		{PatchAttr, "hx-patch"},
		{DeleteAttr, "hx-delete"},
		{PushUrlAttr, "hx-push-url"},
		{ReplaceUrlAttr, "hx-replace-url"},
		{SelectAttr, "hx-select"},
		{SelectOobAttr, "hx-select-oob"},
		{SwapAttr, "hx-swap"},
		{SwapOobAttr, "hx-swap-oob"},
		{TargetAttr, "hx-target"},
		{TriggerAttr, "hx-trigger"},
		{ValsAttr, "hx-vals"},
		{BoostAttr, "hx-boost"},
		{ConfirmAttr, "hx-confirm"},
		{IgnoreAttr, "hx-ignore"},   // htmx 4: was hx-disable in htmx 2
		{DisableAttr, "hx-disable"}, // htmx 4: was hx-disabled-elt in htmx 2
		{EncodingAttr, "hx-encoding"},
		{HeadersAttr, "hx-headers"},
		{IncludeAttr, "hx-include"},
		{IndicatorAttr, "hx-indicator"},
		{PreserveAttr, "hx-preserve"},
		{SyncAttr, "hx-sync"},
		{ValidateAttr, "hx-validate"},
		{ConfigAttr, "hx-config"},
		{StatusAttr, "hx-status"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, string(c.got))
	}
}

func TestToHtmxTriggerName_htmx4(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "after:swap", ToHtmxTriggerName("htmx:after:swap"))
	assert.Equal(t, "click", ToHtmxTriggerName("onclick"))
	assert.Equal(t, "custom-event", ToHtmxTriggerName("custom-event"))
}

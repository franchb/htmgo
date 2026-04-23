package hx

type Attribute = string
type Header = string
type Event = string
type SwapType = string

const (
	GetAttr        Attribute = "hx-get"
	PostAttr       Attribute = "hx-post"
	PutAttr        Attribute = "hx-put"
	PatchAttr      Attribute = "hx-patch"
	DeleteAttr     Attribute = "hx-delete"
	PushUrlAttr    Attribute = "hx-push-url"
	ReplaceUrlAttr Attribute = "hx-replace-url"
	SelectAttr     Attribute = "hx-select"
	SelectOobAttr  Attribute = "hx-select-oob"
	SwapAttr       Attribute = "hx-swap"
	SwapOobAttr    Attribute = "hx-swap-oob"
	TargetAttr     Attribute = "hx-target"
	TriggerAttr    Attribute = "hx-trigger"
	ValsAttr       Attribute = "hx-vals"
	BoostAttr      Attribute = "hx-boost"
	ConfirmAttr    Attribute = "hx-confirm"
	IgnoreAttr     Attribute = "hx-ignore"  // htmx 4: stops htmx processing within the element subtree (existed in htmx 2 with same semantics)
	DisableAttr    Attribute = "hx-disable" // htmx 4: disable form elements during in-flight request (REPURPOSED — in htmx 2, hx-disable disabled htmx processing; that role moved to hx-ignore-only and the v2 hx-disabled-elt semantic moved to v4 hx-disable)
	EncodingAttr   Attribute = "hx-encoding"
	HeadersAttr    Attribute = "hx-headers"
	IncludeAttr    Attribute = "hx-include"
	IndicatorAttr  Attribute = "hx-indicator"
	PreserveAttr   Attribute = "hx-preserve"
	SyncAttr       Attribute = "hx-sync"
	ValidateAttr   Attribute = "hx-validate"
	ConfigAttr     Attribute = "hx-config" // htmx 4: replaces hx-request
	StatusAttr     Attribute = "hx-status" // htmx 4: per-status swap/target control
)

const (
	BoostedHeader     Header = "HX-Boosted"
	RequestHeader     Header = "HX-Request"
	TargetIdHeader    Header = "HX-Target"    // htmx 4 format: tag#id
	TriggerIdHeader   Header = "HX-Trigger"   // htmx 4 format: tag#id
	LocationHeader    Header = "HX-Location"
	PushUrlHeader     Header = "HX-Push-Url"
	RedirectHeader    Header = "HX-Redirect"
	RefreshHeader     Header = "HX-Refresh"
	ReplaceUrlHeader  Header = "HX-Replace-Url"
	CurrentUrlHeader  Header = "HX-Current-Url"
	ReswapHeader      Header = "HX-Reswap"
	RetargetHeader    Header = "HX-Retarget"
	ReselectHeader    Header = "HX-Reselect"
	TriggerHeader     Header = "HX-Trigger"
	SourceHeader      Header = "HX-Source"       // new in htmx 4; tag#id format
	RequestTypeHeader Header = "HX-Request-Type" // new in htmx 4; "full" or "partial"
)

const (
	// Htmx Events (htmx 4 colon-form, verified against htmx.org/dist/htmx.js 4.0.0-beta2)
	AbortEvent                Event = "htmx:abort"
	AfterOnLoadEvent          Event = "htmx:after:init"
	AfterProcessNodeEvent     Event = "htmx:after:process"
	AfterRequestEvent         Event = "htmx:after:request"
	AfterSettleEvent          Event = "htmx:after:settle"
	AfterSwapEvent            Event = "htmx:after:swap"
	BeforeCleanupElementEvent Event = "htmx:before:cleanup"
	BeforeOnLoadEvent         Event = "htmx:before:init"
	BeforeProcessNodeEvent    Event = "htmx:before:process"
	BeforeRequestEvent        Event = "htmx:before:request"
	BeforeSendEvent           Event = "htmx:before:request" // alias for BeforeRequestEvent (htmx 4 consolidated beforeSend into beforeRequest)
	BeforeSwapEvent           Event = "htmx:before:swap"
	ConfigRequestEvent        Event = "htmx:config:request"
	BeforeHistorySaveEvent    Event = "htmx:before:history:update"
	HistoryRestoreEvent       Event = "htmx:before:history:restore"
	HistoryCacheMissEvent     Event = "htmx:before:history:restore" // Deprecated: htmx 4 removed localStorage history caching; this constant is retained as an alias for HistoryRestoreEvent and will be removed in a future release.
	PushedIntoHistoryEvent    Event = "htmx:after:history:push"
	ErrorEvent                Event = "htmx:error"
	ConfirmEvent              Event = "htmx:confirm"
	OnMutationErrorEvent      Event = "htmx:onMutationError" // htmgo-fork custom event; kept as-is (emitted by mutation-error extension)
	PromptEvent               Event = "htmx:prompt"

	// SSE (extracted to hx-sse.js in alpha8 — verified against htmx.org/dist/ext/hx-sse.js 4.0.0-beta2)
	SseConnectedEvent     Event = "htmx:after:sse:connection"
	SseConnectingEvent    Event = "htmx:before:sse:connection"
	SseClosedEvent        Event = "htmx:sse:close"
	SseErrorEvent         Event = "htmx:sse:error"
	SseBeforeMessageEvent Event = "htmx:before:sse:message"
	SseAfterMessageEvent  Event = "htmx:after:sse:message"
	SSEErrorEvent         Event = "htmx:sse:error"          // alias for historical SSEErrorEvent callsites
	SSEOpenEvent          Event = "htmx:after:sse:connection" // alias for historical SSEOpenEvent callsites

	// Misc Events (non-htmx)
	RevealedEvent   Event = "revealed"
	InstersectEvent Event = "intersect"
	PollingEvent    Event = "every"

	// Dom Events (unchanged)
	ClickEvent       Event = "onclick"
	ChangeEvent      Event = "onchange"
	InputEvent       Event = "oninput"
	FocusEvent       Event = "onfocus"
	BlurEvent        Event = "onblur"
	KeyDownEvent     Event = "onkeydown"
	KeyUpEvent       Event = "onkeyup"
	KeyPressEvent    Event = "onkeypress"
	SubmitEvent      Event = "onsubmit"
	LoadDomEvent     Event = "onload"
	LoadEvent        Event = "onload"
	UnloadEvent      Event = "onunload"
	ResizeEvent      Event = "onresize"
	ScrollEvent      Event = "onscroll"
	DblClickEvent    Event = "ondblclick"
	MouseOverEvent   Event = "onmouseover"
	MouseOutEvent    Event = "onmouseout"
	MouseMoveEvent   Event = "onmousemove"
	MouseDownEvent   Event = "onmousedown"
	MouseUpEvent     Event = "onmouseup"
	ContextMenuEvent Event = "oncontextmenu"
	DragStartEvent   Event = "ondragstart"
	DragEvent        Event = "ondrag"
	DragEnterEvent   Event = "ondragenter"
	DragLeaveEvent   Event = "ondragleave"
	DragOverEvent    Event = "ondragover"
	DropEvent        Event = "ondrop"
	DragEndEvent     Event = "ondragend"
)

const (
	SwapTypeTrue        SwapType = "true"
	SwapTypeInnerHtml   SwapType = "innerHTML"
	SwapTypeOuterHtml   SwapType = "outerHTML"
	SwapTypeTextContent SwapType = "textContent"
	SwapTypeBeforeBegin SwapType = "beforebegin"
	SwapTypeAfterBegin  SwapType = "afterbegin"
	SwapTypeBeforeEnd   SwapType = "beforeend"
	SwapTypeAfterEnd    SwapType = "afterend"
	SwapTypeDelete      SwapType = "delete"
	SwapTypeNone        SwapType = "none"
)

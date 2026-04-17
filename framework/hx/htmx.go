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
	IgnoreAttr     Attribute = "hx-ignore"  // htmx 4: stops htmx processing (was hx-disable in htmx 2)
	DisableAttr    Attribute = "hx-disable" // htmx 4: disable elements during request (was hx-disabled-elt in htmx 2)
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
	// Htmx Events (htmx 4 colon-form)
	AbortEvent                Event = "htmx:abort"
	AfterOnLoadEvent          Event = "htmx:after:init"
	AfterProcessNodeEvent     Event = "htmx:after:init"
	AfterRequestEvent         Event = "htmx:after:request"
	AfterSettleEvent          Event = "htmx:after:swap"
	AfterSwapEvent            Event = "htmx:after:swap"
	BeforeCleanupElementEvent Event = "htmx:before:cleanup"
	BeforeOnLoadEvent         Event = "htmx:before:init"
	BeforeProcessNodeEvent    Event = "htmx:before:process"
	BeforeRequestEvent        Event = "htmx:before:request"
	BeforeSendEvent           Event = "htmx:before:request"
	BeforeSwapEvent           Event = "htmx:before:swap"
	ConfigRequestEvent        Event = "htmx:config:request"
	BeforeHistorySaveEvent    Event = "htmx:before:history:update"
	HistoryRestoreEvent       Event = "htmx:before:history:restore"
	HistoryCacheMissEvent     Event = "htmx:before:history:restore"
	PushedIntoHistoryEvent    Event = "htmx:after:history:push"
	ErrorEvent                Event = "htmx:error"
	ConfirmEvent              Event = "htmx:confirm"
	OnMutationErrorEvent      Event = "htmx:onMutationError" // htmgo-fork custom event; kept as-is (emitted by mutation-error extension)
	PromptEvent               Event = "htmx:prompt"

	// SSE (extracted to hx-sse.js in alpha8 — names verified against upstream)
	SseConnectedEvent     Event = "htmx:sseOpen"
	SseConnectingEvent    Event = "htmx:sseConnecting"
	SseClosedEvent        Event = "htmx:sseClose"
	SseErrorEvent         Event = "htmx:sseError"
	SseBeforeMessageEvent Event = "htmx:sseBeforeMessage"
	SseAfterMessageEvent  Event = "htmx:sseAfterMessage"
	SSEErrorEvent         Event = "htmx:sseError" // alias for historical SSEErrorEvent callsites
	SSEOpenEvent          Event = "htmx:sseOpen"

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

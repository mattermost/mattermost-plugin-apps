package main

const (
	ManifestPath = "/manifest.json"
	StaticPath   = "/static"
	BindingsPath = "/bindings"
	InstallPath  = "/install"
	NotifyPath   = "/notify"

	// Commands
	CreateEmbedded  = "/create-embedded"
	Subscribe       = "/subscribe"
	Unsubscribe     = "/unsubscribe"
	CreateTimer     = "/timer/create"
	ExecuteTimer    = "/timer/execute"
	NumBindingsPath = "/num_bindings"

	// Submit responses
	OK               = "/ok"
	OKEmpty          = "/empty"
	NavigateInternal = "/nav/internal"
	NavigateExternal = "/nav/external"

	// Form responses
	FormSimple                    = "/forms/simple"
	FormSimpleSource              = "/forms/simpleSource"
	FormMarkdownError             = "/forms/markdownError"
	FormMarkdownErrorMissingField = "/forms/markdownMissingError"
	FormRefresh                   = "/forms/refresh"
	FormFull                      = "/forms/full"
	FormLookup                    = "/forms/lookup"
	FormFullSource                = "/forms/fullSource"
	FormFullReadonly              = "/forms/fullDisabled"
	FormMultiselect               = "/forms/multiselect"
	FormButtons                   = "/forms/buttons"

	// Lookup responses
	Lookup          = "/lookups/ok"
	LookupMultiword = "/lookups/multiword"
	LookupEmpty     = "/lookups/empty"

	// Error responses
	ErrorDefault                  = "/errors/default"
	ErrorEmpty                    = "/errors/empty"
	ErrorMarkdownForm             = "/errors/markdownform"
	ErrorMarkdownFormMissingField = "/errors/markdownformMissingField"
	Error404                      = "/errors/foo"
	Error500                      = "/errors/internal"

	// Invalid responses
	InvalidHTML        = "/invalid/html"
	InvalidUnknownType = "/invalid/unknown-type"
	InvalidLookup      = "/invalid/lookup"
	InvalidForm        = "/invalid/form"
	InvalidNavigate    = "/invalid/nav"
)

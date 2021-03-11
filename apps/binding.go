package apps

// Binding is the main way for an App to attach its functionality to the
// Mattermost UI.
//
// An App returns the bindings in response to the "bindings" call, that it must
// implement, and can customize in its Manifest. For each context in which it is
// invoked, the bindings call returns a tree of app's bindings, organized by the
// top-level location.
//
// Top level bindings need to define:
//  location - the top-level location to bind to, e.g. "post_menu".
//  bindings - an array of bindings
//
// /post_menu bindings need to define:
//  location - Name of this location. The whole path of locations will be added in the context.
//  icon - optional URL or path to the icon
//  label - Text to show in the item
//  call - Call to perform.
//
// /channel_header bindings need to define:
//  location - Name of this location. The whole path of locations will be added in the context.
//  icon - optional URL or path to the icon
//  label - text to show in the item on mobile and webapp collapsed view.
//  hint - text to show in the tooltip.
//  call - Call to perform.
//
// /command bindings can define "inner" subcommands that are collections of more
// bindings/subcommands, and "outer" subcommands that implement forms and can be
// executed. It is not possible to have command bindings that have subcommands
// and flags. It is possible to have positional parameters in an outer
// subcommand, accomplishing similar user experience.
//
// Inner command bindings need to define:
//  label - the label for the command itself.
//  location - the location of the command, defaults to label.
//  hint - Hint line in autocomplete.
//  description - description line in autocomplete.
//  bindings - subcommands
//
// Outer command bindings need to define:
//  label - the label for the command itself.
//  location - the location of the command, defaults to label.
//  hint - Hint line in autocomplete.
//  description - description line in autocomplete.
//  call or form - either embed a form, or provide a call to fetch it.
//
// Bindings are currently refreshed when a user visits a channel, in the context
// of the current channel, from all the registered Apps. A server-side cache
// implementation is in the works. TODO ticket ref. This allows each App to
// dynamically add things to the UI on a per-channel basis.
//
// Example bindings (hello world app) create a button in the channel header, and
// a "/helloworld send" command:
//   {
//      "type": "ok",
//      "data": [
//          {
//              "location": "/channel_header",
//              "bindings": [
//                  {
//                      "location": "send-button",
//                      "icon": "http://localhost:8080/static/icon.png",
//                      "call": {
//                          "path": "/send-modal"
//                      }
//                  }
//              ]
//          },
//          {
//              "location": "/command",
//              "bindings": [
//                  {
//                      "location": "send",
//                      "label": "send",
//                      "call": {
//                          "path": "/send"
//                      }
//                  }
//              ]
//          }
//      ]
//   }
type Binding struct {
	// For internal use by Mattermost, Apps do not need to set.
	AppID AppID `json:"app_id,omitempty"`

	// Location allows the App to identify where in the UX the Call request
	// comes from. It is optional. For /command bindings, Location is
	// defaulted to Label.
	//
	// TODO: default to Label, Name.
	Location Location `json:"location,omitempty"`

	// Icon is the icon to display, should be either a fully-qualified URL, or a
	// path for an app's static asset.
	Icon string `json:"icon,omitempty"`

	// Label is the (usually short) primary text to display at the location.
	//  - post menu: the item text.
	//  - channel header: the dropdown text.
	//  - command: the name of the command.
	//  - in-post: the title.
	Label string `json:"label,omitempty"`

	// Hint is the secondary text to display
	//  - post menu: not used
	//  - channel header: tooltip
	//  - command: the "Hint" line
	Hint string `json:"hint,omitempty"`

	// Description is the (optional) extended help text, used in modals and
	// autocomplete.
	//  - in-post: is the text of the embed
	Description string `json:"description,omitempty"`

	// RoleID is a role required to see the item (hidden for other users).
	RoleID string `json:"role_id,omitempty"`

	// DependsOnTeam, etc. specifies the scope of the binding and how it can be
	// shared across various user sessions.
	DependsOnTeam    bool `json:"depends_on_team,omitempty"`
	DependsOnChannel bool `json:"depends_on_channel,omitempty"`
	DependsOnUser    bool `json:"depends_on_user,omitempty"`
	DependsOnPost    bool `json:"depends_on_post,omitempty"`

	// A Binding is either to a Call, or is a "container" for other locations -
	// i.e. menu sub-items or subcommands. An app-defined Modal can be displayed
	// by setting AsModal.
	Call     *Call      `json:"call,omitempty"`
	Bindings []*Binding `json:"bindings,omitempty"`

	// Form allows to embed a form into a binding, and avoid the need to
	// Call(type=Form). At the moment, the sole use case is in-post forms, but
	// this may prove useful in other contexts.
	// TODO: Can embedded forms be mutable, and what does it mean?
	Form *Form `json:"form,omitempty"`
}

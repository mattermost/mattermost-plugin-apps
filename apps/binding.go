package apps

// Binding is the main way for an App to attach its functionality to the
// Mattermost UI. An App returns the bindings in response to the "bindings"
// call, that it can customize in its Manifest. App's bindings can be identical,
// or differ for users, channels, and such.
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

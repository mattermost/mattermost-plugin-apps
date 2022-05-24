// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package apps

// Binding is the principal way for an App to attach its functionality to the
// Mattermost UI.  An App can bind to top-level UI elements by implementing the
// (mandatory) "bindings" call. It can also add bindings to messages (posts) in
// Mattermost, by setting "app_bindings" property of the posts.
//
// Mattermost UI Bindings
//
// An App returns its bindings in response to the "bindings" call, that it must
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
//  hint - text to show in the webapp's tooltip.
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
// implementation is in the works,
// https://mattermost.atlassian.net/browse/MM-30472. This allows each App to
// dynamically add things to the UI on a per-channel basis.
//
// Example bindings (hello world app) create a button in the channel header, and
// a "/helloworld send" command:
//  {
//      "type": "ok",
//      "data": [
//          {
//              "location": "/channel_header",
//              "bindings": [
//                  {
//                      "location": "send-button",
//                      "icon": "icon.png",
//                      "label":"send hello message",
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
//                      "icon": "icon.png",
//                      "description": "Hello World app",
//                      "hint":        "[send]",
//                      "bindings": [
//                          {
//                              "location": "send",
//                              "label": "send",
//                              "call": {
//                                  "path": "/send"
//                              }
//                          }
//                      ]
//                  }
//              ]
//          }
//      ]
//  }
//
// In-post Bindings
//
// An App can also create messages (posts) in Mattermost with in-post bindings.
// To do that, it invokes `CreatePost` REST API, and sets the "app_bindings"
// prop, as in the following example:
//
// In post bindings are embedded into posts, and are not registered like the
// rest of the bindings. In order to make in post bindings appear, you must
// create a post with an apps_bindings property. You can add several bindings to
// a post, and each one will appear as a single attachment to the post. Top
// level bindings will define the attachment contents, with the Label becoming
// the title, and the Description becoming the body of the attachment.
// Sub-bindings will become the actions of the post. An action with no
// sub-bindings will be rendered as a button. An action with sub-bindings will
// be rendered as a select. You can identify the action triggered by the
// location.
//
//  {
//     channel_id: "channelID",
//     message: "Some message to appear before the attachment",
//     props: {
//         app_bindings: [{
//             app_id: "my app id",
//             location: "location",
//             label: "title of the attachment",
//             description: "body of the attachment",
//             bindings: [
//                 {
//                     location: "my_select",
//                     label: "Placeholder text",
//                     bindings: [
//                         {
//                             location: "option1",
//                             label: "Option 1",
//                         }, {
//                             location: "option2",
//                             label: "Option 2",
//                         },
//                     ],
//                     call: {
//                         path: "my/path",
//                     },
//                 }, {
//                     location: "my_button",
//                     label: "Button label",
//                     call: {
//                         path: "my/path",
//                     },
//                 },
//             ],
//         }],
//     },
//  }
//
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

	// A Binding is either an action, a form (embedded or fetched), or a
	// "container" for other locations/bindings - i.e. menu sub-items or
	// subcommands. An attempt to specify more than one of these fields is
	// treated as an error.

	// Submit is used to execute the action associated to this binding.
	Submit *Call `json:"submit,omitempty"`

	// Form is used to gather additional input from the user before submitting.
	// At a minimum, it contains the Submit call path or a Source call path. A
	// form may be embedded, or be a source reference meaning that a call to
	// the app will be made to obtain the form.
	Form *Form `json:"form,omitempty"`

	// Bindings specifies sub-location bindings.
	Bindings []Binding `json:"bindings,omitempty"`
}

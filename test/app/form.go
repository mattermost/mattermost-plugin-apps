package main

import (
	"github.com/gorilla/mux"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func formCommandBinding(_ apps.Context) apps.Binding {
	return apps.Binding{
		Label: "form",
		Bindings: []apps.Binding{
			newBinding("buttons", FormButtons),
			newBinding("full-readonly", FormFullReadonly),
			newBinding("full", FormFull),
			newBinding("lookup", FormLookup),
			newBinding("markdown-error-missing-field", FormMarkdownErrorMissingField),
			newBinding("markdown-error", FormMarkdownError),
			newBinding("multiselect", FormMultiselect),
			newBinding("refresh", FormRefresh),
			newBinding("simple", FormSimple),
		},
	}
}

func initHTTPForms(r *mux.Router) {
	handleCall(r, FormButtons, handleFormButtons)
	handleCall(r, FormFull, handleForm(fullForm))
	handleCall(r, FormFullReadonly, handleForm(fullFormReadonly()))
	handleCall(r, FormFullSource, handleForm(simpleFormSource))
	handleCall(r, InvalidForm, handleForm(apps.Form{}))
	handleCall(r, FormLookup, handleForm(lookupForm))
	handleCall(r, FormMarkdownError, handleForm(formWithMarkdownError))
	handleCall(r, FormMarkdownErrorMissingField, handleForm(formWithMarkdownErrorMissingField))
	handleCall(r, FormMultiselect, handleForm(formMultiselect))
	handleCall(r, FormRefresh, handleFormRefresh)
	handleCall(r, FormSimple, handleForm(simpleForm))
	handleCall(r, FormSimpleSource, handleForm(simpleFormSource))
}

var simpleForm = apps.Form{
	Title:  "Simple Form",
	Submit: callOK,
	Fields: []apps.Field{
		{
			Type: apps.FieldTypeText,
			Name: "test_field",
		},
	},
}

var simpleFormSource = apps.Form{
	Source: apps.NewCall(FormSimple),
}

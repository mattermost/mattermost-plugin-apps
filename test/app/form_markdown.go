package main

import (
	"github.com/mattermost/mattermost-plugin-apps/apps"
)

const fullMarkdown = "## Markdown title" +
	"\nHello world" +
	"\nText styles: _italics_ **bold** **_bold-italic_** ~~strikethrough~~ `code`" +
	"\nUsers and channels: @sysadmin ~town-square" +
	"\n```" +
	"\nCode block" +
	"\n```" +
	"\n:+1: :banana_dance:" +
	"\n***" +
	"\n> Quote\n" +
	"\nLink: [here](www.google.com)" +
	"\nImage: ![img](https://gdm-catalog-fmapi-prod.imgix.net/ProductLogo/4acbc64f-552d-4944-8474-b44a13a7bd3e.png?auto=format&q=50&fit=fill)" +
	"\nList:" +
	"\n- this" +
	"\n- is" +
	"\n- a" +
	"\n- list" +
	"\nNumbered list" +
	"\n1. this" +
	"\n2. is" +
	"\n3. a" +
	"\n4. list" +
	"\nItems" +
	"\n- [ ] Item one" +
	"\n- [ ] Item two" +
	"\n- [x] Completed item"

func formMarkdownError(errorPath string) apps.Form {
	return apps.Form{
		Title:  "Test markdown descriptions and errors",
		Header: "Test header",
		Submit: apps.NewCall(errorPath),
		Fields: []apps.Field{
			{
				Name:  "static",
				Type:  apps.FieldTypeStaticSelect,
				Label: "static",
				Description: `| Option | Message  | Image |
| :------------ |:---------------:| -----:|
| Opt1 | You are good     |  :smile: |
| Opt2 | You are awesome              | :+1: |
| Opt3| You are great       |    :smirk:  |`,
				SelectIsMulti: true,
				SelectStaticOptions: []apps.SelectOption{
					{
						Label: "button1",
						Value: "button1",
					},
					{
						Label: "button2",
						Value: "button2",
					},
					{
						Label: "button3",
						Value: "button3",
					},
					{
						Label: "button4",
						Value: "button4",
					},
				},
			},
			{
				Name:        "text",
				Type:        apps.FieldTypeText,
				Label:       "text",
				Description: fullMarkdown, // "Go [here](www.google.com) for more information.",
			},
			{
				Name:  "boolean",
				Type:  apps.FieldTypeBool,
				Label: "boolean",
				Description: `Mark this field only if:
					1. You want
					2. You need
					3. You should`,
			},
		},
	}
}

var formWithMarkdownError = formMarkdownError(ErrorMarkdownForm)
var formWithMarkdownErrorMissingField = formMarkdownError(ErrorMarkdownFormMissingField)

func handleErrorMarkdownForm(_ *apps.CallRequest) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeError,
		Text: "## This is a very **BIG** error.\nYou should probably take a look at it.",
		Data: map[string]map[string]string{
			"errors": {
				"text":    "These are not the emojis you are looking for :sweat_smile:",
				"boolean": "Are you sure you should _mark_ this field?",
				"static":  "## Careful\nThis is an error.",
				"missing": "Some missing field.",
			},
		},
	}
}

func handleErrorMarkdownFormMissingField(_ *apps.CallRequest) apps.CallResponse {
	return apps.CallResponse{
		Type: apps.CallResponseTypeError,
		Data: map[string]map[string]string{
			"errors": {
				"missing": "Some missing field.",
			},
		},
	}
}

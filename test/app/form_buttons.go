package main

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func handleFormButtons(creq *apps.CallRequest) apps.CallResponse {
	// saved as int, but comes back as float64 in go json
	numButtonsFloat, _ := creq.State.(float64)
	numButtons := int(numButtonsFloat)

	switch creq.GetValue("submit", "") {
	case "":
		// initial state, display with numButtons (== 0)
	case "add_buttons":
		numButtons++
	case "error":
		return apps.NewErrorResponse(errors.New("you caused an error :)"))
	default:
		return handleOK(creq)
	}

	buttons := []apps.SelectOption{
		{
			Label: "add buttons",
			Value: "add_buttons",
		},
		{
			Label: "error",
			Value: "error",
		},
	}
	for i := 0; i < numButtons; i++ {
		buttons = append(buttons, apps.SelectOption{
			Label: fmt.Sprintf("button%v", i),
			Value: fmt.Sprintf("button%v", i),
		})
	}

	return apps.NewFormResponse(apps.Form{
		Title:         "Test multiple buttons Form",
		Header:        "Test header",
		SubmitButtons: "submit",
		Submit: &apps.Call{
			Path:  FormButtons,
			State: numButtons,
		},
		Fields: []apps.Field{
			{
				Name:                "submit",
				Type:                apps.FieldTypeStaticSelect,
				Label:               "static",
				SelectStaticOptions: buttons,
			},
		},
	})
}

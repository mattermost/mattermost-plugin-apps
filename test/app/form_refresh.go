package main

import (
	"fmt"
	"strconv"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

func handleFormRefresh(creq *apps.CallRequest) apps.CallResponse {
	n, _ := strconv.ParseUint(creq.GetValue("number", ""), 10, 32)
	numOption := func(nBoxes uint64) apps.SelectOption {
		if n == 1 {
			return apps.SelectOption{
				Label: "1 box",
				Value: "1",
			}
		}

		return apps.SelectOption{
			Label: fmt.Sprintf("%v boxes", nBoxes),
			Value: fmt.Sprintf("%v", nBoxes),
		}
	}
	fieldNumber := apps.Field{
		Name:          "number",
		ModalLabel:    "Number of checks",
		Type:          apps.FieldTypeStaticSelect,
		SelectRefresh: true,
		SelectStaticOptions: []apps.SelectOption{
			numOption(1),
			numOption(3),
			numOption(5),
		},
	}

	if n != 0 {
		fieldNumber.Value = numOption(n)
	}

	m, _ := strconv.ParseUint(creq.GetValue("multiplier", ""), 10, 32)
	multiOption := func(n uint64) apps.SelectOption {
		switch n {
		case 1:
			return apps.SelectOption{Label: "unchanged", Value: "1"}
		case 2:
			return apps.SelectOption{Label: "double", Value: "2"}
		case 3:
			return apps.SelectOption{Label: "triple", Value: "3"}
		default:
			return apps.SelectOption{Label: "ERROR", Value: "ERROR"}
		}
	}
	fieldMultiplier := apps.Field{
		Name:          "multiplier",
		ModalLabel:    "Multiplier for number",
		Type:          apps.FieldTypeStaticSelect,
		SelectRefresh: true,
		SelectStaticOptions: []apps.SelectOption{
			multiOption(1),
			multiOption(2),
			multiOption(3),
		},
	}

	if m != 0 {
		fieldMultiplier.Value = multiOption(m)
	}

	fields := []apps.Field{
		fieldNumber,
		fieldMultiplier,
	}
	for i := uint64(0); i < n*m; i++ {
		fields = append(fields, apps.Field{
			Name:  fmt.Sprintf("box%v", i),
			Label: fmt.Sprintf("box%v", i),
			Type:  apps.FieldTypeBool,
		})
	}

	return apps.NewFormResponse(apps.Form{
		Title:  "Test Refresh Form",
		Header: "Test header",
		Submit: callOK,
		Source: apps.NewCall(FormRefresh),
		Fields: fields,
	})
}

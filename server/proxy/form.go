// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
)

// cleanForm removes:
// - Fields without a name
// - Fields with labels (either natural or defaulted from names) with more than one word
// - Fields that have the same label as previous fields
// - Invalid select static fields and their invalid options
func cleanForm(in apps.Form) (apps.Form, []error) {
	out := in
	out.Fields = []apps.Field{}
	problems := []error{}
	usedLabels := map[string]bool{}
	for _, f := range in.Fields {
		if f.Name == "" {
			problems = append(problems, errors.Errorf("field with no name, label %s", f.Label))
			continue
		}
		if strings.ContainsAny(f.Name, " \t") {
			problems = append(problems, errors.Errorf("field name must be a single word: %q", f.Name))
			continue
		}

		label := f.Label
		if label == "" {
			label = f.Name
		}
		if strings.ContainsAny(label, " \t") {
			problems = append(problems, errors.Errorf("label must be a single word: %q (field: %s)", label, f.Name))
			continue
		}

		if usedLabels[label] {
			problems = append(problems, errors.Errorf("repeated label: %q (field: %s)", label, f.Name))
			continue
		}

		if f.Type == apps.FieldTypeStaticSelect {
			clean, ee := cleanStaticSelect(f)
			problems = append(problems, ee...)
			if len(clean.SelectStaticOptions) == 0 {
				problems = append(problems, errors.Errorf("no options for static select: %s", f.Name))
				continue
			}
			f = clean
		}

		out.Fields = append(out.Fields, f)
		usedLabels[label] = true
	}

	return out, problems
}

// cleanStaticSelect removes:
// - Options with empty label (either natural or defaulted form the value)
// - Options that have the same label as the previous options
// - Options that have the same value as the previous options
func cleanStaticSelect(f apps.Field) (apps.Field, []error) {
	problems := []error{}
	usedLabels := map[string]bool{}
	usedValues := map[string]bool{}
	clean := []apps.SelectOption{}
	for _, option := range f.SelectStaticOptions {
		label := option.Label
		if label == "" {
			label = option.Value
		}
		if label == "" {
			problems = append(problems, errors.Errorf("option with neither label nor value (field %s)", f.Name))
			continue
		}

		if usedLabels[label] {
			problems = append(problems, errors.Errorf("repeated label %q on select option (field %s)", label, f.Name))
			continue
		}

		if usedValues[option.Value] {
			problems = append(problems, errors.Errorf("repeated value %q on select option (field %s)", option.Value, f.Name))
			continue
		}

		usedLabels[label] = true
		usedValues[option.Value] = true
		clean = append(clean, option)
	}

	f.SelectStaticOptions = clean
	return f, problems
}

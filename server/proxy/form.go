// Copyright (c) 2020-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package proxy

import (
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-apps/apps"
	"github.com/mattermost/mattermost-plugin-apps/server/config"
)

// cleanForm removes:
// - Fields without a name
// - Fields with labels (either natural or defaulted from names) with more than one word
// - Fields that have the same label as previous fields
// - Invalid select static fields and their invalid options
func cleanForm(in apps.Form, conf config.Config, appID apps.AppID) (apps.Form, error) {
	out := in
	out.Fields = []apps.Field{}
	var problems error
	usedLabels := map[string]bool{}

	if in.Icon != "" {
		icon, err := normalizeStaticPath(conf, appID, in.Icon)
		if err != nil {
			problems = multierror.Append(problems, errors.Wrap(err, "invalid icon path in form"))
			out.Icon = ""
		} else {
			out.Icon = icon
		}
	}

	if in.Submit == nil && in.Source == nil {
		problems = multierror.Append(problems, errors.New("form must define either a submit or a source"))
	}

	for _, f := range in.Fields {
		if f.Name == "" {
			problems = multierror.Append(problems, errors.Errorf("field with no name, label %s", f.Label))
			continue
		}
		if strings.ContainsAny(f.Name, " \t") {
			problems = multierror.Append(problems, errors.Errorf("field name must be a single word: %q", f.Name))
			continue
		}

		if f.Label == "" {
			f.Label = strings.ReplaceAll(f.Name, "_", "-")
		}
		if strings.ContainsAny(f.Label, " \t") {
			problems = multierror.Append(problems, errors.Errorf("label must be a single word: %q (field: %s)", f.Label, f.Name))
			continue
		}

		if usedLabels[f.Label] {
			problems = multierror.Append(problems, errors.Errorf("repeated label: %q (field: %s)", f.Label, f.Name))
			continue
		}

		switch f.Type {
		case apps.FieldTypeStaticSelect:
			clean, ee := cleanStaticSelect(f)
			if ee != nil {
				problems = multierror.Append(problems, ee)
			}
			if len(clean.SelectStaticOptions) == 0 {
				problems = multierror.Append(problems, errors.Errorf("no options for static select: %s", f.Name))
				continue
			}
			f = clean
		case apps.FieldTypeDynamicSelect:
			if f.SelectDynamicLookup == nil {
				problems = multierror.Append(problems, errors.Errorf("no lookup call for dynamic select: %s", f.Name))
				continue
			}
		}

		out.Fields = append(out.Fields, f)
		usedLabels[f.Label] = true
	}

	return out, problems
}

// cleanStaticSelect removes:
// - Options with empty label (either natural or defaulted form the value)
// - Options that have the same label as the previous options
// - Options that have the same value as the previous options
func cleanStaticSelect(f apps.Field) (apps.Field, error) {
	var problems error
	usedLabels := map[string]bool{}
	usedValues := map[string]bool{}
	clean := []apps.SelectOption{}
	for _, option := range f.SelectStaticOptions {
		label := option.Label
		if label == "" {
			label = option.Value
		}
		if label == "" {
			problems = multierror.Append(problems, errors.Errorf("option with neither label nor value (field %s)", f.Name))
			continue
		}

		if usedLabels[label] {
			problems = multierror.Append(problems, errors.Errorf("repeated label %q on select option (field %s)", label, f.Name))
			continue
		}

		if usedValues[option.Value] {
			problems = multierror.Append(problems, errors.Errorf("repeated value %q on select option (field %s)", option.Value, f.Name))
			continue
		}

		usedLabels[label] = true
		usedValues[option.Value] = true
		clean = append(clean, option)
	}

	f.SelectStaticOptions = clean
	return f, problems
}

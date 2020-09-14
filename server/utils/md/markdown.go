// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package md

import (
	"encoding/json"
	"fmt"
	"strings"
)

type MD string

var _ Markdowner = MD("")

type Markdowner interface {
	fmt.Stringer
	Markdown() MD
}

func (md MD) Markdown() MD   { return md }
func (md MD) String() string { return string(md) }

func Markdownf(format string, args ...interface{}) MD {
	return MD(fmt.Sprintf(format, args...))
}

func JSON(ref interface{}) MD {
	bb, _ := json.MarshalIndent(ref, "", "  ")
	return Markdownf(string(bb))
}

func CodeBlock(in string) MD {
	return Markdownf("\n```\n%s\n```\n", in)
}

func JSONBlock(ref interface{}) MD {
	return Markdownf("\n```json\n%s\n```\n", JSON(ref))
}

func Indent(in Markdowner, prefix string) MD {
	lines := strings.Split(in.String(), "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return Markdownf(strings.Join(lines, "\n"))
}

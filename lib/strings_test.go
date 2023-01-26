// Copyright (c) 2023 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the ISC license

package lib

import (
	"testing"
)

var replacementTestCases = []struct {
	input  string
	output string
}{
	{"", ""},
	{`a`, "a"},
	{`[`, "["},
	{`[[`, "[["},
	{`[[a]]`, "[[a]]"},
	{`a[[CTCP]]b`, "a\x01b"},
	{`[[B]]b[[B]]`, "\x02b\x02"},
	{`a[[\x67]]b`, "a\x67b"},
	{`[[\x67]]`, "\x67"},
	{`www[[\x00]]`, "www\x00"},
	{`www[[\x0D]]`, "www\x0d"},
	{`www[[\xff]]`, "www\xff"},
	{`www[[\xFF]]`, "www\xff"},
	{`[[notanescape]]`, "[[notanescape]]"},
	{`[[[U]]]`, "[\x1f]"},
}

func TestReplaceControlCodes(t *testing.T) {
	for _, testCase := range replacementTestCases {
		actual := ReplaceControlCodes(testCase.input)
		if actual != testCase.output {
			t.Errorf("expected `%s` -> %#v, got %#v", testCase.input, []byte(testCase.output), []byte(actual))
		}
	}
}

// Copyright (c) 2023 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the ISC license

package lib

import (
	"fmt"
	"testing"
)

type stringTestCase struct {
	input  string
	output string
}

func runTestCases(t *testing.T, cases []stringTestCase, converter func(string) string, escape func(string) string) {
	if escape == nil {
		escape = func(in string) string {
			return fmt.Sprintf("%#v", []byte(in))
		}
	}
	t.Helper()
	for i, testCase := range cases {
		actual := converter(testCase.input)
		if actual != testCase.output {
			t.Errorf("test case %d failed: input `%s`\nwant `%s`\ngot  `%s`",
				i, testCase.input, escape(testCase.output), escape(actual))
		}
	}
}

// tests for ReplaceControlCodes
var replacementTestCases = []stringTestCase{
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
	{`www[[\xFF\x00]]`, "www\xff\x00"},
	{`www[[\xFF\x00\x01]]01`, "www\xff\x00\x0101"},
	{`www[[\xaa\xBB\xcc\x35]]01`, "www\xaa\xbb\xcc\x3501"},
	{`www[[\xzz]]`, "www[[\\xzz]]"}, // invalid hex is not an escape
	{`[[notanescape]]`, "[[notanescape]]"},
	{`[[[U]]]`, "[\x1f]"},
}

func TestReplaceControlCodes(t *testing.T) {
	runTestCases(t, replacementTestCases, ReplaceControlCodes, nil)
}

var ansi16MessageTestCases = []stringTestCase{
	{"", ""},
	{"\x01ACTION snorts\x01", ctcpMarker + "ACTION snorts" + ctcpMarker},
	{`a`, "a"},
	{"\x1ea", "\x1b[9ma\x1b[0m"},
	{"a \x0302blue text", "a \x1b[34mblue text\x1b[0m"},
	{"a \x0302\x02bold blue text", "a \x1b[1;34mbold blue text\x1b[0m"},
	{"a \x0302blue text \x02with a bold portion", "a \x1b[34mblue text \x1b[0m\x1b[1;34mwith a bold portion\x1b[0m"},
	{"a \x0372blue text", "a blue text"},
}

func TestAnsi16MessageConversion(t *testing.T) {
	converter := func(in string) string {
		return IRCMessageToAnsi(in, ColorLevelBasic, false)
	}
	runTestCases(t, ansi16MessageTestCases, converter, ansiDebugEscape)
}

var ansi256MessageTestCases = []stringTestCase{
	{"", ""},
	{"\x01ACTION snorts\x01", ctcpMarker + "ACTION snorts" + ctcpMarker},
	{`a`, "a"},
	{"a \x0302blue text", "a \x1b[34mblue text\x1b[0m"},
	{"a \x0302\x02bold blue text", "a \x1b[1;34mbold blue text\x1b[0m"},
	{"a \x0372blue text", "a \x1b[38;5;63mblue text\x1b[0m"},
	{"a \x0372,51blue text on a red background", "a \x1b[38;5;63;48;5;161mblue text on a red background\x1b[0m"},
}

func TestAnsi256MessageConversion(t *testing.T) {
	converter := func(in string) string {
		return IRCMessageToAnsi(in, ColorLevelAnsi256, false)
	}
	runTestCases(t, ansi256MessageTestCases, converter, ansiDebugEscape)
}

var ansi16LineTestCases = []stringTestCase{
	{"", ""},
	{":", ":"},
	{" ", " "},
	{"  ", "  "},
	{" x  ", " x  "},
	{"PING x", "PING x"},
	{"\x02hi\x02", "\x02hi\x02"},
	// final parameter is converted even if it's something weird like PING
	{"PING \x02hi", "PING \x1b[1mhi\x1b[0m"},
	{"@draft+x=\x02hi\x02 :invalid\x02source!a@b PRIV\x02MSG\x02 #chan\x02nel \x02boldface", "@draft+x=\x02hi\x02 :invalid\x02source!a@b PRIV\x02MSG\x02 #chan\x02nel \x1b[1mboldface\x1b[0m"},
	{
		"@draft+x=\x02hi\x02 :invalid\x02source!a@b   PRIV\x02MSG\x02 #chan\x02nel    \x02boldface",
		"@draft+x=\x02hi\x02 :invalid\x02source!a@b   PRIV\x02MSG\x02 #chan\x02nel    \x1b[1mboldface\x1b[0m",
	},
	{
		":invalid\x02source!a@b   PRIV\x02MSG\x02 #chan\x02nel    \x02boldface",
		":invalid\x02source!a@b   PRIV\x02MSG\x02 #chan\x02nel    \x1b[1mboldface\x1b[0m",
	},
	{
		"PRIV\x02MSG\x02 #chan\x02nel    \x02boldface",
		"PRIV\x02MSG\x02 #chan\x02nel    \x1b[1mboldface\x1b[0m",
	},
	{
		"PRIV\x02MSG\x02    \x02boldface",
		"PRIV\x02MSG\x02    \x1b[1mboldface\x1b[0m",
	},
	// this is not recognized as a final parameter
	{
		":PRIV\x02MSG\x02    \x02boldface",
		":PRIV\x02MSG\x02    \x02boldface",
	},
	{"PRIVMSG #chat :\x01ACTION snorts\x01", "PRIVMSG #chat :" + ctcpMarker + "ACTION snorts" + ctcpMarker},
	{"PRIVMSG #chat :", "PRIVMSG #chat :"},
}

func TestAnsi16LineConversions(t *testing.T) {
	converter := func(in string) string {
		return IRCLineToAnsi(in, ColorLevelBasic, false)
	}
	runTestCases(t, ansi16LineTestCases, converter, ansiDebugEscape)
}

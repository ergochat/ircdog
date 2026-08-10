// Copyright (c) 2026 Shivaram Lingamneni <slingamn@cs.stanford.edu>
// released under the ISC license

package lib

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadScript(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "test_script.txt")

	tests := []struct {
		name     string
		content  string
		expected []ScriptCommand
	}{
		{
			name:     "empty file",
			content:  "",
			expected: []ScriptCommand{},
		},
		{
			name:    "single message",
			content: "NICK alice",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
			},
		},
		{
			name:    "multiple messages",
			content: "NICK alice\nUSER alice 0 * :Alice\nJOIN #test",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
				{Type: ScriptMessage, Message: "JOIN #test"},
			},
		},
		{
			name:    "empty lines ignored",
			content: "NICK alice\n\n\nUSER alice 0 * :Alice",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
			},
		},
		{
			name:    "comments ignored",
			content: "# This is a comment\nNICK alice\n# Another comment\nUSER alice 0 * :Alice",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
			},
		},
		{
			name:    "leading whitespace trimmed",
			content: "  NICK alice\n\t\tUSER alice 0 * :Alice\n  \t  JOIN #test",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
				{Type: ScriptMessage, Message: "JOIN #test"},
			},
		},
		{
			name:    "trailing whitespace preserved",
			content: "NICK alice  \nUSER alice 0 * :Alice\t\t\nJOIN #test   ",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice  "},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice\t\t"},
				{Type: ScriptMessage, Message: "JOIN #test   "},
			},
		},
		{
			name:    "sleep with duration format",
			content: "*SLEEP 1s\n*sleep 100ms\n*Sleep 1m30s",
			expected: []ScriptCommand{
				{Type: ScriptSleep, Sleep: 1 * time.Second},
				{Type: ScriptSleep, Sleep: 100 * time.Millisecond},
				{Type: ScriptSleep, Sleep: 90 * time.Second},
			},
		},
		{
			name:    "sleep with float format",
			content: "*SLEEP 1.5\n*sleep 0.1\n*Sleep 2",
			expected: []ScriptCommand{
				{Type: ScriptSleep, Sleep: time.Duration(1.5 * float64(time.Second))},
				{Type: ScriptSleep, Sleep: time.Duration(0.1 * float64(time.Second))},
				{Type: ScriptSleep, Sleep: 2 * time.Second},
			},
		},
		{
			name:    "sleep with extra whitespace",
			content: "*SLEEP   1s\n*sleep\t100ms",
			expected: []ScriptCommand{
				{Type: ScriptSleep, Sleep: 1 * time.Second},
				{Type: ScriptSleep, Sleep: 100 * time.Millisecond},
			},
		},
		{
			name:     "invalid sleep commands ignored",
			content:  "*SLEEP\n*SLEEP invalid\n*NOTASLEEP 1s",
			expected: []ScriptCommand{},
		},
		{
			name: "mixed content",
			content: `# Script with mixed content
NICK alice
  # Another comment
USER alice 0 * :Alice

*SLEEP 1s
JOIN #test
*sleep 0.5
PRIVMSG #test :Hello, world!`,
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
				{Type: ScriptSleep, Sleep: 1 * time.Second},
				{Type: ScriptMessage, Message: "JOIN #test"},
				{Type: ScriptSleep, Sleep: time.Duration(0.5 * float64(time.Second))},
				{Type: ScriptMessage, Message: "PRIVMSG #test :Hello, world!"},
			},
		},
		{
			name:    "windows line endings",
			content: "NICK alice\r\nUSER alice 0 * :Alice\r\n",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "NICK alice"},
				{Type: ScriptMessage, Message: "USER alice 0 * :Alice"},
			},
		},
		{
			name:    "SASL PLAIN",
			content: "*saslplain shivaram honeycomb",
			expected: []ScriptCommand{
				{Type: ScriptMessage, Message: "AUTHENTICATE c2hpdmFyYW0Ac2hpdmFyYW0AaG9uZXljb21i"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(scriptPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			result, err := ReadScript(scriptPath)
			if err != nil {
				t.Fatalf("ReadScript() error = %v", err)
			}

			if len(result) != len(tt.expected) {
				t.Fatalf("ReadScript() returned %d commands, expected %d", len(result), len(tt.expected))
			}

			for i, cmd := range result {
				if cmd.Type != tt.expected[i].Type {
					t.Errorf("Command %d: Type = %v, expected %v", i, cmd.Type, tt.expected[i].Type)
				}
				if cmd.Type == ScriptMessage && cmd.Message != tt.expected[i].Message {
					t.Errorf("Command %d: Message = %q, expected %q", i, cmd.Message, tt.expected[i].Message)
				}
				if cmd.Type == ScriptSleep && cmd.Sleep != tt.expected[i].Sleep {
					t.Errorf("Command %d: Sleep = %v, expected %v", i, cmd.Sleep, tt.expected[i].Sleep)
				}
			}
		})
	}
}

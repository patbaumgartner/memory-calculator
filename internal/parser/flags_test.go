package parser

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "Single flag",
			input:    "-Xmx1G",
			expected: []string{"-Xmx1G"},
		},
		{
			name:     "Multiple flags",
			input:    "-Xmx1G -Xms512M -XX:MaxMetaspaceSize=128M",
			expected: []string{"-Xmx1G", "-Xms512M", "-XX:MaxMetaspaceSize=128M"},
		},
		{
			name:     "Flags with quoted values",
			input:    `-Djava.awt.headless=true -Dfile.encoding="UTF-8" -Dspring.profiles.active='production'`,
			expected: []string{"-Djava.awt.headless=true", "-Dfile.encoding=UTF-8", "-Dspring.profiles.active=production"},
		},
		{
			name:     "Flags with spaces in quoted values",
			input:    `-Dapp.name="My Application" -Dlog.file="/var/log/app.log"`,
			expected: []string{"-Dapp.name=My Application", "-Dlog.file=/var/log/app.log"},
		},
		{
			name:     "Multiple spaces between flags",
			input:    "-Xmx1G    -Xms512M     -XX:MaxMetaspaceSize=128M",
			expected: []string{"-Xmx1G", "-Xms512M", "-XX:MaxMetaspaceSize=128M"},
		},
		{
			name:     "Leading and trailing spaces",
			input:    "  -Xmx1G -Xms512M  ",
			expected: []string{"-Xmx1G", "-Xms512M"},
		},
		{
			name:     "Agent with JAR path",
			input:    "-javaagent:/path/to/agent.jar=option1,option2",
			expected: []string{"-javaagent:/path/to/agent.jar=option1,option2"},
		},
		{
			name:     "Complex real-world example",
			input:    `-Xmx2G -Xms1G -XX:MaxMetaspaceSize=256M -Djava.awt.headless=true -Dspring.profiles.active="production" -javaagent:/opt/agents/jmx.jar`,
			expected: []string{"-Xmx2G", "-Xms1G", "-XX:MaxMetaspaceSize=256M", "-Djava.awt.headless=true", "-Dspring.profiles.active=production", "-javaagent:/opt/agents/jmx.jar"},
		},
		{
			name:     "Escaped quotes",
			input:    `-Dvalue="She said \"Hello\""`,
			expected: []string{`-Dvalue=She said "Hello"`},
		},
		{
			name:     "Mixed quote types",
			input:    `-Dvalue1='single quotes' -Dvalue2="double quotes"`,
			expected: []string{"-Dvalue1=single quotes", "-Dvalue2=double quotes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFlags(tt.input)
			if err != nil {
				t.Errorf("ParseFlags() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseFlags() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseFlagsEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Only spaces",
			input:    "   ",
			expected: nil,
		},
		{
			name:     "Unclosed quotes - should still work",
			input:    `-Dvalue="unclosed quote`,
			expected: []string{`-Dvalue=unclosed quote`},
		},
		{
			name:     "Empty quoted string",
			input:    `-Dvalue=""`,
			expected: []string{"-Dvalue="},
		},
		{
			name:     "Just quotes",
			input:    `"" ''`,
			expected: []string{"", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseFlags(tt.input)
			if err != nil {
				t.Errorf("ParseFlags() error = %v", err)
				return
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseFlags() = %v, want %v", result, tt.expected)
			}
		})
	}
}

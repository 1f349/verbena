package utils

import (
	"reflect"
	"testing"
)

func TestParseDnsKeyValue(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "empty",
			input: "",
			want:  map[string]string{},
		},
		{
			name:  "single pair",
			input: "foo=bar",
			want: map[string]string{
				"foo": "bar",
			},
		},
		{
			name:  "multiple pairs",
			input: "foo=bar baz=qux answer=42",
			want: map[string]string{
				"foo":    "bar",
				"baz":    "qux",
				"answer": "42",
			},
		},
		{
			name:  "quoted value",
			input: `city="New York"`,
			want: map[string]string{
				"city": "New York",
			},
		},
		{
			name:  "multiple quoted values",
			input: `city="New York" title="Senior Software Engineer"`,
			want: map[string]string{
				"city":  "New York",
				"title": "Senior Software Engineer",
			},
		},
		{
			name:  "mixed quoted and unquoted",
			input: `name=John city="New York" age=42`,
			want: map[string]string{
				"name": "John",
				"city": "New York",
				"age":  "42",
			},
		},
		{
			name:  "escaped quote",
			input: `msg="He said \"hello\""`,
			want: map[string]string{
				"msg": `He said "hello"`,
			},
		},
		{
			name:  "escaped backslash",
			input: `path="C:\\Windows"`,
			want: map[string]string{
				"path": `C:\Windows`,
			},
		},
		{
			name:  "leading and trailing spaces",
			input: `   foo=bar   baz=qux   `,
			want: map[string]string{
				"foo": "bar",
				"baz": "qux",
			},
		},
		{
			name:    "missing equals",
			input:   "foo",
			wantErr: true,
		},
		{
			name:    "unterminated quote",
			input:   `foo="bar`,
			wantErr: true,
		},
		{
			name:  "empty quoted value",
			input: `foo=""`,
			want: map[string]string{
				"foo": "",
			},
		},
		{
			name:  "empty unquoted value",
			input: `foo=`,
			want: map[string]string{
				"foo": "",
			},
		},
		{
			name:  "equals in value",
			input: `expr=a=b=c`,
			want: map[string]string{
				"expr": "a=b=c",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDnsKeyValue(tt.input)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKV() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseKV() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

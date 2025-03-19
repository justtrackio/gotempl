package main

import (
	"os"
	"testing"
)

func TestIncludeWithYQSelector(t *testing.T) {
	// Create a temporary test file
	content := `
foo:
  bar: test
  baz: value
`
	tmpfile, err := os.CreateTemp("", "test*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		selector string
		want     string
		wantErr  bool
	}{
		{
			name:     "empty selector",
			selector: "",
			want:     content,
			wantErr:  false,
		},
		{
			name:     "valid selector",
			selector: ".foo.bar",
			want:     "test\n",
			wantErr:  false,
		},
		{
			name:     "invalid selector",
			selector: ".[invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := includeWithYQSelector(tt.selector, tmpfile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("includeWithYQSelector() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("includeWithYQSelector() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOapiAbsoluteRefs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple ref replacement",
			input: `"$ref": "file.json#/components/schemas/User"`,
			want:  `"$ref": "#/components/schemas/User"`,
		},
		{
			name:  "ref with path",
			input: `"$ref": "../folder/file.yaml#/components/schemas/User"`,
			want:  `"$ref": "#/components/schemas/User"`,
		},
		{
			name:  "ref without quote",
			input: `- $ref: "../folder/file.yaml#/components/schemas/User"`,
			want:  `- $ref: "#/components/schemas/User"`,
		},
		{
			name:  "no ref",
			input: `"something": "else"`,
			want:  `"something": "else"`,
		},
		{
			name: "multiple refs",
			input: `{"$ref": "file1.json#/components/schemas/User",
					"$ref": "file2.json#/components/schemas/Post",
					"smth": "else"}`,
			want: `{"$ref": "#/components/schemas/User",
					"$ref": "#/components/schemas/Post",
					"smth": "else"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := oapiAbsoluteRefs(tt.input); got != tt.want {
				t.Errorf("oapiAbsoluteRefs() = %v, want %v", got, tt.want)
			}
		})
	}
}

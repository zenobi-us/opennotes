package cmd

import "testing"

func TestIsDSLStyleQuery(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  bool
	}{
		{name: "field query", query: "type:epic", want: true},
		{name: "pipe query", query: "status:todo | limit:5", want: true},
		{name: "plain text", query: "meeting notes", want: false},
		{name: "empty", query: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDSLStyleQuery(tt.query)
			if got != tt.want {
				t.Fatalf("isDSLStyleQuery(%q)=%v want %v", tt.query, got, tt.want)
			}
		})
	}
}

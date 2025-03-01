package markdown

import "testing"

func TestConvertLinks(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:   "test links 1",
			source: "https://github.com/oktalz/present/releases",
			//revive:disable:line-length-limit
			want:    `<a href="https://github.com/oktalz/present/releases">https://github.com/oktalz/present/releases</a>`,
			wantErr: false,
		},
		{
			name:    "test links 2",
			source:  "[releases](https://github.com/oktalz/present/releases)",
			want:    `<a href="https://github.com/oktalz/present/releases">releases</a>`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

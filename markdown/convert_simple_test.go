package markdown

import "testing"

func TestConvertSimple(t *testing.T) { //nolint:funlen
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "test",
			source:  "Hello world",
			want:    "Hello world",
			wantErr: false,
		},
		{
			name:    "test highlight",
			source:  "Hello `highlight` world",
			want:    "Hello <code>highlight</code> world",
			wantErr: false,
		},
		{
			name:    "test bold",
			source:  "Hello **Bold** world",
			want:    "Hello <strong>Bold</strong> world",
			wantErr: false,
		},
		{
			name:    "test italics",
			source:  "Hello *Italics* world",
			want:    "Hello <em>Italics</em> world",
			wantErr: false,
		},
		{
			name:    "test ~strikethrough~",
			source:  "Hello ~strikethrough~ world",
			want:    "Hello <del>strikethrough</del> world",
			wantErr: false,
		},
		{
			name:    "test ~strikethrough~",
			source:  "Hello ~strikethrough~ world",
			want:    "Hello <del>strikethrough</del> world",
			wantErr: false,
		},
		{
			name:    "test emoji",
			source:  "Hello :warning: :construction: world",
			want:    "Hello &#x26a0;&#xfe0f; &#x1f6a7; world",
			wantErr: false,
		},
		{
			name:    "test emoji 3",
			source:  "Hello ⚠️ 🚧 world",
			want:    "Hello ⚠️ 🚧 world",
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

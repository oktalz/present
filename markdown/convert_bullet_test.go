package markdown

import "testing"

func TestConvertBullet(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    string
		wantErr bool
	}{
		{
			name:    "test bullet 1",
			source:  "- one\n- two\n- three",
			want:    "<ul>\n<li>one</li>\n<li>two</li>\n<li>three</li>\n</ul>",
			wantErr: false,
		},
		{
			name:    "test bullet 1",
			source:  "- one\n  - two\n- three",
			want:    "<ul>\n<li>one\n<ul>\n<li>two</li>\n</ul>\n</li>\n<li>three</li>\n</ul>",
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

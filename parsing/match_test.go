package parsing

import "testing"

func TestFindData(t *testing.T) {
	type args struct {
		fileContent string
		pattern     Pattern
	}
	tests := []struct {
		name  string
		args  args
		want  int
		want1 int
		want2 string
	}{
		{
			name:  "FindData",
			args:  args{fileContent: ".bx{.bx{content}}", pattern: Pattern{Start: ".bx{", End: "}"}},
			want:  4,
			want1: 16,
			want2: ".bx{content}",
		},
		{
			name:  "FindData2",
			args:  args{fileContent: "{content}", pattern: Pattern{Start: "{", End: "}"}},
			want:  1,
			want1: 8,
			want2: "content",
		},
		{
			name: "FindData3",
			args: args{
				fileContent: ".{I have something.{s} and another.{s} .{content}}",
				pattern:     Pattern{Start: ".{", End: "}"},
			},
			want:  2,
			want1: 49,
			want2: "I have something.{s} and another.{s} .{content}",
		},
		{
			name: "FindData4",
			args: args{
				fileContent: ".{I have something.{s.{d}} and another.{s} .{content}}",
				pattern:     Pattern{Start: ".{", End: "}"},
			},
			want:  2,
			want1: 53,
			want2: "I have something.{s.{d}} and another.{s} .{content}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := FindData(tt.args.fileContent, tt.args.pattern)
			if got != tt.want {
				t.Errorf("FindData() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("FindData() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("FindData() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

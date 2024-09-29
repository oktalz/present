package markdown

import "testing"

func Test_processReplace(t *testing.T) {
	type args struct {
		fileContent string
		startStr    string
		endStr      string
		process     func(data string) string
	}
	tests := []struct {
		name string
		args args
		want string
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processReplace(tt.args.fileContent, tt.args.startStr, tt.args.endStr, tt.args.process); got != tt.want {
				t.Errorf("processReplace() = %v, want %v", got, tt.want)
			}
		})
	}
}

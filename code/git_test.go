package code

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_execPipeline(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "exec")
	defer os.RemoveAll(tmpDir)

	// Generate by VSCode plugin
	type args struct {
		dir      string
		commands [][]string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		// Testcases
		{
			name: "echo test",
			args: args{
				dir: tmpDir,
				commands: [][]string{
					[]string{"echo", "foobar"},
					[]string{"grep", "foo"},
				},
			},
			want:    []byte("foobar\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := execPipeline(tt.args.dir, tt.args.commands...)
			if (err != nil) != tt.wantErr {
				t.Errorf("execPipeline() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("execPipeline() = %s, want %s", got, tt.want)
			}
		})
	}
}

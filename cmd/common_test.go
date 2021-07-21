package cmd

import (
	"strings"
	"testing"
)

func TestGoMain(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"config", args{strings.Split("./proot-go --go-verbose ls", " ")}},
		{"help", args{strings.Split("./proot-go --help --go-verbose=0 ls", " ")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			GoMain(tt.args.args)
		})
	}
}

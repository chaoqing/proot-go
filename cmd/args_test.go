package cmd

import (
	"strings"
	"testing"
)

func Test_prootFlagSet(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			"main",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prootFlagSet(); got != nil {
				got.PrintDefaults()
			} else {
				t.Error("prootFlagSet() ==nil")
			}
		})
	}
}

func isSameStringArray(a, b []string) bool{
	if len(a)!=len(b){
		return false
	}

	for i, arg := range a{
		if arg != b[i]{
			return false
		}
	}

	return true
}

func Test_splitArgs(t *testing.T) {
	type args struct {
		binName string
		osArgs []string
	}
	tests := []struct {
		name        string
		args        args
		wantGoArgs  []string
		wantCArgs   []string
		wantCmdArgs []string
	}{
		{
			"proot no args",
			args{"proot", []string{}},
			[]string{}, []string{}, []string{},
		},

		{
			"proot go help",
			args{"proot", strings.Split("--go-help", " ")},
			[]string{"--go-help"}, []string{}, []string{},
		},

		{
			"proot help",
			args{"proot", strings.Split("--help", " ")},
			[]string{}, []string{"--help"}, []string{},
		},

		{
			"proot normal",
			args{"proot-go", strings.Split("make", " ")},
			[]string{}, []string{}, []string{"make"},
		},

		{
			"proot normal with args",
			args{"proot-go", strings.Split("-b /:/ -0 make --help", " ")},
			[]string{}, []string{"-b", "/:/", "-0"}, []string{"make", "--help"},
		},

		{
			"proot mixed args",
			args{"proot-go", strings.Split("proot-go -b /:/ --go-verbose -0 make --help --go-help", " ")},
			[]string{"--go-verbose"}, []string{"-b", "/:/", "-0"}, []string{"make", "--help", "--go-help"},
		},

		{
			"busybox like",
			args{"make", strings.Split("--go-help", " ")},
			[]string{}, []string{}, []string{"make", "--go-help"},
		},
	}



	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotGoArgs, gotCArgs, gotCmdArgs := SplitArgs(tt.args.binName, tt.args.osArgs)
			if !isSameStringArray(gotGoArgs, tt.wantGoArgs) {
				t.Errorf("splitArgs() gotGoArgs = %v, want %v", gotGoArgs, tt.wantGoArgs)
			}
			if !isSameStringArray(gotCArgs, tt.wantCArgs) {
				t.Errorf("splitArgs() gotCArgs = %v, want %v", gotCArgs, tt.wantCArgs)
			}
			if !isSameStringArray(gotCmdArgs, tt.wantCmdArgs) {
				t.Errorf("splitArgs() gotCmdArgs = %v, want %v", gotCmdArgs, tt.wantCmdArgs)
			}
		})
	}
}

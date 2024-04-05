package fsutils

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	type args struct {
		path string
	}
	wd, _ := os.Getwd()
	var tests = []struct {
		name string
		args args
		want bool
	}{
		{name: "Existing VFile", args: struct{ path string }{path: wd + "/testdata/test.json"}, want: true},
		{name: "Non Existing VFile", args: struct{ path string }{path: wd + "/testdata/test-nonexisting.json"}, want: false},
		{name: "Dir As VFile", args: struct{ path string }{path: wd + "/testdata"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.path); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	type args struct {
		path string
	}
	wd, _ := os.Getwd()
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Existing Dir", args: struct{ path string }{path: wd + "/testdata"}, want: true},
		{name: "Non Existing Dir", args: struct{ path string }{path: wd + "/test"}, want: false},
		{name: "VFile AS DIR", args: struct{ path string }{path: wd + "/testdata/test.json"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DirExists(tt.args.path); got != tt.want {
				t.Errorf("DirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	type args struct {
		path string
	}
	wd, _ := os.Getwd()
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Existing Dir", args: struct{ path string }{path: wd + "/testdata"}, want: true},
		{name: "Non Existing Dir", args: struct{ path string }{path: wd + "/test"}, want: false},
		{name: "Existing VFile", args: struct{ path string }{path: wd + "/testdata/test.json"}, want: true},
		{name: "Non Existing VFile", args: struct{ path string }{path: wd + "/testdata/unknown"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathExists(tt.args.path); got != tt.want {
				t.Errorf("DirExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLookupContentType(t *testing.T) {
	type args struct {
		path string
	}
	wd, _ := os.Getwd()

	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "JSON VFile", args: struct{ path string }{path: wd + "/testdata/test.json"}, want: "application/json"},
		{name: "YAML VFile", args: struct{ path string }{path: wd + "/testdata/test.yaml"}, want: "text/yaml"},
		{name: "Dat VFile", args: struct{ path string }{path: wd + "/testdata/test.dat"}, want: "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LookupContentType(tt.args.path); got != tt.want {
				t.Errorf("LookupContentType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDetectContentType(t *testing.T) {
	type args struct {
		path string
	}
	wd, _ := os.Getwd()

	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "TXT", args: struct{ path string }{path: wd + "/testdata/test.dat"}, want: "text/plain; charset=utf-8", wantErr: false},
		{name: "INVALID", args: struct{ path string }{path: wd + "/testdata/test.abc"}, want: "", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DetectContentType(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("DetectContentType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DetectContentType() got = %v, want %v", got, tt.want)
			}
		})
	}
}

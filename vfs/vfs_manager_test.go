package vfs

import (
	"reflect"
	"testing"
)

func TestGetManager(t *testing.T) {
	tests := []struct {
		name string
		want VFileSystem
	}{
		{
			name: "localFs",
			want: &fileSystems{
				fileSystems: map[string]VFileSystem{"file": &OsFs{BaseVFS: &BaseVFS{VFileSystem: &OsFs{}}}, "": &OsFs{BaseVFS: &BaseVFS{VFileSystem: &OsFs{}}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFileSystems_IsSupported(t *testing.T) {
	var scheme string
	var output bool

	testManager := GetManager()

	scheme = "file"
	output = testManager.IsSupported(scheme)
	if output == false {
		t.Error()
	}
	scheme = "test"
	output = testManager.IsSupported(scheme)
	if output == true {
		t.Error()
	}
}

func Test_InvalidFS(t *testing.T) {
	testManager := GetManager()
	u := GetRawPath("dummy:///raw-abc.txt")
	_, err := testManager.CreateRaw(u)
	if err == nil {
		t.Errorf("CreateRaw() error = %v", err)
	}
	if err.Error() != "Unsupported scheme dummy for in the url "+u {
		t.Errorf("Test_InvalidFS() error = %v", err)
	}
}

func TestFileSystems_Schemes(t *testing.T) {
	testManager := GetManager()

	output := testManager.Schemes()
	if output[0] != "file" {
		t.Errorf("Schemes() default scheme added is file")
	}
}

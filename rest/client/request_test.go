package client

import (
	"net/http"
	"testing"
)

func TestRequest_Method(t *testing.T) {
	tests := []struct {
		name   string
		url    string
		method string
		want   string
	}{{
		name:   "Test1",
		url:    "http://localhost:8080",
		method: http.MethodGet,
		want:   http.MethodGet,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := client.NewRequest(tt.url, tt.method)
			got := req.Method()
			if got != tt.want {
				t.Errorf("Error in validation :: got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRequest_Options(t *testing.T) {
	req := client.NewRequest("http://localhost:8080", http.MethodGet)

	reqQueryParam := req.AddQueryParam("k", "v")
	if !reqQueryParam.queryParam.Has("k") {
		t.Errorf("Error in adding query-params")
	}

	reqAddHeader := req.AddHeader("header", "testing")
	if reqAddHeader.header.Get("header") != "testing" {
		t.Errorf("Error in adding headers")
	}

	body := struct {
		key string
	}{key: "hello-world"}
	reqAddBody := req.SetBody(body)
	if reqAddBody.body == nil {
		t.Errorf("Error in adding body")
	}

	reqAddContentType := req.SetContentType("application/json")
	if reqAddContentType.contentType != "application/json" {
		t.Errorf("Error in adding content type")
	}

	file1 := &MultipartFile{
		ParamName: "file1",
		FilePath:  "./testdata/test.json",
	}
	file2 := &MultipartFile{
		ParamName: "file2",
		FilePath:  "./testdata/test2.json",
	}
	reqAddMultipartFiles := req.SetMultipartFiles(file1, file2)
	if len(reqAddMultipartFiles.multiPartFiles) == 0 {
		t.Errorf("Error in adding multipart files")
	}
}

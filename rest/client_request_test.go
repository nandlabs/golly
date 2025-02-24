package rest

import (
	"net/http"
	"testing"

	"oss.nandlabs.io/golly/testing/assert"
)

func TestRequest_AddFormData(t *testing.T) {
	req := &Request{}
	req.AddFormData("key1", "value1", "value2")

	if req.formData.Get("key1") != "value1" {
		t.Errorf("Expected key1 to be value1, got %s", req.formData.Get("key1"))
	}

	values := req.formData["key1"]
	if len(values) != 2 || values[1] != "value2" {
		t.Errorf("Expected key1 to have two values, got %v", values)
	}
}

func TestRequest_AddQueryParam(t *testing.T) {
	req := &Request{}
	req.AddQueryParam("key1", "value1", "value2")

	if req.queryParam.Get("key1") != "value1" {
		t.Errorf("Expected key1 to be value1, got %s", req.queryParam.Get("key1"))
	}

	values := req.queryParam["key1"]
	if len(values) != 2 || values[1] != "value2" {
		t.Errorf("Expected key1 to have two values, got %v", values)
	}
}

func TestRequest_AddPathParam(t *testing.T) {
	req := &Request{}
	req.AddPathParam("key1", "value1")

	if req.pathParams["key1"] != "value1" {
		t.Errorf("Expected key1 to be value1, got %s", req.pathParams["key1"])
	}
}

func TestRequest_AddHeader(t *testing.T) {
	req := &Request{header: http.Header{}}
	req.AddHeader("Content-Type", "application/json")

	if req.header.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type to be application/json, got %s", req.header.Get("Content-Type"))
	}
}

func TestRequest_SetBody(t *testing.T) {
	req := &Request{}
	body := map[string]string{"key": "value"}
	req.SetBody(body)

	assert.Equal(t, body, req.body)
}

func TestRequest_SetContentType(t *testing.T) {
	req := &Request{}
	req.SetContentType("application/json")

	if req.contentType != "application/json" {
		t.Errorf("Expected contentType to be application/json, got %s", req.contentType)
	}
}

func TestRequest_SetMultipartFiles(t *testing.T) {
	req := &Request{}
	files := []*MultipartFile{
		{ParamName: "file1", FilePath: "path/to/file1"},
		{ParamName: "file2", FilePath: "path/to/file2"},
	}
	req.SetMultipartFiles(files...)

	if len(req.multiPartFiles) != 2 {
		t.Errorf("Expected 2 multipart files, got %d", len(req.multiPartFiles))
	}
	if req.multiPartFiles[0].ParamName != "file1" || req.multiPartFiles[1].ParamName != "file2" {
		t.Errorf("Multipart files not set correctly")
	}
}

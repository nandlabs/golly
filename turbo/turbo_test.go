package turbo

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sync"
	"testing"
)

func TestNewRouter(t *testing.T) {
	tests := []struct {
		name string
		want *Router
	}{
		{
			name: "InitTest",
			want: NewRouter(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRouter(); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("NewRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_findRoute(t *testing.T) {
	var tlr = make(map[string]*Route)
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		req *http.Request
	}
	route := &Route{
		path:         "abc",
		isPathVar:    false,
		childVarName: "",
		handlers:     make(map[string]http.Handler),
		subRoutes:    make(map[string]*Route),
		queryParams:  nil,
	}
	tlr["abc"] = route
	testUrl, _ := url.Parse("/abc")
	a := args{req: &http.Request{
		URL: testUrl,
	}}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
		want1  []Param
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           tlr,
			},
			args:  a,
			want:  route,
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			got, gotMap := router.findRoute(tt.args.req)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findRoute() got = %v, want %v", got, tt.want)
			}

			if reflect.TypeOf(gotMap) != reflect.TypeOf(tt.want1) {
				t.Errorf("findRoute() got = %v, want %v", gotMap, tt.want1)
			}
		})
	}
}

func TestRouter_GetPathParams(t *testing.T) {
	req := &http.Request{}
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id  string
		val string
		r   *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key",
				val: "value",
				r:   req,
			},
			want: "value",
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key2",
				val: "73",
				r:   req,
			},
			want: "73",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			var params []Param = nil
			params = []Param{}
			params = append(params,
				Param{
					key:   tt.args.id,
					value: tt.args.val,
				})
			got, _ := router.GetPathParams(tt.args.id, tt.args.r.WithContext(context.WithValue(tt.args.r.Context(), "params", params)))
			logger.Info(tt.args.r.Context().Value("params"))
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetPathParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetPathParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_GetPathParamsFail(t *testing.T) {
	req := &http.Request{}
	router := &Router{
		unManagedRouteHandler:    nil,
		unsupportedMethodHandler: nil,
		topLevelRoutes:           nil,
	}
	got, err := router.GetPathParams("foo", req)
	if err != nil {
		if got != "err" {
			t.Errorf("GetPathParams() = %v, want %v", got, "err")
		}
	}
}

func TestRouter_GetIntPathParams(t *testing.T) {
	req := &http.Request{}
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id  string
		val string
		r   *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key",
				val: "2134",
				r:   req,
			},
			want: 2134,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key2",
				val: "7337",
				r:   req,
			},
			want: 7337,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			var params []Param = nil
			params = []Param{}
			params = append(params,
				Param{
					key:   tt.args.id,
					value: tt.args.val,
				})
			got, _ := router.GetIntPathParams(tt.args.id, tt.args.r.WithContext(context.WithValue(tt.args.r.Context(), "params", params)))

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetIntPathParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetIntPathParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_GetIntPathParamsFail(t *testing.T) {
	req := &http.Request{}
	router := &Router{
		unManagedRouteHandler:    nil,
		unsupportedMethodHandler: nil,
		topLevelRoutes:           nil,
	}
	got, err := router.GetIntPathParams("foo", req)
	if err != nil {
		if got != -1 {
			t.Errorf("GetIntPathParams() = %v, want %v", got, "err")
		}
	}
}

func TestRouter_GetFloatPathParams(t *testing.T) {
	req := &http.Request{}
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id  string
		val string
		r   *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key",
				val: "73.37",
				r:   req,
			},
			want: 73.37,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key2",
				val: "73.33333337",
				r:   req,
			},
			want: 73.33333337,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			var params []Param = nil
			params = []Param{}
			params = append(params,
				Param{
					key:   tt.args.id,
					value: tt.args.val,
				})
			got, _ := router.GetFloatPathParams(tt.args.id, tt.args.r.WithContext(context.WithValue(tt.args.r.Context(), "params", params)))

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetFloatPathParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetFloatPathParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_GetFloatPathParamsFail(t *testing.T) {
	req := &http.Request{}
	router := &Router{
		unManagedRouteHandler:    nil,
		unsupportedMethodHandler: nil,
		topLevelRoutes:           nil,
	}
	got, err := router.GetFloatPathParams("foo", req)
	if err != nil {
		if got != -1 {
			t.Errorf("GetFloatPathParams() = %v, want %v", got, "err")
		}
	}
}

func TestRouter_GetBoolPathParams(t *testing.T) {
	req := &http.Request{}
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id  string
		val string
		r   *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key",
				val: "true",
				r:   req,
			},
			want: true,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id:  "key2",
				val: "false",
				r:   req,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			var params []Param = nil
			params = []Param{}
			params = append(params,
				Param{
					key:   tt.args.id,
					value: tt.args.val,
				})
			got, _ := router.GetBoolPathParams(tt.args.id, tt.args.r.WithContext(context.WithValue(tt.args.r.Context(), "params", params)))

			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetBoolPathParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetBoolPathParams() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_GetBoolPathParamsFail(t *testing.T) {
	req := &http.Request{}
	router := &Router{
		unManagedRouteHandler:    nil,
		unsupportedMethodHandler: nil,
		topLevelRoutes:           nil,
	}
	got, err := router.GetBoolPathParams("foo", req)
	if err != nil {
		if got != false {
			t.Errorf("GetBoolPathParams() = %v, want %v", got, "err")
		}
	}
}

var strUrl, _ = url.Parse("https://foo.com?test1=value1&test2=value2&test3=")

func TestRouter_GetQueryParams(t *testing.T) {
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id string
		r  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: strUrl},
			},
			want: "value1",
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test2",
				r:  &http.Request{URL: strUrl},
			},
			want: "value2",
		},
		{
			name: "Test3",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test3",
				r:  &http.Request{URL: strUrl},
			},
			want: "err",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			got, _ := router.GetQueryParams(tt.args.id, tt.args.r)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetQueryParams() Type Got = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetQueryParams() Value Got = %v, want %v", got, tt.want)
			}
		})
	}
}

var intUrl, _ = url.Parse("https://foo.com?test1=73&test2=7337")
var intUrlFail, _ = url.Parse("https://foo.com?test1=foo")

func TestRouter_GetIntQueryParams(t *testing.T) {
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id string
		r  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: intUrl},
			},
			want: 73,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test2",
				r:  &http.Request{URL: intUrl},
			},
			want: 7337,
		},
		{ // Failure Test Case
			name: "Test3",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: intUrlFail},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			got, _ := router.GetIntQueryParams(tt.args.id, tt.args.r)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetIntQueryParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetIntQueryParams() Value Got = %v, want %v", got, tt.want)
			}
		})
	}
}

var floatUrl, _ = url.Parse("https://foo.com?test1=7.3&test2=73.37")
var floatUrlFail, _ = url.Parse("https://foo.com?test1=hello")

func TestRouter_GetFloatQueryParams(t *testing.T) {
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id string
		r  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: floatUrl},
			},
			want: 7.3,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test2",
				r:  &http.Request{URL: floatUrl},
			},
			want: 73.37,
		},
		{ // Failure Test Case
			name: "Test3",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: floatUrlFail},
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			got, _ := router.GetFloatQueryParams(tt.args.id, tt.args.r)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetFloatQueryParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetFloatQueryParams() Value Got = %v, want %v", got, tt.want)
			}
		})
	}
}

var boolUrl, _ = url.Parse("https://foo.com?test1=true&test2=false")
var boolUrlFail, _ = url.Parse("https://foo.com?test1=fail")

func TestRouter_GetBoolQueryParams(t *testing.T) {
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		id string
		r  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: boolUrl},
			},
			want: true,
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test2",
				r:  &http.Request{URL: boolUrl},
			},
			want: false,
		},
		{ // Failure Test Case
			name: "Test3",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           nil,
			},
			args: args{
				id: "test1",
				r:  &http.Request{URL: boolUrlFail},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			got, _ := router.GetBoolQueryParams(tt.args.id, tt.args.r)
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("GetBoolQueryParams() = %v, want %v", got, tt.want)
			}
			if got != tt.want {
				t.Errorf("GetBoolQueryParams() Value Got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_Get(t *testing.T) {
	type fields struct {
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		path string
		f    func(w http.ResponseWriter, r *http.Request)
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/health",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("hello from turbo"))
				},
			},
			want: &Route{},
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/health/:id",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("hello from turbo"))
				},
			},
			want: &Route{},
		},
		{
			name: "Test3",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("hello from turbo"))
				},
			},
			want: &Route{},
		},
		{
			name: "Test4",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/getCustomer/:id/getData",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("hello from turbo"))
				},
			},
			want: &Route{},
		}, {
			name: "Test5",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/health/{id}",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("hello from turbo"))
				},
			},
			want: &Route{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			if got := router.Get(tt.args.path, tt.args.f); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Apply() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_Add(t *testing.T) {
	type fields struct {
		lock                     sync.RWMutex
		unManagedRouteHandler    http.Handler
		unsupportedMethodHandler http.Handler
		topLevelRoutes           map[string]*Route
	}
	type args struct {
		path    string
		f       func(w http.ResponseWriter, r *http.Request)
		methods []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Route
	}{
		{
			name: "Test1",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/foo",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("fonzi says hello"))
				},
				methods: []string{"PUT", "POST"},
			},
		},
		{
			name: "Test2",
			fields: fields{
				unManagedRouteHandler:    nil,
				unsupportedMethodHandler: nil,
				topLevelRoutes:           make(map[string]*Route),
			},
			args: args{
				path: "/api/v1/fonzi",
				f: func(w http.ResponseWriter, r *http.Request) {
					_ = json.NewEncoder(w).Encode([]byte("don't delete fonzi"))
				},
				methods: []string{"DELETE"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := &Router{
				unManagedRouteHandler:    tt.fields.unManagedRouteHandler,
				unsupportedMethodHandler: tt.fields.unsupportedMethodHandler,
				topLevelRoutes:           tt.fields.topLevelRoutes,
			}
			if got := router.Add(tt.args.path, tt.args.f, tt.args.methods...); reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("Test Passes"))
}

func dummyFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("dummy Filter Added")
		next.ServeHTTP(w, r)
	})
}

func TestRouter_ServeHTTP(t *testing.T) {
	var router = NewRouter()
	router.Get("/api/fooTest", dummyHandler)
	router.Delete("/api/deleteFoo", dummyHandler)
	router.Put("/api/putFoo/:id", dummyHandler)
	router.Post("/api/putBar/:id", dummyHandler)
	router.Get("/api/foo", dummyHandler).AddFilter(dummyFilter)

	type args struct {
		path   string
		f      func(w http.ResponseWriter, r *http.Request)
		method string
	}

	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "Test1",
			args: args{
				path:   "/api/fooTest",
				f:      dummyHandler,
				method: GET,
			},
			want: http.StatusOK,
		},
		{
			name: "Test2",
			args: args{
				path:   "/api/fooTest",
				f:      dummyHandler,
				method: PUT,
			},
			want: http.StatusMethodNotAllowed,
		},
		{
			name: "Test3",
			args: args{
				path:   "/api/fooTest/bar",
				f:      dummyHandler,
				method: PUT,
			},
			want: http.StatusNotFound,
		},
		{
			name: "Test4",
			args: args{
				path:   "///api///fooTest//",
				f:      dummyHandler,
				method: GET,
			},
			want: http.StatusMovedPermanently,
		},
		{
			name: "Test5",
			args: args{
				path:   "api///fooTest//",
				f:      dummyHandler,
				method: GET,
			},
			want: http.StatusMovedPermanently,
		},
		{
			name: "Test6",
			args: args{
				path:   "",
				f:      dummyHandler,
				method: GET,
			},
			want: http.StatusMovedPermanently,
		},
		{
			name: "Test7",
			args: args{
				path:   "/api/foo",
				f:      dummyHandler,
				method: GET,
			},
			want: http.StatusOK,
		},
		{
			name: "Test8",
			args: args{
				path:   "/api/putFoo/",
				f:      dummyHandler,
				method: PUT,
			},
			want: http.StatusNotFound,
		},
		{
			name: "Test9",
			args: args{
				path:   "/api/putBar/123",
				f:      dummyHandler,
				method: POST,
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tt.args.method, tt.args.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			router.ServeHTTP(w, r)
			if w.Result().StatusCode != tt.want {
				t.Errorf("ServeHttp() got = %v, want = %v", w.Result().StatusCode, tt.want)
			}
		})
	}

}

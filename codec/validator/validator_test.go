package validator_test

import (
	"go.nandlabs.io/golly/codec/validator"
	"testing"
)

var sv = validator.NewStructValidator()

func TestSkipValidation(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name: "Test-pass-1",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:"min=10"`
				Mobile int    `json:"mobile" constraints:""`
			}{Name: "Testings", Age: 20, Mobile: 123456789},
		},
		{
			name: "Test-pass-2",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:""`
				Mobile int    `json:"mobile"`
			}{Name: "Testings", Age: 20, Mobile: 123456789},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if err != nil {
				t.Errorf("Error in validation: %s", err)
			}
		})
	}
}

func TestSuccessValidation(t *testing.T) {
	type ReqMsg struct {
		Name string `json:"name" constraints:"notnull=true"`
		Age  int    `json:"age" constraints:"min=0"`
	}

	msg := ReqMsg{
		Name: "Test",
		Age:  11,
	}
	if err := sv.Validate(msg); err != nil {
		t.Errorf("Error in validation: %s", err)
	}
}

/**
min, max validations test
*/

func TestNumericValidations(t *testing.T) {

	testsPass := []struct {
		Name  string
		input interface{}
	}{
		{
			Name: "Test-pass-1",
			input: struct {
				MinC1 int `json:"minC1" constraints:"min=10"`
				MaxC1 int `json:"maxC1" constraints:"max=49"`
			}{MinC1: 12, MaxC1: 45},
		},
		/**
		exclusive min/max validation test
		*/
		{
			Name: "Test-pass-2",
			input: struct {
				MinC4 int `json:"minC4" constraints:"exclusiveMin=10"`
				MaxC4 int `json:"maxC4" constraints:"exclusiveMax=50"`
			}{MinC4: 10, MaxC4: 50},
		},
		{
			Name: "Test-pass-3",
			input: struct {
				Num7 int `json:"num7" constraints:"multipleOf=5"`
			}{Num7: 10},
		},
	}

	for _, tt := range testsPass {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if err != nil {
				t.Errorf("Error in validation: %s", err)
			}
		})
	}

	testsError := []struct {
		Name  string
		input interface{}
		want  string
	}{
		{
			Name: "Test-fail-1",
			input: struct {
				MinC2 int `json:"minC2" constraints:"min=10"`
				MaxC2 int `json:"maxC2" constraints:"max=49"`
			}{MinC2: 7, MaxC2: 45},
			want: "min value validation failed for field MinC2",
		},
		{
			Name: "Test-fail-2",
			input: struct {
				MinC3 int `json:"minC3" constraints:"min=10"`
				MaxC3 int `json:"maxC3" constraints:"max=49"`
			}{MinC3: 12, MaxC3: 55},
			want: "max value validation failed for field MaxC3",
		},
		/**
		exclusive min/max validation test
		*/
		{
			Name: "Test-fail-3",
			input: struct {
				MinC5 int `json:"minC5" constraints:"exclusiveMin=10"`
				MaxC5 int `json:"maxC5" constraints:"exclusiveMax=50"`
			}{MinC5: 9, MaxC5: 50},
			want: "exclusive min validation failed for field MinC5",
		},
		{
			Name: "Test-fail-4",
			input: struct {
				MinC6 int `json:"minC6" constraints:"exclusiveMin=10"`
				MaxC6 int `json:"maxC6" constraints:"exclusiveMax=50"`
			}{MinC6: 10, MaxC6: 51},
			want: "exclusive max validation failed for field MaxC6",
		},
		{
			Name: "Test-fail-5",
			input: struct {
				Num8 int `json:"num8" constraints:"multipleOf=5"`
			}{Num8: 11},
			want: "multipleOf validation failed for field Num8",
		},
		{
			Name: "Test-fail-6",
			input: struct {
				Name string `json:"name" constraints:"min=5"`
			}{Name: "hello_world"},
			want: "invalid validation applied to the field Name",
		},
		{
			Name: "Test-fail-7",
			input: struct {
				Name string `json:"name" constraints:"max=5"`
			}{Name: "hello_world"},
			want: "invalid validation applied to the field Name",
		},
		{
			Name: "Test-fail-8",
			input: struct {
				Name string `json:"name" constraints:"exclusiveMin=5"`
			}{Name: "hello_world"},
			want: "invalid validation applied to the field Name",
		},
		{
			Name: "Test-fail-9",
			input: struct {
				Name string `json:"name" constraints:"exclusiveMax=5"`
			}{Name: "hello_world"},
			want: "invalid validation applied to the field Name",
		},
		{
			Name: "Test-fail-10",
			input: struct {
				Name string `json:"name" constraints:"multipleOf=5"`
			}{Name: "hello_world"},
			want: "invalid validation applied to the field Name",
		},
	}

	for _, tt := range testsError {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if tt.want != err.Error() {
				t.Errorf("Got: %s, want: %s", err, tt.want)
			}
		})
	}
}

func TestStringValidation(t *testing.T) {
	testsPass := []struct {
		Name  string
		input interface{}
	}{
		{
			Name: "Test-pass-1",
			input: struct {
				Name string `json:"name" constraints:"notnull=true"`
			}{Name: "testing"},
		},
		{
			Name: "Test-pass-2",
			input: struct {
				Str1T1 string `json:"str1T1" constraints:"min-length=10"`
				Str2T1 string `json:"str2T1" constraints:"max-length=15"`
			}{Str1T1: "hello_world", Str2T1: "hello_world_go"},
		},
		/**
		pattern validations
		*/
		{
			Name: "Test-pass-3",
			input: struct {
				Str4 string `json:"str4" constraints:"pattern=^[tes]{4}.*"`
			}{Str4: "test1234"},
		},
		{
			Name: "Test-pass-4",
			input: struct {
				Str4 string `json:"str4" constraints:"pattern=gray|grey"`
			}{Str4: "grey"},
		},
	}

	for _, tt := range testsPass {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if err != nil {
				t.Errorf("Error in validation: %s", err)
			}
		})
	}

	testsFail := []struct {
		Name  string
		input interface{}
		want  string
	}{
		{
			Name: "Test-fail-1",
			input: struct {
				Str1T2 string `json:"str1T2" constraints:"min-length=10"`
				Str2T2 string `json:"str2T2" constraints:"max-length=15"`
			}{Str1T2: "hell_worl", Str2T2: "hello_world_go"},
			want: "min-length validation failed for field Str1T2",
		},
		{
			Name: "Test-fail-2",
			input: struct {
				Str1T3 string `json:"str1T3" constraints:"min-length=10"`
				Str2T3 string `json:"str2T3" constraints:"max-length=15"`
			}{Str1T3: "hello_world", Str2T3: "hello_world_from_go"},
			want: "max-length validation failed for field Str2T3",
		},
		/**
		pattern validations
		*/
		{
			Name: "Test-fail-3",
			input: struct {
				Str5 string `json:"str5" constraints:"pattern=^[tes]{4}.*"`
			}{Str5: "abcd1234"},
			want: "pattern validation failed for field Str5",
		},
		{
			Name: "Test-fail-4",
			input: struct {
				Str6 string `json:"str6" constraints:"pattern=["`
			}{Str6: "tsst1234"},
			want: "invalid constraint pattern with value '[' for field Str6",
		},
		{
			Name: "Test-fail-5",
			input: struct {
				Str string `json:"str" constraints:"pattern=gray|grey"`
			}{Str: "gry"},
			want: "pattern validation failed for field Str",
		},
		{
			Name: "Test-fail-6",
			input: struct {
				Name string `json:"name" constraints:"notnull=true"`
			}{Name: ""},
			want: "notnull validation failed for field Name",
		},
		{
			Name: "Test-fail-7",
			input: struct {
				Name string `json:"name" constraints:"notnull=dummy"`
			}{Name: ""},
			want: "invalid constraint notnull with value 'dummy' for field Name",
		},
		{
			Name: "Test-fail-8",
			input: struct {
				Age int `json:"name" constraints:"notnull=dummy"`
			}{Age: 22},
			want: "invalid validation applied to the field Age",
		},
		{
			Name: "Test-fail-9",
			input: struct {
				Age int `json:"name" constraints:"min-length=5"`
			}{Age: 22},
			want: "invalid validation applied to the field Age",
		},
		{
			Name: "Test-fail-10",
			input: struct {
				Age int `json:"name" constraints:"max-length=14"`
			}{Age: 22},
			want: "invalid validation applied to the field Age",
		},
		{
			Name: "Test-fail-11",
			input: struct {
				Age int `json:"name" constraints:"pattern=["`
			}{Age: 22},
			want: "invalid validation applied to the field Age",
		},
	}

	for _, tt := range testsFail {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if tt.want != err.Error() {
				t.Errorf("Got: %s, want: %s", err, tt.want)
			}
		})
	}
}

/**
Nested Structure Testing
**Not working as per latest algo**
*/

type Example struct {
	Reference
	Summary     string      `json:"summary,omitempty" constraints:""`
	Description string      `json:"description,omitempty" constraints:""`
	Value       interface{} `json:"example,omitempty" constraints:""`
}

type Reference struct {
	Ref            string `json:"ref" constraints:""`
	RefDescription string `json:"ref-description" constraints:""`
	RefSummary     string `json:"ref-summary" constraints:""`
}

func TestNested(t *testing.T) {
	msg := Example{
		Reference: Reference{
			Ref:            "reference",
			RefSummary:     "ref summary",
			RefDescription: "ref description",
		},
		Summary:     "summary",
		Description: "description",
		Value:       nil,
	}

	if err := sv.Validate(msg); err != nil {
		t.Errorf("Error in validation: %s", err)
	}
}

type ExampleFail struct {
	ReferenceFail
	Summary     string      `json:"summary,omitempty" constraints:""`
	Description string      `json:"description,omitempty" constraints:""`
	Value       interface{} `json:"example,omitempty" constraints:""`
}

type ReferenceFail struct {
	Ref            string `json:"ref" constraints:"min-length=10"`
	RefDescription string `json:"ref-description" constraints:""`
	RefSummary     string `json:"ref-summary" constraints:""`
}

func TestNestedFail(t *testing.T) {
	msg := ExampleFail{
		ReferenceFail: ReferenceFail{
			Ref:            "reference",
			RefSummary:     "ref summary",
			RefDescription: "ref description",
		},
		Summary:     "summary",
		Description: "description",
		Value:       nil,
	}

	err := sv.Validate(msg)
	got := err.Error()
	want := "min-length validation failed for field Ref"
	if got != want {
		t.Errorf("Expected: %s, got: %s", got, want)
	}
}

/*
*
Empty Struct Validation
*/
func TestEmptyStruct(t *testing.T) {
	type EmptyExample struct {
		Field string `json:"field" constraints:""`
	}
	msg := EmptyExample{}
	if err := sv.Validate(msg); err != nil {
		t.Errorf("Error in validation: %s", err)
	}

}

func TestConstStruct(t *testing.T) {
	type ConstExample struct {
		Summary     string `json:"summary" constraints:"-"`
		Description string `json:"description" constraints:""`
	}
	msg := ConstExample{
		Summary:     "testing",
		Description: "this is testing",
	}
	if err := sv.Validate(msg); err != nil {
		t.Errorf("Error in validation: %s", err)
	}
}

func TestEnumValidation(t *testing.T) {
	testsPass := []struct {
		Name  string
		input interface{}
	}{
		{
			Name: "Test-pass-1",
			input: struct {
				Status     string `json:"status" constraints:"enum=Success,Error,Not Reachable"`
				StatusCode int    `json:"statusCode" constraints:"enum=200,404,500"`
			}{Status: "Success", StatusCode: 200},
		},
	}

	for _, tt := range testsPass {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if err != nil {
				t.Errorf("Error in validation: %s", err)
			}
		})
	}

	testsError := []struct {
		Name  string
		input interface{}
		want  string
	}{
		{
			Name: "Test-fail-2",
			input: struct {
				Status2     string `json:"status2" constraints:"enum=Success,Error,Not Reachable"`
				StatusCode2 int    `json:"statusCode2" constraints:"enum=200,404,500"`
			}{Status2: "Success", StatusCode2: 503},
			want: "enum validation failed for field StatusCode2",
		},
	}

	for _, tt := range testsError {
		t.Run(tt.Name, func(t *testing.T) {
			err := sv.Validate(tt.input)
			if tt.want != err.Error() {
				t.Errorf("Got: %s, want: %s", err, tt.want)
			}
		})
	}
}

func TestCacheSuccess(t *testing.T) {
	withCache := validator.NewStructValidatorWithCache()

	// same structs with different fields will give the cached results on cached enabled
	testsWithCache := []struct {
		Name  string
		input interface{}
	}{
		{
			Name: "Test-pass-1",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:"min=10"`
				Mobile int    `json:"mobile" constraints:""`
			}{Name: "Testings", Age: 20, Mobile: 123456789},
		},
		{
			Name: "Test-pass-2",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:"min=10"`
				Mobile int    `json:"mobile" constraints:""`
			}{Name: "Testings", Age: 5, Mobile: 123456789},
		},
	}
	for _, tt := range testsWithCache {
		t.Run(tt.Name, func(t *testing.T) {
			err := withCache.Validate(tt.input)
			if err != nil {
				t.Errorf("Error in validation: %s", err)
			}
		})
	}
}

func TestCacheErrs(t *testing.T) {
	withoutCache := validator.NewStructValidator()

	// same structs with different values and caching disabled will parse field with each request
	testsWithoutCache := []struct {
		Name  string
		input interface{}
		want  interface{}
	}{
		{
			Name: "Test-fail-1",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:"min=10"`
				Mobile int    `json:"mobile" constraints:""`
			}{Name: "Testings", Age: 20, Mobile: 123456789},
			want: nil,
		},
		{
			Name: "Test-fail-2",
			input: struct {
				Name   string `json:"name" constraints:"min-length=5"`
				Age    int    `json:"age" constraints:"min=10"`
				Mobile int    `json:"mobile" constraints:""`
			}{Name: "Testings", Age: 5, Mobile: 123456789},
			want: "min value validation failed for field Age",
		},
	}

	for _, tt := range testsWithoutCache {
		t.Run(tt.Name, func(t *testing.T) {
			err := withoutCache.Validate(tt.input)
			if err != nil {
				if tt.want != err.Error() {
					t.Errorf("Got: %s, want: %s", err, tt.want)
				}
			}
		})
	}
}

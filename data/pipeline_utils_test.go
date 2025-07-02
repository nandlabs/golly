package data

import (
	"errors"
	"testing"
)

type mockPipeline map[string]any

func (m mockPipeline) Get(key string) (any, error) {
	v, ok := m[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

// Add Clone method to satisfy Pipeline interface
func (m mockPipeline) Clone() Pipeline {
	return m
}

// Add Delete method to satisfy Pipeline interface
func (m mockPipeline) Delete(key string) error {
	return nil
}

// Add GetError method to satisfy Pipeline interface
func (m mockPipeline) GetError() error {
	return nil
}

// Add Has method to satisfy Pipeline interface
func (m mockPipeline) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Add HasError method to satisfy Pipeline interface
func (m mockPipeline) HasError() bool {
	return false
}

// Add Id method to satisfy Pipeline interface
func (m mockPipeline) Id() string {
	return "mock"
}

// Add Keys method to satisfy Pipeline interface
func (m mockPipeline) Keys() []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Add Map method to satisfy Pipeline interface
func (m mockPipeline) Map() map[string]any {
	return m
}

// Correct Merge method to match Pipeline interface
func (m mockPipeline) Merge(other Pipeline) error {
	return nil
}

// Add MergeFrom method to satisfy Pipeline interface
func (m mockPipeline) MergeFrom(other map[string]any) error {
	return nil
}

// Add Set method to satisfy Pipeline interface
func (m mockPipeline) Set(key string, value any) error {
	m[key] = value
	return nil
}

// Add SetError method to satisfy Pipeline interface
func (m mockPipeline) SetError(err error) {}

func TestExtractValue_SimpleKey(t *testing.T) {
	p := mockPipeline{"foo": 42}
	v, err := ExtractValue[int](p, "foo")
	if err != nil || v != 42 {
		t.Errorf("expected 42, got %v, err=%v", v, err)
	}
}

func TestExtractValue_DotNotation(t *testing.T) {
	p := mockPipeline{"user": map[string]any{"city": "delhi"}}
	v, err := ExtractValue[string](p, "user.city")
	if err != nil || v != "delhi" {
		t.Errorf("expected delhi, got %v, err=%v", v, err)
	}
}

func TestExtractValue_ArrayIndex(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "address": map[string]any{"city": "blr"}},
		map[string]any{"name": "foo", "address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[1].address.city")
	if err != nil || v != "nyc" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_ArrayFilter(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "address": map[string]any{"city": "blr"}},
		map[string]any{"name": "foo", "address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[name==\"nanda\"].address.city")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_MapKey(t *testing.T) {
	users := map[string]any{
		"nanda": map[string]any{"address": map[string]any{"city": "blr"}},
		"foo":   map[string]any{"address": map[string]any{"city": "nyc"}},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[nanda].address.city")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_NotFound(t *testing.T) {
	p := mockPipeline{"foo": 42}
	_, err := ExtractValue[int](p, "bar")
	if err == nil {
		t.Error("expected error for missing key")
	}
}

func TestExtractValue_InvalidPath(t *testing.T) {
	p := mockPipeline{"foo": 42}
	_, err := ExtractValue[int](p, "foo.bar.baz")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestExtractValue_TypeMismatch(t *testing.T) {
	p := mockPipeline{"foo": "bar"}
	_, err := ExtractValue[int](p, "foo")
	if err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestExtractValue_StructArrayFilter(t *testing.T) {
	type Address struct{ City string }
	type User struct {
		Name    string
		Address Address
	}
	users := []User{{Name: "nanda", Address: Address{"blr"}}, {Name: "foo", Address: Address{"nyc"}}}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[Name==\"nanda\"].Address.City")
	if err != nil || v != "blr" {
		t.Errorf("expected blr, got %v, err=%v", v, err)
	}
}

func TestExtractValue_StructArrayIndex(t *testing.T) {
	type Address struct{ City string }
	type User struct {
		Name    string
		Address Address
	}
	users := []User{{Name: "nanda", Address: Address{"blr"}}, {Name: "foo", Address: Address{"nyc"}}}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[string](p, "users[1].Address.City")
	if err != nil || v != "nyc" {
		t.Errorf("expected nyc, got %v, err=%v", v, err)
	}
}

func TestExtractValue_ComplexFilter_Numeric(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "age": 30},
		map[string]any{"name": "foo", "age": 20},
		map[string]any{"name": "bar", "age": 40},
	}
	p := mockPipeline{"users": users}

	v, err := ExtractValue[map[string]any](p, "users[age>25]")
	if err != nil || v["name"] != "nanda" {
		t.Errorf("expected nanda, got %v, err=%v", v["name"], err)
	}
	v2, err := ExtractValue[map[string]any](p, "users[age<=20]")
	if err != nil || v2["name"] != "foo" {
		t.Errorf("expected foo, got %v, err=%v", v2["name"], err)
	}
}

func TestExtractValue_ComplexFilter_LogicalAndOr(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "age": 30, "city": "blr"},
		map[string]any{"name": "foo", "age": 20, "city": "nyc"},
		map[string]any{"name": "bar", "age": 40, "city": "blr"},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[map[string]any](p, "users[city==\"blr\" && age>35]")
	if err != nil || v["name"] != "bar" {
		t.Errorf("expected bar, got %v, err=%v", v["name"], err)
	}
	v2, err := ExtractValue[map[string]any](p, "users[city==\"nyc\" || age<25]")
	if err != nil || v2["name"] != "foo" {
		t.Errorf("expected foo, got %v, err=%v", v2["name"], err)
	}
}

func TestExtractValue_ComplexFilter_Grouping(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "age": 30, "city": "blr"},
		map[string]any{"name": "foo", "age": 20, "city": "nyc"},
		map[string]any{"name": "bar", "age": 40, "city": "blr"},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[map[string]any](p, "users[(city==\"blr\" && age<35) || (city==\"nyc\")]")
	if err != nil || v["name"] != "nanda" {
		t.Errorf("expected nanda, got %v, err=%v", v["name"], err)
	}
}

func TestExtractValue_ComplexFilter_NotEqual(t *testing.T) {
	users := []any{
		map[string]any{"name": "nanda", "city": "blr"},
		map[string]any{"name": "foo", "city": "nyc"},
	}
	p := mockPipeline{"users": users}
	v, err := ExtractValue[map[string]any](p, "users[city!=\"nyc\"]")
	if err != nil || v["name"] != "nanda" {
		t.Errorf("expected nanda, got %v, err=%v", v["name"], err)
	}
}

func TestExtractValue_NestedPathWithFilter(t *testing.T) {
	users := []any{
		map[string]any{
			"name": "nanda",
			"address": map[string]any{"city": "blr", "zip": 560001, "phones": []any{
				map[string]any{"type": "home", "number": "123"},
				map[string]any{"type": "work", "number": "456"},
			}},
		},
		map[string]any{
			"name": "foo",
			"address": map[string]any{"city": "nyc", "zip": 10001, "phones": []any{
				map[string]any{"type": "home", "number": "789"},
			}},
		},
	}
	p := mockPipeline{"users": users}

	// Filter at first level, then nested field
	v, err := ExtractValue[string](p, "users[name==\"foo\"].address.city")
	if err != nil || v != "nyc" {
		t.Errorf("expected nyc, got %v, err=%v", v, err)
	}

	// Filter at first and second level
	v2, err := ExtractValue[string](p, "users[name==\"nanda\"].address.phones[type==\"work\"].number")
	if err != nil || v2 != "456" {
		t.Errorf("expected 456, got %v, err=%v", v2, err)
	}

	// Filter at second level only
	v3, err := ExtractValue[string](p, "users[0].address.phones[type==\"home\"].number")
	if err != nil || v3 != "123" {
		t.Errorf("expected 123, got %v, err=%v", v3, err)
	}

	v4, err := ExtractValue[string](p, "users[address.zip>20000 && address.city==\"blr\"].name")
	if err != nil || v4 != "nanda" {
		t.Errorf("expected nanda, got %v, err=%v", v4, err)
	}
	v5, err := ExtractValue[string](p, "users[address.zip>20000].address.phones[type==\"home\"].number")
	if err != nil || v5 != "123" {
		t.Errorf("expected 123, got %v, err=%v", v5, err)
	}
	// Numeric filter at nested level
	v6, err := ExtractValue[string](p, "users[address.zip>20000 && address.phones[type==\"home\"].number==123].name")
	if err != nil || v6 != "nanda" {
		t.Errorf("expected 123, got %v, err=%v", v6, err)
	}
}

func TestEvaluateCondition(t *testing.T) {
	data := map[string]any{
		"age":    30,
		"name":   "nanda",
		"city":   "blr",
		"active": true,
		"user": map[string]any{
			"address": map[string]any{"city": "blr", "zip": 560001},
		},
		"scores": []any{10, 20, 30},
		"users": []any{
			map[string]any{"name": "nanda", "age": 30, "city": "blr", "phones": []any{
				map[string]any{"type": "home", "number": "123"},
				map[string]any{"type": "work", "number": "456"},
			}},
			map[string]any{"name": "alex", "age": 25, "city": "nyc", "phones": []any{
				map[string]any{"type": "home", "number": "678"},
				map[string]any{"type": "work", "number": "901"},
			}},
		},
	}
	p := mockPipeline{"": data}
	tests := []struct {
		cond     string
		expected bool
	}{
		{"age==30", true},
		{"age>25", true},
		{"age<25", false},
		{"name==\"nanda\"", true},
		{"city==\"nyc\"", false},
		{"user.address.city==\"blr\"", true},
		{"scores[1]==20", true},
		{"users[name==\"alex\"].city==\"nyc\"", true},
		{"users[age>28].name==\"nanda\"", true},
		{"users[city==\"blr\"].age==30", true},
		{"users[city==\"blr\"].age==25", false},
		{"age>=30 && city==\"blr\"", true},
		{"age>=30 && city==\"nyc\"", false},
		{"age>=30 || city==\"nyc\"", true},
		{"(age>=30 && city==\"nyc\") || (age<25)", false},
		{"users[city==\"blr\"].name==nanda && users[phones[type==\"home\"]].name==nanda", true},
	}
	for _, test := range tests {
		result := EvaluateCondition(p, test.cond)
		if result != test.expected {
			t.Errorf("EvaluateCondition(%q) = %v, want %v", test.cond, result, test.expected)
		}
	}
}

func TestSetValue_SimpleKey(t *testing.T) {
	p := mockPipeline{}
	err := SetValue(p, "foo", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := p.Get("foo")
	if v != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestSetValue_DotNotation(t *testing.T) {
	user := mockPipeline{"city": "delhi"}
	p := mockPipeline{"user": user}
	err := SetValue(p, "user.city", "blr")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := user.Get("city")
	if v != "blr" {
		t.Errorf("expected blr, got %v", v)
	}
}

func TestSetValue_ArrayFilter(t *testing.T) {
	users := []any{
		mockPipeline{"name": "nanda", "city": "blr"},
		mockPipeline{"name": "foo", "city": "nyc"},
	}
	p := mockPipeline{"users": users}
	err := SetValue(p, "users[name==\"foo\"].city", "sfo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if users[1].(mockPipeline)["city"] != "sfo" {
		t.Errorf("expected sfo, got %v", users[1].(mockPipeline)["city"])
	}
}

func TestSetValue_NestedPipeline(t *testing.T) {
	address := mockPipeline{"city": "blr"}
	user := mockPipeline{"address": address}
	p := mockPipeline{"user": user}
	err := SetValue(p, "user.address.city", "nyc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, _ := address.Get("city")
	if v != "nyc" {
		t.Errorf("expected nyc, got %v", v)
	}
}

func TestSetValue_NotPipelineError(t *testing.T) {
	p := mockPipeline{"foo": 123}
	err := SetValue(p, "foo.bar.test", 42)
	if err == nil {
		t.Error("expected error for non-pipeline final item")
	}
}

func TestSetValue_IntermediateMissing(t *testing.T) {
	p := mockPipeline{}
	err := SetValue(p, "foo.bar.test", 42)
	if err == nil {
		t.Error("expected error for non-pipeline final item")
	}
}

func TestSetValue_IntermediatePresent(t *testing.T) {
	p := mockPipeline{}
	p.Set("foo", mockPipeline{"bar": mockPipeline{}})
	err := SetValue(p, "foo.bar.test", 42)
	if err != nil {
		t.Error("Expected no error for valid path")
	}
	val, err := ExtractValue[int](p, "foo.bar.test")
	if err != nil && val != 42 {
		t.Errorf("expected 42, got %v, err=%v", val, err)
	}
}

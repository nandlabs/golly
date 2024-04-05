package secrets

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

type TestStore struct {
	providerName string
}

func (t *TestStore) Get(key string, ctx context.Context) (*Credential, error) {

	panic("implement me")
}

func (t *TestStore) Write(key string, credential *Credential, ctx context.Context) error {

	panic("implement me")
}

func (t *TestStore) Provider() string {
	return t.providerName
}

func TestGetManager(t *testing.T) {
	tests := []struct {
		name string
		want *Manager
	}{
		{name: "GetStoreMangerTest",
			want: &Manager{
				stores: nil,
				once:   sync.Once{},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetManager(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestManager_Register_Store(t *testing.T) {
	type fields struct {
		stores map[string]Store
	}

	var testStoreName = "TestStore"
	var want = &TestStore{providerName: testStoreName}
	tests := []struct {
		name   string
		fields fields
		want   *TestStore
	}{
		{
			name: "testStore",
			fields: fields{
				stores: make(map[string]Store),
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				stores: tt.fields.stores,
				once:   sync.Once{},
			}
			m.Register(tt.want)
			if got := m.Store(testStoreName); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetManager() = %v, want %v", got, tt.want)
			}
		})
	}
}

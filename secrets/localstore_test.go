package secrets

import (
	"context"
	"os"
	"path"
	"sync"
	"testing"
	"time"
)

func TestLocalStore_Get(t *testing.T) {

	type StoreStub struct {
		credentials map[string]*Credential
		storeFile   string
		masterKey   string
	}
	type args struct {
		key string
		ctx context.Context
	}
	tests := []struct {
		name     string
		fields   StoreStub
		args     args
		wantCred *Credential
		wantErr  bool
	}{
		{
			name: "testGet",
			fields: StoreStub{
				credentials: map[string]*Credential{
					"test": {
						Value:       []byte("testValue"),
						LastUpdated: time.Now(),
						Version:     "1.0",
						MetaData:    nil,
					},
				},
				storeFile: "test-File",
				masterKey: "master-key-0001",
			},
			args: args{
				key: "test",
				ctx: nil,
			},
			wantCred: &Credential{
				Value:       []byte("testValue"),
				LastUpdated: time.Now(),
				Version:     "1.0",
				MetaData:    nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &localStore{
				credentials: tt.fields.credentials,
				storeFile:   tt.fields.storeFile,
				masterKey:   tt.fields.masterKey,
				mutex:       sync.RWMutex{},
			}
			gotCred, err := ls.Get(tt.args.key, tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCred.Str() != tt.wantCred.Str() {
				t.Errorf("Get() gotCred = %v, want %v", gotCred.Str(), tt.wantCred.Str())
			}
		})
	}
}

func TestLocalStore_Provider(t *testing.T) {
	type fields struct {
		credentials map[string]*Credential
		storeFile   string
		masterKey   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{

		{
			name: "local-provider-test",
			fields: fields{
				credentials: nil,
				storeFile:   "",
				masterKey:   "",
			},
			want: LocalStoreProvider,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &localStore{
				credentials: tt.fields.credentials,
				storeFile:   tt.fields.storeFile,
				masterKey:   tt.fields.masterKey,
				mutex:       sync.RWMutex{},
			}
			if got := ls.Provider(); got != tt.want {
				t.Errorf("Provider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocalStore_Write(t *testing.T) {
	dir, err := os.Getwd()
	os.MkdirAll(path.Join(dir, "testdata"), os.ModePerm)
	type fields struct {
		credentials map[string]*Credential
		storeFile   string
		masterKey   string
	}
	type args struct {
		key        string
		credential *Credential
		ctx        context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "write-test",
			fields: fields{
				credentials: make(map[string]*Credential),
				storeFile:   path.Join(dir, "testdata", "test-store.dat"),
				masterKey:   "thisisamasterkey",
			},
			args: args{
				key: "test-key-01",
				credential: &Credential{
					Value:       []byte("test-value"),
					LastUpdated: time.Now(),
					Version:     "1.0",
					MetaData:    nil,
				},
				ctx: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ls := &localStore{
				credentials: tt.fields.credentials,
				storeFile:   tt.fields.storeFile,
				masterKey:   tt.fields.masterKey,
				mutex:       sync.RWMutex{},
			}
			if err = ls.Write(tt.args.key, tt.args.credential, tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	os.RemoveAll(path.Join(dir, "testdata"))
}

func TestNewLocalStore(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	testDataDir := path.Join(dir, "testdata")
	os.MkdirAll(testDataDir, os.ModePerm)

	// Clean up test data directory after all tests
	t.Cleanup(func() {
		os.RemoveAll(testDataDir)
	})

	type args struct {
		storeFile string
		masterKey string
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		setupFunc func() // Optional setup before test
	}{
		{
			name: "new-store-file-creation",
			args: args{
				storeFile: path.Join(testDataDir, "new-store.dat"),
				masterKey: "thisisamasterkeyABC", // 19 bytes - should pad to 16
			},
			wantErr: false,
		},
		{
			name: "existing-valid-store-file",
			args: args{
				storeFile: path.Join(testDataDir, "existing-store.dat"),
				masterKey: "thisisamasterkeyABC",
			},
			wantErr: false,
			setupFunc: func() {
				// Pre-create a valid store file with encrypted credentials
				validStore, _ := NewLocalStore(path.Join(testDataDir, "existing-store.dat"), "thisisamasterkeyABC")
				ls := validStore.(*localStore)
				cred := &Credential{
					Value:       []byte("test-secret"),
					LastUpdated: time.Now(),
					Version:     "1.0",
					MetaData:    nil,
				}
				ls.Write("test-key", cred, context.Background())
			},
		},
		{
			name: "invalid-master-key-length",
			args: args{
				storeFile: path.Join(testDataDir, "store-bad-key.dat"),
				masterKey: "short", // Only 5 bytes - invalid for AES
			},
			wantErr: false, // NewLocalStore doesn't fail on initialization, only on write
		},
		{
			name: "store-file-in-nonexistent-directory",
			args: args{
				storeFile: path.Join(testDataDir, "nonexistent", "subdir", "store.dat"),
				masterKey: "thisisamasterkeyABC",
			},
			wantErr: false, // Directory doesn't need to exist for NewLocalStore
		},
		{
			name: "corrupted-store-file",
			args: args{
				storeFile: path.Join(testDataDir, "corrupted-store.dat"),
				masterKey: "thisisamasterkeyABC",
			},
			wantErr: true, // Should fail when trying to decrypt corrupted file
			setupFunc: func() {
				// Create a corrupted file with invalid encrypted data
				os.WriteFile(
					path.Join(testDataDir, "corrupted-store.dat"),
					[]byte("this is not valid encrypted data"),
					0600,
				)
			},
		},
		{
			name: "load-credentials-with-correct-key",
			args: args{
				storeFile: path.Join(testDataDir, "correct-key-store.dat"),
				masterKey: "originalkeyph1234",
			},
			wantErr: false,
			setupFunc: func() {
				// Pre-create store with credentials
				originalStore, _ := NewLocalStore(path.Join(testDataDir, "correct-key-store.dat"), "originalkeyph1234")
				ls := originalStore.(*localStore)
				cred := &Credential{
					Value:       []byte("secret-value-here"),
					LastUpdated: time.Now(),
					Version:     "1.0",
				}
				ls.Write("test-key", cred, context.Background())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			gotLs, err := NewLocalStore(tt.args.storeFile, tt.args.masterKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLocalStore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotLs == nil {
				t.Errorf("NewLocalStore() expected a valid Store but got nil")
			}

			// Verify provider is set correctly
			if !tt.wantErr && gotLs.Provider() != LocalStoreProvider {
				t.Errorf("Provider() = %v, want %v", gotLs.Provider(), LocalStoreProvider)
			}
		})
	}
}

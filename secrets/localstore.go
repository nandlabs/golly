package secrets

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
)

const (
	LocalStoreProvider = "localStore"
)

// localStore will want the credential in a local file.
type localStore struct {
	credentials map[string]*Credential
	storeFile   string
	masterKey   string
	mutex       sync.RWMutex
}

func NewLocalStore(storeFile, masterKey string) (s Store, err error) {
	var fileContent []byte
	var decryptedContent []byte
	var credentials = make(map[string]*Credential)
	var fileInfo os.FileInfo
	fileInfo, err = os.Stat(storeFile)
	s = &localStore{
		credentials: credentials,
		storeFile:   storeFile,
		masterKey:   masterKey,
		mutex:       sync.RWMutex{},
	}
	if err == nil && !fileInfo.IsDir() {
		fileContent, err = os.ReadFile(storeFile)
		if err == nil {
			decryptedContent, err = AesDecrypt(fileContent, []byte(masterKey))
			if err == nil {
				decoder := gob.NewDecoder(bytes.NewReader(decryptedContent))
				err = decoder.Decode(&credentials)
			}
		}
	} else {
		err = nil
	}
	return
}

func (ls *localStore) Get(key string, ctx context.Context) (cred *Credential, err error) {
	ls.mutex.RLock()
	defer ls.mutex.RUnlock()
	if v, ok := ls.credentials[key]; ok {
		cred = v
	} else {
		err = fmt.Errorf("Unable to find a credential with key %s", key)
	}

	return
}

func (ls *localStore) Write(key string, credential *Credential, ctx context.Context) (err error) {
	ls.mutex.Lock()
	defer ls.mutex.Unlock()
	ls.credentials[key] = credential
	var b = &bytes.Buffer{}
	var encodedData []byte
	encoder := gob.NewEncoder(b)
	err = encoder.Encode(ls.credentials)
	if err == nil {
		encodedData, err = AesEncrypt([]byte(ls.masterKey), b.Bytes())
		if err == nil {
			err = os.WriteFile(ls.storeFile, encodedData, 0600)
		}
	}
	return
}

func (ls *localStore) Provider() string {
	return LocalStoreProvider
}

package secrets

import (
	"testing"
)

func TestAes(t *testing.T) {
	type args struct {
		key     string
		message string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "simple-message-16bit",
			args: args{
				key:     "This12BitKey0001",
				message: "This is a simple message",
			},
			wantErr: false,
		}, {
			name: "simple-message-24bit",
			args: args{
				key:     "This24BitKeyWillBeUsed01",
				message: "This is a simple message",
			},
			wantErr: false,
		},
		{
			name: "simple-message-32bit",
			args: args{
				key:     "thisisa32bitkeythisisa32bitkey02",
				message: "This is a simple message",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		var err error
		var encryptedMsg string
		var decryptedMsg string

		t.Run(tt.name, func(t *testing.T) {
			encryptedMsg, err = AesEncryptStr(tt.args.key, tt.args.message)

			if (err != nil) != tt.wantErr {
				t.Errorf("AesEncryptStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			decryptedMsg, err = AesDecryptStr(encryptedMsg, tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecryptStr() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if decryptedMsg != tt.args.message {
				t.Errorf("AesEncryptStr() gotEncrypted = %v, want %v", decryptedMsg, tt.args.message)
			}
		})
	}
}

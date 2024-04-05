package l3

import (
	"reflect"
	"testing"
)

// TestGetLogger --> Testing BaseLogger object creation
func TestGetLogger(t *testing.T) {
	tests := []struct {
		name string
		want Logger
	}{
		{
			name: "BaseLogger",
			want: Get(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Get(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogger_IsEnabled(t *testing.T) {
	type fields struct {
		level           Level
		pkgName         string
		errorEnabled    bool
		warnEnabled     bool
		infoEnabled     bool
		debugEnabled    bool
		traceEnabled    bool
		includeFunction bool
		includeLine     bool
	}
	type args struct {
		level Level
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		// TODO: Add test cases.
		{
			name: "WarnTest_true",
			fields: fields{
				level:       Warn,
				warnEnabled: true,
			},
			args: args{
				level: Warn,
			},
			want: true,
		},
		{
			name: "WarnTest_Fail",
			fields: fields{
				level:       Info,
				infoEnabled: true,
			},
			args: args{
				level: Warn,
			},
			want: false,
		},
		{
			name: "ErrorTest",
			fields: fields{
				level:        Err,
				errorEnabled: true,
			},
			args: args{
				level: Err,
			},
			want: true,
		},
		{
			name: "ErrorTest_Fail",
			fields: fields{
				level:        Trace,
				traceEnabled: true,
			},
			args: args{
				level: Err,
			},
			want: false,
		},
		{
			name: "InfoTest",
			fields: fields{
				level:       Info,
				infoEnabled: true,
			},
			args: args{
				level: Info,
			},
			want: true,
		},
		{
			name: "InfoTest_Fail",
			fields: fields{
				level:        Trace,
				traceEnabled: true,
			},
			args: args{
				level: Info,
			},
			want: false,
		},
		{
			name: "DebugTest",
			fields: fields{
				level:       Debug,
				warnEnabled: true,
			},
			args: args{
				level: Debug,
			},
			want: true,
		},
		{
			name: "DebugTest_Fail",
			fields: fields{
				level:        Trace,
				traceEnabled: true,
			},
			args: args{
				level: Debug,
			},
			want: false,
		},
		{
			name: "TraceTest",
			fields: fields{
				level:       Trace,
				warnEnabled: true,
			},
			args: args{
				level: Trace,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &BaseLogger{
				level:           tt.fields.level,
				pkgName:         tt.fields.pkgName,
				errorEnabled:    tt.fields.errorEnabled,
				warnEnabled:     tt.fields.warnEnabled,
				infoEnabled:     tt.fields.infoEnabled,
				debugEnabled:    tt.fields.debugEnabled,
				traceEnabled:    tt.fields.traceEnabled,
				includeFunction: tt.fields.includeFunction,
				includeLine:     tt.fields.includeLine,
			}
			if got := l.IsEnabled(tt.args.level); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

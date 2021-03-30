/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-03-30 10:12:21
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control_test.go
 * @Description:
 */

package flowcontrol

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test_defaultTafHash ...
func Test_defaultTafHash(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want uint32
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultTafHash(tt.args.key); got != tt.want {
				t.Errorf("defaultTafHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithForwardRadio(t *testing.T) {
	type args struct {
		r uint32
	}
	tests := []struct {
		name string
		args args
		want FlowControllerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithForwardRadio(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithForwardRadio() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestWithHashFunc ...
func TestWithHashFunc(t *testing.T) {
	type args struct {
		h func(string) uint32
	}
	tests := []struct {
		name string
		args args
		want FlowControllerOption
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithHashFunc(tt.args.h); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithHashFunc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlowController_Forward(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		f    FlowController
		args args
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Forward(tt.args.key); got != tt.want {
				t.Errorf("FlowController.Forward() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestNew ...
func TestNew(t *testing.T) {
	type args struct {
		opts []FlowControllerOption
	}
	tests := []struct {
		name string
		args args
		want *FlowController
	}{
		{
			name: "test1",
			args: args{[]FlowControllerOption{
				WithForwardRadio(10),
				WithHashFunc(defaultTafHash),
			}},
			want: &FlowController{
				Radio: 10,
				Hash:  defaultTafHash,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.opts...)
			assert.ObjectsAreEqualValues(got, tt.want)
		})
	}
}

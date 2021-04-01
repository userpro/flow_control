/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-04-01 20:01:42
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control_test.go
 * @Description:
 */

package flowcontrol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				WithForwardRadio([]uint32{10}),
				WithHashFunc(defaultTafHash),
			}},
			want: &FlowController{
				Radio: []uint32{10},
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

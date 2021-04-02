/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-04-02 16:13:43
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control_test.go
 * @Description:
 */

package flowcontrol

import (
	"math/rand"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandStringRunes 随机字符串生成器
func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
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
				WithForwardRadio([]uint64{10}),
				WithHashFunc(defaultTafHash),
			}},
			want: &FlowController{
				Radio: []uint64{10},
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

// TestFlowController_Forward1 ...
func TestFlowController_Forward1(t *testing.T) {
	type testData struct {
		a int
	}

	c1 := func(ch <-chan unsafe.Pointer) {
		for v := range ch {
			d := (*testData)(v)
			t.Logf("[c1] data.a(%d)", d.a)
		}
	}

	c2 := func(ch <-chan unsafe.Pointer) {
		for v := range ch {
			d := (*testData)(v)
			t.Logf("[c2] data.a(%d)", d.a)
		}
	}

	c3 := func(ch <-chan unsafe.Pointer) {
		for v := range ch {
			d := (*testData)(v)
			t.Logf("[c3] data.a(%d)", d.a)
		}
	}

	f := New(
		WithForwardRadio([]uint64{10, 20, 70}),
		WithEnableConsumer(true),
		WithConsumerBufferSize(10),
		WithConsumerFunc([]ConsumerFunc{c1, c2, c3}),
	)

	for i := 0; i < 10000; i++ {
		tmp := i
		f.Push(RandStringRunes(32), unsafe.Pointer(&testData{
			a: tmp,
		}))
	}

	time.Sleep(time.Second)
}

// TestFlowController_Forward2 ...
func TestFlowController_Forward2(t *testing.T) {
	type testData struct {
		a int
	}

	c1 := func(ch <-chan unsafe.Pointer) {
		for v := range ch {
			d := (*testData)(v)
			t.Logf("[c1] data.a(%d)", d.a)
		}
	}

	f := New(
		WithEnableConsumer(true),
		WithConsumerBucketNum(10),
		WithConsumerBufferSize(10),
		WithConsumerFunc([]ConsumerFunc{c1}),
	)

	for i := 0; i < 1000; i++ {
		tmp := i
		f.Push(RandStringRunes(32), unsafe.Pointer(&testData{
			a: tmp,
		}))
	}

	time.Sleep(time.Second)
}

// TestFlowController_Forward3 ...
func TestFlowController_Forward3(t *testing.T) {
	f := New(
		WithForwardRadio([]uint64{10, 10, 20}),
		WithEnableConsumer(false),
	)

	for i := 0; i < 1000; i++ {
		tmp := i
		t.Logf("id: %d, group: %d", tmp, f.Forward(RandStringRunes(32)))
	}
}

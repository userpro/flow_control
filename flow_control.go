/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-07-15 14:57:20
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control.go
 * @Description:
 */

package flowcontrol

import (
	"crypto/md5"
	"fmt"
	"sync"
	"unsafe"
)

var (
	defaultFlowControlOptions = flowControllerOptions{
		Radio:          []uint64{100},
		Hash:           defaultTafHash,
		EnableConsumer: false,
	}
)

// FlowController 流量控制
type FlowController interface {
	Forward(key string) int
	Push(key string, data unsafe.Pointer) int
}

// defaultTafHash
func defaultTafHash(key string) uint64 {
	md5Ctx := md5.New()
	md5Ctx.Reset()
	_, _ = md5Ctx.Write([]byte(key))
	cipherStr := md5Ctx.Sum(nil)
	hash := (int32(cipherStr[3]&0xFF) << 24) |
		(int32(cipherStr[2]&0xFF) << 16) |
		(int32(cipherStr[1]&0xFF) << 8) |
		(int32(cipherStr[0] & 0xFF))
	return uint64(hash)
}

// HashFunc 哈希函数原型
type HashFunc func(string) uint64

// ConsumerFunc 消费者函数原型
type ConsumerFunc func(ch <-chan unsafe.Pointer)

// FlowControllerOption 选项函数
type FlowControllerOption func(*flowControllerOptions)

// flowControllerOptions 可配置项
type flowControllerOptions struct {
	Radio              []uint64
	Hash               HashFunc
	EnableConsumer     bool
	ConsumerBufferSize uint64
	ConsumerBucketNum  uint64
	Consumer           []ConsumerFunc
}

// WithForwardRadio 划分比例
func WithForwardRadio(r []uint64) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.Radio = r
	}
}

// WithHashFunc hash函数
func WithHashFunc(h func(string) uint64) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.Hash = h
	}
}

// WithEnableConsumer 开启消费者模式
func WithEnableConsumer(ok bool) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.EnableConsumer = ok
	}
}

// WithConsumerBufferSize 消费者buffer大小
func WithConsumerBufferSize(size uint64) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.ConsumerBufferSize = size
	}
}

// WithConsumerBucketNum 消费者bucket数量
func WithConsumerBucketNum(num uint64) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.ConsumerBucketNum = num
	}
}

// WithConsumerFunc 消费者实现
func WithConsumerFunc(f []ConsumerFunc) FlowControllerOption {
	return func(fopt *flowControllerOptions) {
		fopt.Consumer = f
	}
}

type flowController struct {
	Radio              []uint64
	Hash               HashFunc
	EnableConsumer     bool
	ConsumerBufferSize uint64
	ConsumerBucketNum  uint64
	Consumer           []ConsumerFunc

	radio  []int
	mod    uint64
	buffer []chan unsafe.Pointer
	once   sync.Once
}

// Forward 是否转发
// 单纯函数 只用于进行流量划分
func (f *flowController) Forward(key string) int {
	return f.radio[f.Hash(key)%f.mod]
}

// Push 灌入数据
func (f *flowController) Push(key string, data unsafe.Pointer) int {
	k := f.Forward(key)
	f.buffer[k] <- data
	return len(f.buffer[k])
}

// consumerDo 消费者启动
func (f *flowController) consumerDo() {
	f.once.Do(
		func() {
			for i, ch := range f.buffer {
				go f.Consumer[i](ch)
			}
		},
	)
}

// initRadio 初始化分配比例
func (f *flowController) initRadio() {
	if f.Radio == nil {
		f.Radio = []uint64{100}
	}

	// 确保划分总额为100
	sum := uint64(0)
	for _, v := range f.Radio {
		sum += v
	}
	if sum < 100 {
		f.Radio = append(f.Radio, 100-sum)
	}
	if sum > 100 {
		panic(fmt.Sprintf("invalid radio sum(%d), limited 100", sum))
	}

	// 划分区间映射到数组
	f.radio = make([]int, 100)
	r := uint64(0)
	for count, v := range f.Radio {
		for i := r; i < r+v; i++ {
			f.radio[i] = count
		}
		r += v
	}

	f.mod = uint64(len(f.radio))
}

// initConsumer 初始化消费者
func (f *flowController) initConsumer() {
	if !f.EnableConsumer {
		return
	}

	if f.Consumer == nil || len(f.Consumer) == 0 {
		panic("enable but not set consumer")
	}

	if len(f.Radio) == 1 {
		// 全量
		f.radio = make([]int, f.ConsumerBucketNum)
		for i := range f.radio {
			f.radio[i] = i
		}
	} else if len(f.Radio) > 1 {
		// 如果设置大于1个划分区间
		f.ConsumerBucketNum = uint64(len(f.Radio))
	}

	f.buffer = make([]chan unsafe.Pointer, f.ConsumerBucketNum)
	for i := range f.buffer {
		f.buffer[i] = make(chan unsafe.Pointer, f.ConsumerBufferSize)
	}

	// 补齐消费者实现个数
	for len(f.Consumer) < len(f.buffer) {
		f.Consumer = append(f.Consumer, f.Consumer[0])
	}

	f.consumerDo()
}

// New ...
func New(opts ...FlowControllerOption) FlowController {
	options := defaultFlowControlOptions
	for _, o := range opts {
		o(&options)
	}

	f := &flowController{
		Radio:              options.Radio,
		Hash:               options.Hash,
		EnableConsumer:     options.EnableConsumer,
		ConsumerBufferSize: options.ConsumerBufferSize,
		ConsumerBucketNum:  options.ConsumerBucketNum,
		Consumer:           options.Consumer,
	}

	f.initRadio()
	f.initConsumer()

	return f
}

/*
 * @Author: dongzhzheng
 * @Date: 2021-03-29 16:45:44
 * @LastEditTime: 2021-04-01 20:36:32
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/flow_control.go
 * @Description:
 */

package flowcontrol

import (
	"crypto/md5"
	"sync"
	"unsafe"
)

var (
	defaultFlowControlOptions = FlowControllerOptions{
		Radio:          []uint32{100},
		Hash:           defaultTafHash,
		EnableConsumer: false,
	}

	md5Ctx = md5.New()
)

// defaultTafHash
func defaultTafHash(key string) uint64 {
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
type FlowControllerOption func(*FlowControllerOptions)

// FlowControllerOptions 可配置项
type FlowControllerOptions struct {
	Radio              []uint32
	Hash               HashFunc
	EnableConsumer     bool
	ConsumerBufferSize uint32
	ConsumerBucketNum  uint32
	Consumer           []ConsumerFunc
}

// WithForwardRadio 划分比例
func WithForwardRadio(r []uint32) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.Radio = r
	}
}

// WithHashFunc hash函数
func WithHashFunc(h func(string) uint64) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.Hash = h
	}
}

// WithEnableConsumer 开启消费者模式
func WithEnableConsumer(ok bool) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.EnableConsumer = ok
	}
}

// WithConsumerBufferSize 消费者buffer大小
func WithConsumerBufferSize(size uint32) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.ConsumerBufferSize = size
	}
}

// WithConsumerBucketNum 消费者bucket数量
func WithConsumerBucketNum(num uint32) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.ConsumerBucketNum = num
	}
}

// WithConsumerFunc 消费者实现
func WithConsumerFunc(f []ConsumerFunc) FlowControllerOption {
	return func(fopt *FlowControllerOptions) {
		fopt.Consumer = f
	}
}

// FlowController 流量控制
type FlowController struct {
	Radio              []uint32
	Hash               HashFunc
	EnableConsumer     bool
	ConsumerBufferSize uint32
	ConsumerBucketNum  uint32
	Consumer           []ConsumerFunc

	radio  []int
	buffer []chan unsafe.Pointer
	once   sync.Once
}

// Forward 是否转发
// 单纯函数 只用于进行流量划分
func (f *FlowController) Forward(key string) int {
	return f.radio[((f.Hash(key) & 0x7fffffff) % 100)]
}

// Push 灌入数据
func (f *FlowController) Push(key string, data unsafe.Pointer) int {
	k := f.Forward(key)
	f.buffer[k] <- data
	return len(f.buffer[k])
}

// consumerDo 消费者启动
func (f *FlowController) consumerDo() {
	f.once.Do(
		func() {
			for i, ch := range f.buffer {
				go f.Consumer[i](ch)
			}
		},
	)
}

// initRadio 初始化分配比例
func (f *FlowController) initRadio() {
	if f.Radio == nil {
		return
	}

	// 确保划分总额为100
	sum := uint32(0)
	for _, v := range f.Radio {
		sum += v
	}
	if sum < 100 {
		f.Radio = append(f.Radio, 100-sum)
	}
	// 如果划分区间只有一个 == 全量没有划分
	if len(f.Radio) == 1 {
		return
	}

	// 划分区间映射到数组
	f.radio = make([]int, 100)
	r := uint32(0)
	for count, v := range f.Radio {
		for i := r; i < r+v; i++ {
			f.radio[i] = count
		}
		r += v
	}
}

// initConsumer 初始化消费者
func (f *FlowController) initConsumer() {
	if !f.EnableConsumer ||
		f.Consumer == nil ||
		len(f.Consumer) == 0 {
		return
	}

	// 如果设置大于1个划分区间
	if f.Radio != nil && len(f.Radio) > 1 {
		f.ConsumerBucketNum = uint32(len(f.Radio))
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
func New(opts ...FlowControllerOption) *FlowController {
	options := defaultFlowControlOptions
	for _, o := range opts {
		o(&options)
	}

	f := &FlowController{
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

<!--
 * @Author: dongzhzheng
 * @Date: 2021-03-30 14:26:57
 * @LastEditTime: 2021-06-09 19:27:11
 * @LastEditors: dongzhzheng
 * @FilePath: /flow_control/README.md
 * @Description: 
-->

## flow_control

流量控制组件，具备如下功能：

* 流量划分（非随机采样划分，保证key稳定划分）
* 生产者消费者模式

## Demo(流量划分)

~~~go
// TestFlowController_Forward3 流量划分
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
~~~

## Demo(生产消费者模式)

~~~go
// TestFlowController_Forward1 生产消费者模式
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
		WithForwardRadio([]uint64{10, 20, 70}), // 当radio的和小于100时，剩余流量由c1消费
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
~~~

## Demo(流量划分+生产消费者模式)

~~~go
// TestFlowController_Forward2 流量划分+生产消费者模式
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
~~~
## Demo(结合datasource)

~~~go
/*
	生产者消费者模式，结合datasource库的演示
*/
import (
	"context"
	"hash/fnv"
	"unsafe"

	"git.code.oa.com/pcg_tkd_gnkf/datasource"
	"git.code.oa.com/trpc-go/trpc-go/log"

	report "git.woa.com/pcg_tkd_gnkf/reporter"

	kafkawrapper "git.code.oa.com/pcg_tkd_gnkf/datasource/pkg/kafkaWrapper"
	flowcontrol "git.code.oa.com/pcg_tkd_gnkf/flow_control"

	goutils "github.com/userpro/go-utils"
)

var dsKafka datasource.Datasource

/* 省略dsKafka的初始化 */

// distributeData 分发的数据
type distributeData struct {
	ctx  context.Context
	data *kafkawrapper.Data
}

// 生产者
func producer(ctx context.Context) {
	q := []*kafkawrapper.Query{{
		Data: make(chan []*kafkawrapper.Data, 1000),
		Ctx:  make(chan context.Context, 1000),
		Ack:  make(chan error, 1000),
	}}

	if err := dsKafka.Read(ctx, q); err != nil {
		log.ErrorContextf(ctx, "kafka read err(%v)", err)
		return
	}

	// 设置hash函数
	h := fnv.New64a()
	hashFunc := func(key string) uint64 {
		h.Reset()
		if n, err := h.Write(goutils.Str2Bytes(key)); err != nil {
			log.ErrorContextf(ctx, "guid to fnv64a n(%d) err(%d)", n, err)
		}

		return h.Sum64()
	}

	// flowcontrol实例
	fc := flowcontrol.New(
		flowcontrol.WithHashFunc(hashFunc),                                 // 设置自定义hash函数
		flowcontrol.WithEnableConsumer(true),                               // 启用生产者消费者模式
		flowcontrol.WithConsumerBucketNum(uint32(10)),                      // 消费者个数
		flowcontrol.WithConsumerBufferSize(uint32(1000)),                   // 每个消费者的buffer大小
		flowcontrol.WithConsumerFunc([]flowcontrol.ConsumerFunc{consumer}), // 消费者函数
	)

	for kData := range q[0].Data {
		// 忽略kafka context
		<-q[0].Ctx

		// kafka commit
		// 数据异常告警, 异常排除后kafka灌入正常数据消费即可
		q[0].Ack <- nil

		for _, msg := range kData {
			d := &distributeData{
				ctx:  ctx,
				data: msg,
			}
			// 返回值为队列长度，Push方法线程安全
			queueLen := fc.Push(goutils.Bytes2Str(msg.Key), unsafe.Pointer(d))
			// 监控上报
			report.PropertyAvg("req-bucket_taskq_length", float64(queueLen))
		}
	}
}

// 消费者
func consumer(ch <-chan unsafe.Pointer) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("err(%+v) stack(%s)", err, goutils.DumpStacks(2048))
		}
	}()

	for d := range ch {
		packet := (*distributeData)(d)
		doTask(packet.ctx, packet.data)
	}
}

// doTask 业务代码
func doTask(ctx context.Context, msg *kafkawrapper.Data) {
	// TODO ...
}

~~~
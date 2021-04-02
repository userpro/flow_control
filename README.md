<!--
 * @Author: dongzhzheng
 * @Date: 2021-03-30 14:26:57
 * @LastEditTime: 2021-04-02 11:16:38
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
/*
	流量划分
*/

~~~
## Demo(生产者消费者模式)

~~~go
/*
	生产者消费者模式，结合datasource库的演示
*/
import (
	"context"
	"hash/fnv"
	"unsafe"

	"git.code.oa.com/trpc-go/trpc-go/log"

	kafkawrapper "git.code.oa.com/pcg_tkd_gnkf/datasource/pkg/kafkaWrapper"
	flowcontrol "git.code.oa.com/pcg_tkd_gnkf/flow_control"

	goutils "github.com/userpro/go-utils"
)

// distributeData 分发的数据
type distributeData struct {
	ctx  context.Context
	data *kafkawrapper.Data
}

// 生产者
func producer(ctx context.Context) {
	q := []*kafkawrapper.Query{{
		Data: make(chan []*kafkawrapper.Data, taskConfig.KafkaConsumerBuffSize),
		Ctx:  make(chan context.Context, taskConfig.KafkaConsumerBuffSize),
		Ack:  make(chan error, taskConfig.KafkaConsumerBuffSize),
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
		flowcontrol.WithHashFunc(hashFunc),   // 设置自定义hash函数
		flowcontrol.WithEnableConsumer(true), // 启用生产者消费者模式
		flowcontrol.WithConsumerBucketNum(uint32(10)),    // 消费者个数
		flowcontrol.WithConsumerBufferSize(uint32(1000)), // 每个消费者的buffer大小
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
<!--
 * @Author: dongzhzheng
 * @Date: 2021-03-30 14:26:57
 * @LastEditTime: 2021-04-02 16:21:42
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

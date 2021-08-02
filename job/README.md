# Job

Job 是对异步任务的抽象。该组件依赖 RabbitMQ 实现异步消息的排队和消费。
这里采用了发布订阅模型，使用一个队列发布消息，可以启动多个消费者同时订阅以增加并发量。
建议将同一类型的任务放到同一个队列。

## 定义一个任务

可以使用任意类型来表示一个任务，只要满足实现了 `Job` 接口。JobManager 在发送任务前，会调用 `Marshal`
函数对任务数据序列化，以此作为消息体发送到消息队列中。

```go
const MyJobQueueName = "my_job_queue_name"

type MyJob struct {
    data string `json:"data"`
}

func (m *MyJob) Queue() string {
    return MyJobQueueName
}

func (m *MyJob) Delay() int {
    return 0 // no delay
}

func (m *MyJob) Marshal() ([]byte, error) {
	return json.Marshal(c)
}
```

## 定义一个任务处理器

任务处理器可以是任意 `JobHandler` 形式的函数。在函数内可以使用 `Jobber` 来访问每一个接收到的任务数据，
并根据处理情况对其做相应的处理（确认、重试等）。以下示例代码展示了 `Jobber` 的基本用法，省略了任务处理逻辑。

```go
func MyJobHandler(ctx context.Context, jobber Jobber) error {
    job := new(MyJob)
	err := json.Unmarshal(jobber.Body(), job)
    if err != nil {
        return jobber.Skip() // 忽略掉无法解析的消息
    }

    // 当前任务重试次数小于10次
    if jobber.Attempt() < 10 {
        // 重试当前任务，时间间隔300秒
        return jobber.Retry(context.Background(), 300)
    }

    // 处理成功
    return jobber.OK()
}
```

## 启动JobManager

```go
manager := NewJobManager(JobManagerOptions{
    // RabbitMQ 连接参数
})

// 发布一个任务
job := &MyJob{"whatever"}
manager.Dispatch(context.Background(), job)

// 另外启动一个进程或线程消费任务
// 注意：这个函数是阻塞的，直到调用 manager.Close() 才会退出
manager.Do(context.Background(), MyJobQueueName, MyJobHandler)
```

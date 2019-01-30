# golang 实现多个worker去解决大量的任务

- 可以设置worker在队列空的时候结束
- 提供友好的协程同步控制机制
- 可以灵活的自定义job

## 使用方式

`
    group := gworker.NewWorkerGroup(10)
    group.Start()

    for i := 1; i <= 10000000; i++ {
        job := jobs.NewJob([]byte(fmt.Sprintf("job %d", i)))
        group.DispatchJob(job)
    }

    group.WaitEmpty()
`

## 函数方法

### WorkerGroup

`
    投递任务
    DispatchJob(job interfaces.Job)

    开启任务
    Start()

    等待结束
    Wait()

    取消为空时循环等待
    WaitEmpty()

`









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

    开启任务队列分发
    Start()

    等待结束,只有在向worker的队列中发送 stopjob 才结束
    Wait()

    在队列为空时退出
    WaitEmpty()

`









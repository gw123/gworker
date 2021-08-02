package job

import (
	"context"
	"fmt"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/rabbiter"
	"sync"

	"github.com/pkg/errors"
)

// JobManagerOptions JobManager 的 Rabbiter 配置
type JobManagerOptions struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Vhost    string `mapstructure:"vhost"`
}

// BuildURL 构建连接 RabbitMQ 的 URL
func (o *JobManagerOptions) BuildURL() string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		o.Username, o.Password, o.Host, o.Port, o.Vhost)
}

// JobManager
type JobManager struct {
	r rabbiter.Factory
	p rabbiter.Publisher

	mux    sync.RWMutex
	closed bool
	cs     []rabbiter.Consumer
}

type Jobber struct {
	rabbiter.Jobber
}

func (j *Jobber) Body() []byte {
	return j.Jobber.Message().Body
}

// NewJobManager 创建 JobManager 实例
func NewJobManager(opts JobManagerOptions) *JobManager {
	r := rabbiter.NewRabbiter(rabbiter.Options{
		Host: opts.BuildURL(),
	})

	p := r.Publisher()
	return &JobManager{
		r:      r,
		p:      p,
		cs:     []rabbiter.Consumer{},
		closed: false,
	}
}

// Dispatch 序列化 Job ，发送到消息队列
func (m *JobManager) Dispatch(ctx context.Context, job gworker.Job) error {
	body, err := job.Marshal()
	if err != nil {
		return errors.Wrap(err, "job manager failed to marshal job")
	}

	queueName := job.Queue()
	exchange := &rabbiter.Exchange{
		Name:        queueName,
		Type:        "direct",
		AutoDeleted: false,
		Durable:     true,
	}
	message := &rabbiter.Message{
		Body:       body,
		RoutingKey: queueName,
	}

	delay := job.Delay()
	if delay > 0 {
		queue := &rabbiter.Queue{
			Name:       queueName,
			Durable:    true,
			RoutingKey: queueName,
		}
		err = m.p.PublishWithDelay(ctx, queue, message, delay)
	} else {
		err = m.p.Publish(ctx, exchange, message)
	}
	if err != nil {
		return errors.Wrap(err, "job manager failed to publish job")
	}

	return nil
}

// Do 从消息队列消费消息并交给 JobHandler 处理。这是一个阻塞函数。
func (m *JobManager) Do(ctx context.Context, queueName string, handler gworker.JobHandler) error {
	m.mux.RLock()
	if m.closed {
		m.mux.RUnlock()
		return errors.New("job manager closed")
	}
	m.mux.RUnlock()

	exchange := &rabbiter.Exchange{Name: queueName, Type: "direct", Durable: true}
	queue := &rabbiter.Queue{Name: queueName, Durable: true, RoutingKey: queueName}

	c := m.r.Consumer(exchange, queue, func(ctx context.Context, job rabbiter.Jobber) error {
		return handler(ctx, &Jobber{job})
	})
	if err := c.Listen(ctx); err != nil {
		return errors.Wrapf(err, "job manager could not listen to queue %s", queueName)
	}

	m.mux.Lock()
	m.cs = append(m.cs, c)
	m.mux.Unlock()

	return c.Wait()
}

// Do 从消息队列消费消息并交给 JobHandler 处理。这是一个阻塞函数。
func (m *JobManager) DoJobs(ctx context.Context, job gworker.Job) error {
	m.mux.RLock()
	if m.closed {
		m.mux.RUnlock()
		return errors.New("job manager closed")
	}
	m.mux.RUnlock()

	exchange := &rabbiter.Exchange{Name: job.Queue(), Type: "direct", Durable: true}
	queue := &rabbiter.Queue{Name: job.Queue(), Durable: true, RoutingKey: job.Queue()}

	c := m.r.Consumer(exchange, queue, func(ctx context.Context, rawJob rabbiter.Jobber) error {
		return job.JobHandler(ctx, &Jobber{rawJob})
	})

	if err := c.Listen(ctx); err != nil {
		return errors.Wrapf(err, "job manager could not listen to queue %s", job.Queue())
	}

	m.mux.Lock()
	m.cs = append(m.cs, c)
	m.mux.Unlock()
	return c.Wait()
}

// Close 关闭当前阻塞的消费者
func (m *JobManager) Close() error {
	m.mux.Lock()
	defer m.mux.Unlock()

	for _, c := range m.cs {
		c.Stop()
	}
	m.closed = true
	m.cs = m.cs[:0]

	return nil
}

package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gw123/glog"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/mocks"
	"github.com/gw123/gworker/rabbiter"
	rabbiterMocks "github.com/gw123/gworker/rabbiter/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"testing"
)

func TestJobManagerOptionsBuilURL(t *testing.T) {
	cases := []struct {
		opts   JobManagerOptions
		expect string
	}{
		{
			opts:   JobManagerOptions{"guest", "guest", "localhost", 5672, "/"},
			expect: "amqp://guest:guest@localhost:5672//",
		},
		{
			opts:   JobManagerOptions{"guest", "pass", "10.0.0.4", 15672, "/abc"},
			expect: "amqp://guest:pass@10.0.0.4:15672//abc",
		},
	}

	for _, c := range cases {
		actual := c.opts.BuildURL()
		assert.Equal(t, c.expect, actual, "URL should match")
	}
}

func TestJobManagerDispatch(t *testing.T) {
	p := &rabbiterMocks.Publisher{}
	m := &JobManager{p: p}

	ctx := context.Background()
	body := []byte("this is message body")
	queue := "queue_name"

	job := &mocks.Job{}
	job.On("Marshal").Return(body, nil)
	job.On("Queue").Return(queue)
	job.On("Delay").Return(0)

	p.On("Publish", ctx,
		&rabbiter.Exchange{Name: queue, Type: "direct", Durable: true},
		&rabbiter.Message{Body: body, RoutingKey: queue}).
		Return(nil)

	err := m.Dispatch(ctx, job)
	assert.Nil(t, err, "should not return error")
}

func TestJobManagerDispatchDelay(t *testing.T) {
	p := &rabbiterMocks.Publisher{}
	m := &JobManager{p: p}

	ctx := context.Background()
	body := []byte("this is message body")
	queue := "queue_name"
	delay := 10

	job := &mocks.Job{}
	job.On("Marshal").Return(body, nil)
	job.On("Queue").Return(queue)
	job.On("Delay").Return(delay)

	p.On("PublishWithDelay", ctx,
		&rabbiter.Queue{Name: queue, RoutingKey: queue, Durable: true},
		&rabbiter.Message{Body: body, RoutingKey: queue},
		delay).
		Return(nil)

	err := m.Dispatch(ctx, job)
	assert.Nil(t, err, "should not return error")
}

func TestJobManagerDo(t *testing.T) {
	f := &rabbiterMocks.Factory{}
	m := &JobManager{r: f, cs: []rabbiter.Consumer{}}

	ctx := context.Background()
	queue := "queue_name"
	handler := func(ctx context.Context, job gworker.Jobber) error {
		return nil
	}

	hfType := fmt.Sprintf("%T", rabbiter.HandlerFunc(nil))

	c := &rabbiterMocks.Consumer{}
	f.On("Consumer",
		&rabbiter.Exchange{Name: queue, Type: "direct", Durable: true},
		&rabbiter.Queue{Name: queue, Durable: true, RoutingKey: queue},
		mock.AnythingOfType(hfType),
	).Return(c).Once()

	c.On("Listen", ctx).Return(nil).Once()
	c.On("Wait").Return(nil).Once()

	m.Do(ctx, queue, handler)

	assert.Equal(t, 1, len(m.cs), "consumer should be under management")
}

func TestJobManagerClose(t *testing.T) {
	c := &rabbiterMocks.Consumer{}
	m := &JobManager{cs: []rabbiter.Consumer{c}}
	c.On("Stop").Return(nil).Once()

	m.Close()
	assert.Equal(t, 0, len(m.cs), "consumer queue should be empty")
	assert.True(t, m.closed, "job manager status should be closed")
}

type MyJob struct {
	Uuid       string    `json:"uuid"`
	CreatedAt  time.Time `json:"created_at"`
	MockTry    int       `json:"mock_try"`
	CurrentTry int       `json:"current_try"`
}

func NewMyJob() *MyJob {

	return &MyJob{
		Uuid:      uuid.New().String(),
		CreatedAt: time.Now(),
	}
}

func (m MyJob) UUID() string {
	return m.Uuid
}

func (m MyJob) Queue() string {
	return "my_job"
}

func (m MyJob) Delay() int {
	return 3
}

func (m MyJob) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m MyJob) JobHandler(ctx context.Context, job gworker.Jobber) error {
	tmpJob := MyJob{}
	if err := json.Unmarshal(job.Body(), &tmpJob); err != nil {
		return err
	}

	glog.DefaultLogger().Infof("MyJob UUID: %s, CreatedAt: %s",
		tmpJob.UUID(), tmpJob.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func TestConsumerAndProductor(t *testing.T) {
	opts := JobManagerOptions{"test", "test", "localhost", 5672, "/"}

	jobManager := NewJobManager(opts)
	ctx := context.Background()

	go func() {
		if err := jobManager.DoJobs(ctx, MyJob{}); err != nil {
			glog.DefaultLogger().Error(err)
		}
	}()
	time.Sleep(time.Second)

	for i := 0; i < 10000; i++ {
		jb1 := NewMyJob()
		if err := jobManager.Dispatch(ctx, jb1); err != nil {
			glog.DefaultLogger().Error(err)
		}
	}

	time.Sleep(10 * time.Second)
	if err := jobManager.Close(); err != nil {
		glog.DefaultLogger().Error(err)
	}
}

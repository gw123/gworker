package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/google/uuid"

	"github.com/gw123/glog"
	"github.com/gw123/gworker"
	"github.com/gw123/gworker/job"
)

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
	if err := json.Unmarshal(job.Body(), &m); err != nil {
		return err
	}

	glog.DefaultLogger().Infof("MyJob UUID: %s, CreatedAt: %s",
		m.UUID(), m.CreatedAt.Format("2006-01-02 15:04:05"))
	return nil
}

func main() {
	e := os.Environ()
	spew.Dump(e)

	//opts := job.JobManagerOptions{"18618184632", "2020passwd@gw1992", "sh2", 5672, "/"}
	username := os.Getenv("Username")
	password := os.Getenv("Password")
	host := os.Getenv("Host")

	glog.DefaultLogger().Info(username, password, host)
	opts := job.JobManagerOptions{username, password, host, 5672, "/"}

	jobManager := job.NewJobManager(opts)
	ctx := context.Background()

	go func() {
		job := MyJob{}
		consumer, err := jobManager.Do(ctx, job.Queue(), job.JobHandler)
		if err != nil {
			glog.DefaultLogger().Error(err)
		}
		go func() {
			time.Sleep(time.Second * 6)
			consumer.Stop()
			glog.Info("consumer stop")
		}()

		consumer.Wait()

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

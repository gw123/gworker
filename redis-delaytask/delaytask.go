package gworker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gw123/glog/common"
	"sync"
	"time"
	"unicode"

	goredis "github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
)

/*
本模块提供延迟处理任务功能，新增功能无需修改此文件，只要在自己业务包内新增代码即可。
添加一个新的延迟处理类型的流程：
1. 新建一个类型，如 XxxDelayTask，内嵌 TaskType
2. 实现 DelayDoer 接口
3. 在 init 函数中调用 RegisterUnmarshalFunc 对新功能注册 unmarshal 函数
4. 调用 AddDelayTask 添加延迟任务
5. 延迟任务将在 eventLoopDelayTask 线程中被执行、删除
*/

const (
	REDIS_WAITING_DELAY_TASK_ZSET = "waiting_delay_task_zset"
	REDIS_RUNNING_DELAY_TASK_ZSET = "running_delay_task_zset"
	REDIS_DELAY_TASK_HASH         = "delay_task_hash"
)

//弹出小于指定时间的第一个元素
const ZsetPopByScoreLua = `
local key = KEYS[1]
local currentTime = ARGV[1]
local list = redis.call("ZRANGEBYSCORE", key, "-inf", currentTime,"LIMIT", 0,1) 
for i = 1, #list do
	 redis.call('ZREM', key, list[i])
end
return list
`

var ZPopByScoreLuaScript = goredis.NewScript(ZsetPopByScoreLua)
var emptyErr = errors.New("empty")

// DelayDoer 执行延迟任务接口
type DelayDoer interface {
	Do(context.Context, common.Logger) error // 执行任务
	GetUUID() string                         // 获取任务的唯一标示
}

// DelayTasker 延迟任务接口
type DelayTasker interface {
	DelayDoer
	Done(context.Context, common.Logger) // 删除任务，发生问题输出 log，继续执行
}

// TaskType 延迟任务的类型
type TaskType struct {
	Type string // 任务类型，序列化时必填
}

// Task 延迟任务通用信息
type Task struct {
	TaskType
	doer    DelayDoer
	taskStr string // redis key 执行任务前必填，以便删除 redis
	deleted bool   // 是否已删除
}

// Do 执行任务
func (p *Task) Do(ctx context.Context, logger common.Logger) error {
	if p == nil || p.doer == nil || ctx == nil || logger == nil {
		return fmt.Errorf("nil pointer")
	}
	return p.doer.Do(ctx, logger)
}

// Done 删除任务，发生问题输出 log，继续执行
func (p *Task) Done(ctx context.Context, logger common.Logger) {
	if p == nil {
		logger.Errorf("nil pointer")
	}

	if p.deleted {
		logger.Errorf("delay task %v already been deleted", logger)
		return
	}
	p.deleted = true
}

func (p *Task) GetUUID() string {
	return p.doer.GetUUID()
}

var gUnmarshalFuncMap = make(map[string]func(string, common.Logger) (DelayDoer, error))

type DelayJobManager struct {
	redisClient  *goredis.Client
	prefix       string
	loggerClient common.Logger
	flag         bool
}

func NewDelayJobManager(redisClient *goredis.Client, loggerClient common.Logger, prefix string) *DelayJobManager {
	return &DelayJobManager{redisClient: redisClient, loggerClient: loggerClient, prefix: prefix, flag: true}
}

func (d *DelayJobManager) getRedisWaitingDelayTaskZsetName() string {
	return d.prefix + REDIS_WAITING_DELAY_TASK_ZSET
}

func (d *DelayJobManager) getRedisRunningDelayTaskZsetName() string {
	return d.prefix + REDIS_RUNNING_DELAY_TASK_ZSET
}

func (d *DelayJobManager) getRedisDelayTaskHashName() string {
	return d.prefix + REDIS_DELAY_TASK_HASH
}

// 任务结束从正在运行队列删除指定的任务
func (d *DelayJobManager) deleteDelayTask(taskStr string) error {
	err := d.redisClient.ZRem(d.getRedisRunningDelayTaskZsetName(), taskStr)
	if err != nil {
		return errors.Errorf("zrem %+v from %v error: %v", taskStr, d.getRedisRunningDelayTaskZsetName(), err)
	}

	err = d.redisClient.HDel(d.getRedisDelayTaskHashName(), taskStr)
	if err != nil {
		return errors.Errorf("hdel %+v from %v error: %v", taskStr, d.getRedisDelayTaskHashName(), err)
	}
	return nil
}

// 解析出来任务
func (d *DelayJobManager) GetTask(taskKey string) (DelayTasker, error) {
	body, err := d.redisClient.HGet(d.getRedisDelayTaskHashName(), taskKey).Result()
	if err != nil {
		return nil, errors.Errorf("hdel %+v from %v error: %v", taskKey, d.getRedisDelayTaskHashName(), err)
	}

	tmp := &TaskType{}
	if err := json.Unmarshal([]byte(body), tmp); err != nil {
		return nil, err
	}

	// 获得具体 unmarshal 函数
	unmarshalFunc, ok := gUnmarshalFuncMap[tmp.Type]
	if !ok {
		return nil, errors.Errorf("no unmarshalfunc for delay task type: %v", tmp.Type)
	}

	// 解析出具体类型
	doer, err := unmarshalFunc(body, d.loggerClient)
	if err != nil {
		return nil, errors.Errorf("unmarshal delay task %v type %v error: %v", body, tmp.Type, err)
	}

	return &Task{
		TaskType: TaskType{
			Type: tmp.Type,
		},
		taskStr: body,
		doer:    doer,
	}, nil
}

//弹出小于指定时间的第一个元素
func (d *DelayJobManager) popTaskKeyByScore(key string, score int64) (string, error) {
	val, err := ZPopByScoreLuaScript.Eval(d.redisClient, []string{key}, score).Result()
	if err != nil {
		return "", err
	}
	if arr, ok := val.([]interface{}); ok {
		if len(arr) <= 0 {
			return "", emptyErr
		} else {
			return arr[0].(string), nil
		}
	}
	return "", errors.New("invalid val")
}

// GetDelayTasks 获得符合执行时间的任务，函数保证获取的 Task 不为 nil
func (d *DelayJobManager) popWaitingDelayTask() (DelayTasker, error) {
	curTimeMS := time.Now().UnixNano() / int64(time.Millisecond)
	taskKey, err := d.popTaskKeyByScore(d.getRedisWaitingDelayTaskZsetName(), curTimeMS)
	if err != nil {
		return nil, err
	}
	taskDoer, err := d.GetTask(taskKey)
	if err != nil {
		return nil, err
	}

	return taskDoer, nil
}

// GetDelayTasks 获得已经运行但是长时间没有心跳的任务
func (d *DelayJobManager) popRunningFailDelayTask() (DelayTasker, error) {
	// 假如任务20秒没有执行完 并且没有更新时间戳则认为这个任务已经异常
	curTimeMS := time.Now().Add(-time.Second*20).UnixNano() / int64(time.Millisecond)
	taskKey, err := d.popTaskKeyByScore(d.getRedisRunningDelayTaskZsetName(), curTimeMS)
	if err != nil {
		return nil, err
	}
	return d.GetTask(taskKey)
}

// Run 执行任务
func (d *DelayJobManager) Run(ctx context.Context) {
	go func() {
		for d.flag {
			task, err := d.popWaitingDelayTask()
			if err == emptyErr {
				time.Sleep(time.Second)
				continue
			}
			if err != nil {
				d.loggerClient.Errorf("popWaitingDelayTask %v error: %v", d.getRedisWaitingDelayTaskZsetName(), err)
				time.Sleep(time.Second)
				continue
			}
			d.loggerClient.Infof("popWaitingDelayTask new task %+v", task)

			z := &goredis.Z{
				Score:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
				Member: task.GetUUID(),
			}
			// 添加任务到执行的队列
			err = d.redisClient.ZAdd(d.getRedisRunningDelayTaskZsetName(), z).Err()
			if err != nil {
				d.loggerClient.Errorf("zadd delay task %v to %v error: %v", task, d.getRedisWaitingDelayTaskZsetName(), err)
				time.Sleep(time.Second)
				continue
			}

			func() {
				// 执行完成删除任务
				defer func() {
					if err := recover(); err != nil {
						d.loggerClient.Errorf("deleteDelayTask %+v recover %s", task, err)
					}

					if err := d.deleteDelayTask(task.GetUUID()); err != nil {
						d.loggerClient.Errorf("deleteDelayTask %+v err %s", task, err)
					}
					d.loggerClient.Infof("deleteDelayTask %s", task.GetUUID())
				}()

				// 执行任务
				wg := sync.WaitGroup{}
				ch := make(chan bool, 1)
				wg.Add(2)
				go func() {
					defer wg.Done()
					if err := task.Do(ctx, d.getLogger()); err != nil {
						d.loggerClient.Errorf("%v do delay task: %#v error: %v", task, err)
					}
					ch <- true
					task.Done(ctx, d.loggerClient)
					//d.loggerClient.Infof("Do over %+v", task)
				}()

				// 执行任务过程中不断更新运行时间戳防止任务被其他worker抢占
				go func() {
					defer wg.Done()
					stop := false
					for !stop {
						select {
						case <-time.Tick(time.Second * 3):
							d.loggerClient.Debugf("更新定时器 %+v", task.GetUUID())
							z := &goredis.Z{
								Score:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
								Member: task.GetUUID(),
							}
							err = d.redisClient.ZAdd(d.getRedisRunningDelayTaskZsetName(), z).Err()
							if err != nil {
								d.loggerClient.Errorf("zadd delay task %v to %v error: %v", task, d.getRedisWaitingDelayTaskZsetName(), err)
							}
						case <-ch:
							//d.loggerClient.Infof("更新定时器携程结束1 %+v", task)
							stop = true
						}
					}
					//d.loggerClient.Infof("更新时间戳携程结束2 %s", task.GetUUID())
				}()
				wg.Wait()

			}()
		}
	}()

	d.monitor()
}

func (d *DelayJobManager) monitor() error {
	for d.flag {
		time.Sleep(time.Second)
		task, err := d.popRunningFailDelayTask()
		// 说明所有任务正常执行
		if err == emptyErr {
			continue
		}

		if err != nil {
			d.loggerClient.Errorf("popRunningFailDelayTask err %v", err)
			continue
		}

		// 任务异常重新投递队列让其他worker执行
		d.loggerClient.Infof("任务异常重新投递队列让其他worker执行 %v", task)
		if err := d.deleteDelayTask(task.GetUUID()); err != nil {
			d.loggerClient.Errorf("addFailDelayTask to deleteDelayTask err %v", err)
		}

		if err := d.AddDelayTask(time.Now(), task); err != nil {
			d.loggerClient.Errorf("addFailDelayTask to AddDelayTask err %v", err)
		}
	}
	return nil
}

func (d *DelayJobManager) getLogger() common.Logger {
	return d.loggerClient
}

// AddDelayTask 添加延迟任务
func (d *DelayJobManager) AddDelayTask(t time.Time, taskDoer DelayDoer) error {
	buf, err := json.Marshal(taskDoer)
	if err != nil {
		return fmt.Errorf("marshal delay task %+v error: %v", taskDoer, err)
	}
	z := &goredis.Z{
		Score:  float64(time.Now().UnixNano() / int64(time.Millisecond)),
		Member: taskDoer.GetUUID(),
	}
	// score 参数填充毫秒级时间戳
	err = d.redisClient.ZAdd(d.getRedisWaitingDelayTaskZsetName(), z).Err()
	if err != nil {
		return fmt.Errorf("zadd delay task %v to %v error: %v", taskDoer, d.getRedisWaitingDelayTaskZsetName(), err)
	}

	err = d.redisClient.HSet(d.getRedisDelayTaskHashName(), taskDoer.GetUUID(), buf).Err()
	if err != nil {
		return fmt.Errorf("hset delay task %v to %v error: %v", taskDoer, d.getRedisDelayTaskHashName(), err)
	}
	return nil
}

func IsValidFunctionName(name string) error {
	if name == "" {
		return fmt.Errorf("function name is illegal")
	}
	for _, r := range name {
		if !unicode.IsLetter(r) {
			return fmt.Errorf("function name contains illegal char '%v'", r)
		}
	}
	return nil
}

func RegisterUnmarshalFunc(taskType string, unmarshalFunc func(string, common.Logger) (DelayDoer, error)) error {
	if err := IsValidFunctionName(taskType); err != nil {
		return errors.Errorf("type name \"%v\" check error: %v", taskType, err)
	} else if _, ok := gUnmarshalFuncMap[taskType]; ok {
		return errors.Errorf("already register type: %v", taskType)
	}
	gUnmarshalFuncMap[taskType] = unmarshalFunc
	return nil
}

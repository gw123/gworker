package gworker

import (
	"encoding/json"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

const CacheTaskName = "update-cache"

type CacheTask struct {
	ComId    uint
	Channel  string
	TaskName string

	Template string
	Phone    string
	Params   []string
}

func NewCacheTask(comId uint, channel, template, phone string, params []string) *CacheTask {
	return &CacheTask{
		ComId:    comId,
		Channel:  channel,
		TaskName: CacheTaskName,
		Template: template,
		Phone:    phone,
		Params:   params,
	}
}

func (S *CacheTask) GetTaskName() string {
	return CacheTaskName
}

func (S *CacheTask) ToJson() string {
	data, err := json.Marshal(S)
	if err != nil {
		return ""
	}
	return string(data)
}

func (S *CacheTask) GetHandleFun() interface{} {
	return func(data string) error {
		newTask := new(CacheTask)
		err := json.Unmarshal([]byte(data), newTask)
		if err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}
		if len(newTask.Params) == 0 {
			return errors.Wrap(err, "缺少参数")
		}
		glog.Infof("发送短信到%s,模板：%s,参数:%s", newTask.Phone, newTask.Template, newTask.Params[0])
		return nil
	}
}

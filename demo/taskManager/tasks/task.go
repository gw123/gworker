package tasks

import (
	"encoding/json"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

type SMSTask struct {
	TaskName string
	Template string
	Phone    string
	Params   []string
}

func NewSMSTask(phone string, params []string) *SMSTask {
	return &SMSTask{
		TaskName: "sms",
		Template: "a001",
		Phone:    phone,
		Params:   params,
	}
}

func (S *SMSTask) GetTaskName() string {
	return "sms"
}

func (S *SMSTask) ToJson() string {
	data, err := json.Marshal(S)
	if err != nil {
		return ""
	}
	return string(data)
}

func (S *SMSTask) GetHandleFun() interface{} {
	return func(data string) error {
		newTask := new(SMSTask)
		err := json.Unmarshal([]byte(data), newTask)
		if err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}
		glog.Infof("ON SMS TASK %s", data)
		if len(newTask.Params) == 0 {
			glog.Errorf("缺少参数 %s", data)
			return errors.Wrap(err, "缺少参数")
		}
		glog.Infof("发送短信到%s,模板：%s,参数:%s", newTask.Phone, newTask.Template, newTask.Params[0])
		return nil
	}
}

package gworker

import (
	"encoding/json"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

const SMSTaskName = "sms"

type SMSTask struct {
	ComId    uint
	Channel  string
	TaskName string

	Type          string
	PhoneNumbers  []string
	TemplateParam string
}

func NewSMSTask(comId uint, channel, template string, phones []string, params string) *SMSTask {
	return &SMSTask{
		ComId:         comId,
		Channel:       channel,
		TaskName:      SMSTaskName,
		Type:          template,
		PhoneNumbers:  phones,
		TemplateParam: params,
	}
}

func (S *SMSTask) GetTaskName() string {
	return SMSTaskName
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
		if len(newTask.TemplateParam) == 0 {
			return errors.Wrap(err, "缺少参数")
		}
		glog.Infof("发送短信到%+v,模板：%s,参数:%s", newTask.PhoneNumbers, newTask.Type, newTask.TemplateParam)
		return nil
	}
}

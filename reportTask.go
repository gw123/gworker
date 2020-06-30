package gworker

import (
	"encoding/json"
	"github.com/gw123/glog"
	"github.com/pkg/errors"
)

const ReportTaskName = "sms"

type ReportTask struct {
	ComId    uint
	Channel  string
	TaskName string

	Template string
	Phone    string
	Params   []string
}

func NewReportTask(comId uint, channel, template, phone string, params []string) *ReportTask {
	return &ReportTask{
		ComId:    comId,
		Channel:  channel,
		TaskName: ReportTaskName,
		Template: template,
		Phone:    phone,
		Params:   params,
	}
}

func (S *ReportTask) GetTaskName() string {
	return ReportTaskName
}

func (S *ReportTask) ToJson() string {
	data, err := json.Marshal(S)
	if err != nil {
		return ""
	}
	return string(data)
}

func (S *ReportTask) GetHandleFun() interface{} {
	return func(data string) error {
		newTask := new(ReportTask)
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

package smsTask

import (
	"github.com/gw123/glog"
	"github.com/pkg/errors"
	"time"
)

const SMSTaskName = "sms"

type SMSTask struct {
	ComId    uint
	Id    string
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

func (s *SMSTask) GetID() string {
	return s.Id
}

func (s *SMSTask) Trace() []string {
     return nil
}

func (s *SMSTask) GetName() string {
	return SMSTaskName
}

func (s *SMSTask) RetryCount() int  {
	return 2
}

func (s *SMSTask) Delay() time.Duration  {
	return time.Second * 2
}

func (s *SMSTask) Handle( ) error {
	if len(s.TemplateParam) == 0 {
		return errors.New( "缺少参数")
	}
	glog.Infof("发送短信到%+v,模板：%s,参数:%s", s.PhoneNumbers, s.Type, s.TemplateParam)
	return nil
}

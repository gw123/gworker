package gworker

type Options struct {
	Broker        string      `yaml:"broker" mapstructure:"broker"`
	DefaultQueue  string      `yaml:"default_queue" mapstructure:"default_queue"`
	ResultBackend string      `yaml:"result_backend" mapstructure:"result_backend"`
	AMQP          *AMQPOptions `yaml:"amqp" mapstructure:"amqp"`
}

/***
   AMQPOptions
   ExchangeType  交换机类型 ,当交换机类型是direct时候在投递任务队列使用BingingKey作为队列名,
				 其他情况默认发送到DefaultQueue使用BingingKey作为路由
   PrefetchCount 预先拉取任务数量
 */
type AMQPOptions struct {
	Exchange      string `yaml:"exchange" `
	ExchangeType  string `yaml:"exchange_type"  mapstructure:"exchange_type"`
	BindingKey    string `yaml:"binding_key" mapstructure:"binding_key"`
	PrefetchCount int    `yaml:"prefetch_count" mapstructure:"prefetch_count"`
	AutoDelete    bool   `yaml:"auto_delete" mapstructure:"auto_delete"`
}

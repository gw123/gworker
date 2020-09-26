package gworker

type Options struct {
	Broker        string      `yaml:"broker" mapstructure:"broker"`
	DefaultQueue  string      `yaml:"default_queue" mapstructure:"default_queue"`
	ResultBackend string      `yaml:"result_backend" mapstructure:"result_backend"`
	AMQP          *AMQPOptions `yaml:"amqp" mapstructure:"amqp"`
}

type AMQPOptions struct {
	Exchange      string `yaml:"exchange" `
	ExchangeType  string `yaml:"exchange_type"  mapstructure:"exchange_type"`
	BindingKey    string `yaml:"binding_key" mapstructure:"binding_key"`
	PrefetchCount int    `yaml:"prefetch_count" mapstructure:"prefetch_count"`
	AutoDelete    bool   `yaml:"auto_delete" mapstructure:"auto_delete"`
}

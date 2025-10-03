package eventbus

type SendOpt func(option *SendOption)

type SendOption struct {
	ShardingKey *string
}

func WithShardingKey(key string) SendOpt {
	return func(o *SendOption) {
		o.ShardingKey = &key
	}
}

type ConsumerOpt func(option *ConsumerOption)

type ConsumerOption struct {
	Orderly *bool
}

func WithConsumerOrderly(orderly bool) ConsumerOpt {
	return func(option *ConsumerOption) {
		option.Orderly = &orderly
	}
}

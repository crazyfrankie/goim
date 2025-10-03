package rmq

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	
	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/safego"
	"github.com/crazyfrankie/goim/types/consts"
)

type producerImpl struct {
	nameServer string
	topic      string
	p          rocketmq.Producer
}

func NewProducer(nameServer, topic, group string, retries int) (eventbus.Producer, error) {
	if nameServer == "" {
		return nil, fmt.Errorf("name server is empty")
	}

	if topic == "" {
		return nil, fmt.Errorf("topic is empty")
	}

	p, err := rocketmq.NewProducer(
		producer.WithNsResolver(primitive.NewPassthroughResolver([]string{nameServer})),
		producer.WithRetry(retries),
		producer.WithGroupName(group),
		producer.WithCredentials(primitive.Credentials{
			AccessKey: os.Getenv(consts.RMQAccessKey),
			SecretKey: os.Getenv(consts.RMQSecretKey),
		}),
		// producer.WithNsResolver(primitive.NewGRPCCredentialsResolver(nil)),
		// producer.WithInstanceName("rocketmq-cnngf291ea363b7a"),
	)
	if err != nil {
		return nil, fmt.Errorf("NewProducer failed, nameServer: %s, topic: %s, err: %w", nameServer, topic, err)
	}

	err = p.Start()
	if err != nil {
		return nil, fmt.Errorf("start producer error: %w", err)
	}

	safego.Go(context.Background(), func() {
		signal.WaitExit()
		if err := p.Shutdown(); err != nil {
			logs.Errorf("shutdown producer error: %s", err.Error())
		}
	})

	return &producerImpl{
		nameServer: nameServer,
		topic:      topic,
		p:          p,
	}, nil
}

func (r *producerImpl) Send(ctx context.Context, body []byte, opts ...eventbus.SendOpt) error {
	_, err := r.p.SendSync(context.Background(), primitive.NewMessage(r.topic, body))
	if err != nil {
		return fmt.Errorf("[producerImpl] send message failed: %w", err)
	}
	return err
}

func (r *producerImpl) BatchSend(ctx context.Context, bodyArr [][]byte, opts ...eventbus.SendOpt) error {
	option := eventbus.SendOption{}
	for _, opt := range opts {
		opt(&option)
	}

	var msgArr []*primitive.Message
	for _, body := range bodyArr {
		msg := primitive.NewMessage(r.topic, body)

		if option.ShardingKey != nil {
			msg.WithShardingKey(*option.ShardingKey)
		}

		msgArr = append(msgArr, msg)
	}

	_, err := r.p.SendSync(ctx, msgArr...)
	return err
}

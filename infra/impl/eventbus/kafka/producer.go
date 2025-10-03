package kafka

import (
	"context"

	"github.com/IBM/sarama"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/safego"
)

type producerImpl struct {
	topic string
	p     sarama.SyncProducer
}

func NewProducer(broker, topic string) (eventbus.Producer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		return nil, err
	}

	safego.Go(context.Background(), func() {
		signal.WaitExit()
		if err := producer.Close(); err != nil {
			logs.Errorf("close producer error: %s", err.Error())
		}
	})

	return &producerImpl{
		topic: topic,
		p:     producer,
	}, nil
}

func (r *producerImpl) Send(ctx context.Context, body []byte, opts ...eventbus.SendOpt) error {
	return r.BatchSend(ctx, [][]byte{body}, opts...)
}

func (r *producerImpl) BatchSend(ctx context.Context, bodyArr [][]byte, opts ...eventbus.SendOpt) error {
	option := eventbus.SendOption{}
	for _, opt := range opts {
		opt(&option)
	}

	var msgArr []*sarama.ProducerMessage
	for _, body := range bodyArr {
		msg := &sarama.ProducerMessage{
			Topic: r.topic,
			Value: sarama.ByteEncoder(body),
		}

		if option.ShardingKey != nil {
			msg.Key = sarama.StringEncoder(*option.ShardingKey)
		}

		msgArr = append(msgArr, msg)
	}

	err := r.p.SendMessages(msgArr)
	if err != nil {
		return err
	}

	return nil
}

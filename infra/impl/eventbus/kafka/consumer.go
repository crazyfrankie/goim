package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/crazyfrankie/goim/infra/contract/eventbus"
	"github.com/crazyfrankie/goim/pkg/lang/signal"
	"github.com/crazyfrankie/goim/pkg/logs"
	"github.com/crazyfrankie/goim/pkg/safego"
)

type consumerImpl struct {
	broker        string
	topic         string
	groupID       string
	handler       eventbus.ConsumerHandler
	consumerGroup sarama.ConsumerGroup
}

func RegisterConsumer(broker string, topic, groupID string, handler eventbus.ConsumerHandler, opts ...eventbus.ConsumerOpt) error {
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // Start consuming from the earliest message
	config.Consumer.Group.Session.Timeout = 30 * time.Second

	o := &eventbus.ConsumerOption{}
	for _, opt := range opts {
		opt(o)
	}
	// TODO: orderly

	consumerGroup, err := sarama.NewConsumerGroup([]string{broker}, groupID, config)
	if err != nil {
		return err
	}

	c := &consumerImpl{
		broker:        broker,
		topic:         topic,
		groupID:       groupID,
		handler:       handler,
		consumerGroup: consumerGroup,
	}

	ctx := context.Background()
	safego.Go(ctx, func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, c); err != nil {
				logs.Errorf("consumer group consume: %v", err)
				break
			}
		}
	})

	safego.Go(ctx, func() {
		signal.WaitExit()

		if err := c.consumerGroup.Close(); err != nil {
			logs.Errorf("consumer group close: %v", err)
		}
	})

	return nil
}

func (c *consumerImpl) Setup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerImpl) Cleanup(sess sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerImpl) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()

	for msg := range claim.Messages() {
		m := &eventbus.Message{
			Topic: msg.Topic,
			Group: c.groupID,
			Body:  msg.Value,
		}
		if err := c.handler.HandleMessage(ctx, m); err != nil {
			continue
		}

		sess.MarkMessage(msg, "") // TODO: Consumer policies can be configured
	}
	return nil
}

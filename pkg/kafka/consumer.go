package kafka

import (
	"context"
	"errors"
	"sync"

	"github.com/Shopify/sarama"
	log "github.com/sirupsen/logrus"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewConsumer(ctx context.Context, addrs []string, groupName string, version string, opts ...Option) (s *Consumer, err error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// config
	config := sarama.NewConfig()
	config.Version, err = sarama.ParseKafkaVersion(version)
	if err != nil {
		return nil, err
	}
	// 分区分配策略
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true
	if defaultOpts.enableSASL {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = defaultOpts.username
		config.Net.SASL.Password = defaultOpts.password
	}

	s = &Consumer{}
	s.ctx, s.cancel = context.WithCancel(ctx)

	s.consumer, err = sarama.NewConsumerGroup(addrs, groupName, config)
	if err != nil {
		return nil, err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		for {
			select {
			case err := <-s.consumer.Errors():
				log.WithError(err).Error()
			case <-s.ctx.Done():
				return
			}
		}
	}()

	return s, nil
}

func (c *Consumer) Close() {
	c.cancel()
	if err := c.consumer.Close(); err != nil {
		log.WithError(err).Error()
	}
	c.wg.Wait()
	log.Infof("consumer closed")
}

func (c *Consumer) Consumer(topics []string, handler sarama.ConsumerGroupHandler) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			if err := c.consumer.Consume(c.ctx, topics, handler); err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.WithError(err).Error()
			}
			log.Infof("rebalance happens, the consumer session will need to be recreated")
		}
	}()
}

package kafka_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/limingyao/excellent-go/pkg/kafka"
	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

type consumerHandler struct {
	limiter *rate.Limiter
	wg      sync.WaitGroup
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Infof("setup")
	h.limiter = rate.NewLimiter(1, 10)
	return nil
}

func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Infof("cleanup")
	h.wg.Wait()
	return nil
}

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.limiter.Wait(context.Background()); err == nil {
			h.wg.Add(1)
			go func(message *sarama.ConsumerMessage) {
				defer h.wg.Done()

				log.Infof("topic %s, partition: %d, offset: %d", message.Topic, message.Partition, message.Offset)
				log.Infof("message claimed: value [%s], timestamp: %v", message.Value, message.Timestamp)
			}(message)
			session.MarkMessage(message, "")
		} else {
			log.WithError(err).Error()
		}
	}

	return nil
}

func TestConsumer_Consumer(t *testing.T) {
	c, err := kafka.NewConsumer(context.Background(), []string{"dev.machine:9092"}, "test-group", sarama.V0_10_2_1.String())
	if err != nil {
		t.Error(err)
		return
	}

	handler := &consumerHandler{}
	c.Consumer([]string{"test-topic"}, handler)

	time.Sleep(5 * time.Minute)
}

func TestConsumer_Close(t *testing.T) {
	c, err := kafka.NewConsumer(context.Background(), []string{"dev.machine:9092"}, "test-group", sarama.V0_10_2_1.String())
	if err != nil {
		t.Error(err)
		return
	}

	handler := &consumerHandler{}
	c.Consumer([]string{"test-topic"}, handler)

	time.Sleep(10 * time.Second)
	c.Close()

	time.Sleep(time.Minute)
}

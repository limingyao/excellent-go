package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Shopify/sarama"
)

type Producer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

func NewProducer(addrs []string, opts ...Option) (s *Producer, err error) {
	defaultOpts := defaultOptions
	for _, o := range opts {
		o.apply(&defaultOpts)
	}

	// config
	config := sarama.NewConfig()
	// only wait for the leader to ack
	config.Producer.RequiredAcks = sarama.WaitForAll
	if defaultOpts.enableCompression {
		config.Producer.Compression = sarama.CompressionSnappy
	}
	config.Producer.Return.Successes = true
	// flush batches every 100 ms
	config.Producer.Flush.Frequency = defaultOpts.flushFrequency
	config.Producer.MaxMessageBytes = defaultOpts.maxMessageBytes
	if defaultOpts.enableSASL {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = defaultOpts.username
		config.Net.SASL.Password = defaultOpts.password
	}

	s = &Producer{}

	s.syncProducer, err = sarama.NewSyncProducer(addrs, config)
	if err != nil {
		return nil, err
	}
	s.asyncProducer, err = sarama.NewAsyncProducer(addrs, config)
	if err != nil {
		return nil, err
	}
	go s.startAsyncMonitor()

	return s, nil
}

func (s *Producer) ProduceSync(ctx context.Context, topic, key string, data interface{}) (partition int32, offset int64, err error) {
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	default:
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return 0, 0, err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(buf),
	}

	return s.syncProducer.SendMessage(msg)
}

func (s *Producer) ProduceAsync(ctx context.Context, topic, key string, data interface{}) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(buf),
	}

	s.asyncProducer.Input() <- msg

	return nil
}

func (s *Producer) startAsyncMonitor() {
	for {
		select {
		case msg := <-s.asyncProducer.Successes():
			log.Printf("produced message success, partition %d offset %d", msg.Partition, msg.Offset)
		case err := <-s.asyncProducer.Errors():
			log.Printf("produced message err, topic: %s, msg: [%s], err: %v", err.Msg.Topic, err.Msg.Value, err.Error())
		case <-time.After(60 * time.Second):
			// 超时策略，避免kafka没有消息后一直等待的问题
			log.Printf("kafka async producer wait time out %d s", 60)
		}
	}
}

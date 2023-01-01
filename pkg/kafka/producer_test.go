package kafka_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/limingyao/excellent-go/pkg/kafka"
	"github.com/stretchr/testify/assert"
)

func TestProducer_ProduceSync(t *testing.T) {
	ast := assert.New(t)

	kfk, err := kafka.NewProducer([]string{"dev.machine:9092"})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		_, _, err := kfk.ProduceSync(context.Background(), "test-topic", fmt.Sprintf("%d", i), "hello world")
		ast.Nil(err)
	}
}

func TestProducer_ProduceAsync(t *testing.T) {
	ast := assert.New(t)

	kfk, err := kafka.NewProducer([]string{"dev.machine:9092"})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := kfk.ProduceAsync(context.Background(), "test-topic", fmt.Sprintf("%d", i), "hello world")
		ast.Nil(err)
	}
	time.Sleep(1 * time.Second)
}

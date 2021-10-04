package kafka

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestProducer_ProduceSync(t *testing.T) {
	assert := assert.New(t)

	kfk, err := NewProducer([]string{"dev.machine:9092"})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 100; i++ {
		_, _, err := kfk.ProduceSync(context.Background(), "test-topic", fmt.Sprintf("%d", i), "hello world")
		assert.Nil(err)
	}
}

func TestProducer_ProduceAsync(t *testing.T) {
	assert := assert.New(t)

	kfk, err := NewProducer([]string{"dev.machine:9092"})
	if err != nil {
		t.Error(err)
		return
	}

	for i := 0; i < 10; i++ {
		err := kfk.ProduceAsync(context.Background(), "test-topic", fmt.Sprintf("%d", i), "hello world")
		assert.Nil(err)
	}
	time.Sleep(1 * time.Second)
}

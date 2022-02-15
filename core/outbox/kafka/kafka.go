package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/iyarkov2/chat/core/outbox"
	"github.com/iyarkov2/chat/core/util"
)

const (
	TaskType = "KafkaPublish"
)

type Config struct {
	bootstrapServers string
}

type Worker struct {
	producer *kafka.Producer
}

func NewProducer(config Config) (Worker, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":       config.bootstrapServers,
	})
	if err != nil {
		return Worker{}, fmt.Errorf("failed to create producer: %w", err)
	}
	return Worker{ producer: producer }, nil
}

func (p Worker) Do(ctx context.Context, task outbox.Task) error {
	if task.Type != TaskType {
		return fmt.Errorf("unsupported task type %s", task.Type)
	}
	msg, ok := task.Metadata.(*kafka.Message)
	if !ok {
		return fmt.Errorf("unsupported metadata type %T", task.Metadata)
	}

	deliveryChannel := make(chan kafka.Event, 1)
	if err := p.producer.Produce(msg, deliveryChannel); err != nil {
		return fmt.Errorf("publish failed %w", err)
	}

	// Wait for delivery report
	e := <- deliveryChannel
	message := e.(*kafka.Message)
	if message.TopicPartition.Error != nil {
		return fmt.Errorf("publish failed: %w\n", message.TopicPartition.Error)
	}

	log := util.GetLogger(ctx)
	log.Debug().Msgf("message published to topic:%s partition: %d, offset: %d", message.TopicPartition.Topic, message.TopicPartition.Partition, message.TopicPartition.Offset)
	return nil
}
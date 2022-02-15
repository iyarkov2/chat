package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"time"
)

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		panic(fmt.Errorf("failed to create Kafka producer, %w", err))
	}

	timeTicker := time.NewTicker(5 * time.Second)

	topic := "tick"
	key := []byte("timer_1")
	callback := make(chan kafka.Event)

	for {
		ts := <- timeTicker.C
		fmt.Printf("Time to publishins ")
		msg := kafka.Message {
			TopicPartition: kafka.TopicPartition {
				Topic : &topic,
				Partition: kafka.PartitionAny,
			},
			Value: []byte(time.Now().String()),
			Key: key,
		}
		if err = producer.Produce(&msg, callback); err != nil {
			panic(fmt.Errorf("publish failed, %w", err))
		}

		e := <-callback
		message := e.(*kafka.Message)
		if message.TopicPartition.Error != nil {
			panic(fmt.Errorf("publish failed: %w\n", message.TopicPartition.Error))
		}
		fmt.Printf("done at %s\n", ts)
	}
}



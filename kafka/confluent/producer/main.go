package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/google/uuid"
	app "github.com/iyarkov2/chat/kafka"
	"time"
)

var producer *confluent.Producer

func main() {
	app.WithTimer("Create Producer", func() error {
		var err error
		producer, err = kafka.NewProducer(app.ConfluentConfig)
		return err
	})

	// Go-routine to handle message delivery reports and
	// possibly other event types (errors, stats, etc)
	go func() {
		for e := range producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					app.Log.Info().Msgf("Failed to deliver message: %v\n", ev.TopicPartition)
				} else {
					app.Log.Info().Msgf("Successfully produced record to topic %s partition [%d] @ offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()

	app.ExperimentSerial(publish)

	// Wait for all messages to be delivered
	producer.Flush(15 * 1000)

}

func publish() {
	id := uuid.New()
	app.WithTimer("Publish " + id.String(), func() error {
		ch := make(chan confluent.Event)
		topic := app.Topic
		headers := []confluent.Header {
			{
				Key: "ce_id",
				Value: []byte(id.String()),
			},
		}
		msg := kafka.Message {
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny },
			Value: []byte(time.Now().String()),
			Headers: headers,
		}
		err := producer.Produce(&msg, ch)
		if err != nil {
			return err
		}
		ev := <- ch
		app.Log.Info().Msgf("Published %s", ev)
		return nil
	})
}


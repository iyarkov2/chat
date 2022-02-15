package main

import (
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	app "github.com/iyarkov2/chat/kafka"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var consumer *confluent.Consumer

func main() {
	app.WithTimer("Create Consumer", func() error {
		// Create Consumer instance
		var err error
		consumer, err = kafka.NewConsumer(app.ConfluentConfig)
		return err
	})

	app.WithTimer("Subscribe", func() error {
		return consumer.SubscribeTopics([]string{app.Topic}, nil)
	})

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Process messages
	run := true
	for run == true {
		select {
		case sig := <-sigchan:
			fmt.Printf("Caught signal %v: terminating\n", sig)
			run = false
		default:
			msg, err := consumer.ReadMessage(100 * time.Millisecond)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					// app.Log.Info().Msg("Still listening")
				} else {
					app.Log.Error().Msg("Consumer Error")
				}
				continue
			}

			msgId := "Unknown"
			for _, h := range msg.Headers {
				if h.Key == "ce_id" {
					msgId = string(h.Value)
				}
			}

			app.Log.Info().Msgf("Received %s, -> [%s]from %d partition, offset: %d, key %s", msgId, msg.Value, msg.TopicPartition.Partition, msg.TopicPartition.Offset, msg.Key)
		}
	}

	fmt.Printf("Closing consumer\n")

	app.Must("Closing consumer", func() error{
		return consumer.Close()
	})
}

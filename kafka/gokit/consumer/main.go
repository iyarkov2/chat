package main

import (
	"context"
	events "github.com/cloudevents/sdk-go/v2"
	app "github.com/iyarkov2/chat/kafka"
	"time"
)

func main() {
	app.WithGokit("gokit-consumer")

	app.Must("Register Consumer", func() error {
		return app.Broker.RegisterSubscription(app.Topic, sub{})
	})

	app.WithTimer("Start Broker", func() error {
		return app.Broker.Start()
	})

	for {
		time.Sleep(10 * time.Second)
		app.Log.Info().Msg("Still running")
	}
}

type sub struct {

}

func (s sub) On(ctx context.Context, event events.Event) error {
	app.Log.Info().Msgf("Message Received %s", event.String())

	return nil
}
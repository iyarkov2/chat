package main

import (
	"context"
	"encoding/binary"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/getbread/gokit/messaging"
	"github.com/getbread/gokit/requestid"
	"github.com/google/uuid"
	app "github.com/iyarkov2/chat/kafka"
	"github.com/linkedin/goavro/v2"
	"time"
)

var publisher messaging.Publisher

func main() {

	app.WithGokit("gokit-producer")

	app.Must("Register Producer", func() error {
		var err error
		publisher, err = app.Broker.RegisterPublisher(app.Topic)
		return err
	})

	app.WithTimer("Start Broker", func() error {
		return app.Broker.Start()
	})

	app.ExperimentOnce(publish)
	//experimentParallel()
}

func experimentParallel() {
	goroutines := 1000
	batchSize := 10

	app.Log.Info().Msg("---------------------------------------")
	app.Log.Info().Msgf("Gorutines: %d Batch size: %d", goroutines, batchSize)
	ch := make(chan int, goroutines)
	app.WithTimer("Batch", func() error {
		for t := 0; t < goroutines; t++ {
			idx := t
			go func() {
				for i := 0; i < batchSize; i++ {
					publish()
				}
				ch <- idx
			}()
		}
		counter := 0
		for range ch {
			// app.Log.Debug().Msgf("2. Publisher %d done", i)
			counter++
			if counter == goroutines {
				close(ch)
			}
		}
		return nil
	})

	app.Log.Info().Msg("Done")
}

func encode(ctx context.Context, schemaID uint32, in interface{}, codec *goavro.Codec) ([]byte, error) {
	prefix := make([]byte, 5)
	prefix[0] = 0 // magic byte
	binary.BigEndian.PutUint32(prefix[1:], schemaID) // Schema ID
	body, err := codec.BinaryFromNative(nil, in)
	if err != nil {
	return nil, err
	}
	result := make([]byte, len(prefix) + len(body))
	_ = copy(result, prefix)
	_ = copy(result[5:], body)
	app.Log.Info().Msg("Encoding done")
	return result, nil
}

func publish() {
	var codec *goavro.Codec
	app.Must("New Codec", func() error {
		var err error
		codec, err = goavro.NewCodec(`
            {
              "type": "record",
              "name": "FooBar",
              "fields" : [
                {"name": "name", "type": "string" }
              ]
            }`,
		)
		return err
	})

	const SchemaID = 100002
	//
	//datacodec.AddEncoder("bread/avro", func(ctx context.Context, in interface{}) ([]byte, error) {
	//	return encode(ctx, SchemaID, in, codec)
	//})
	//
	id := uuid.New()
	app.WithTimer("Publish " + id.String(), func() error {
		ctx, err := requestid.SetOnContext(context.Background(), id)
		if err != nil {
			return err
		}
		e := event.New()
		e.SetID(id.String())
		e.SetType("event")
		//e.SetTime(time.Now().UTC())
		e.SetSource("test app")


		e.SetExtension("MessageID", id.String())
		e.SetExtension("CreatedAt", time.Now().String())
		e.SetExtension("PublishedAt", time.Now().String())
		e.SetExtension("PublisherID", "test-node-1")
		e.SetExtension("SchemaID", SchemaID)

		e.SetExtension("SchemaVersion", 2)
		e.SetExtension("Entity", "FooBar")
		e.SetExtension("EntityID", "0000-1111-2222")
		e.SetExtension("Action", "Pull")


		value := map[string]interface{} {
			"name": "Non-ASCII text - Привет, encoded by Event",
		}

		data, err := encode(ctx, SchemaID, value, codec)
		if err != nil {
			return err
		}
		err = e.SetData("bread/avro", data)

		return publisher.Publish(ctx, e)
	})
}


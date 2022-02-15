// Package config loads and holds all relevant service configurations
package config

import (
	"fmt"
	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
	goKitCfg "github.com/getbread/gokit/config"
	"github.com/getbread/gokit/messaging/kafka"
	"github.com/getbread/gokit/tools/backoff"
	"github.com/rs/zerolog"
	"os"
	"time"
)

const (
	Topic = "producer-latency-test"
)


type config struct {
	goKitCfg.Kafka

	confluent struct {


	}
}

var (
	Log zerolog.Logger
	Broker *kafka.Broker
	ConfluentConfig *confluent.ConfigMap

	cfg config
)

// Get returns the config loaded from yaml configuration files
func init() {
	// My local configuration file. Change to where your local config is
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic( err )
	}

	err = os.Setenv("CONFIG_FILE_PATH", homeDir + "/Projects/chat/local_config")
	if err != nil {
		panic( err )
	}

	err = goKitCfg.GetCustomForDefaultPath("yaml", &cfg)
	if err != nil {
		panic(fmt.Sprintf("could not parse the service configuration: %s", err))
	}

	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	Log = zerolog.New(zerolog.ConsoleWriter {Out: os.Stdout, TimeFormat: "2006-01-02T15:04:0543"}).With().Timestamp().Logger()

	ConfluentConfig = &confluent.ConfigMap {
		"bootstrap.servers": cfg.Kafka.Broker,
		"security.protocol": "SASL_SSL",
		"sasl.mechanisms": "PLAIN",
		"sasl.username": cfg.Key.Unmask(),
		"sasl.password": cfg.Secret.Unmask(),
		"group.id": "Confluent-consumer",
		"auto.offset.reset": "earliest",
	}

	Log.Info().Msg("App Init completed")
}

func WithGokit(appId string) {
	options := []kafka.ConfigOption{
		kafka.WithRemoteSASLConnection(cfg.Key.Unmask(), cfg.Secret.Unmask()),
		kafka.RetryStrategy(3, backoff.DefaultExponential),
	}

	Broker = kafka.NewConnector (
		appId,
		[]string{cfg.Kafka.Broker},
		Log,
		cfg.Kafka.TopicPrefix,
		options...,
	)
}

func WithTimer(name string, action func() error) {
	start := time.Now()
	err := action()
	t := time.Now()
	elapsed := t.Sub(start)
	Log.Info().Msgf("%s: %d ms", name, elapsed.Milliseconds())

	if err != nil {
		Log.Panic().Err(err).Msgf("Operation %s failed", name)
	}
}

func Must(name string, action func() error) {
	err := action()
	if err != nil {
		Log.Panic().Err(err).Msgf("Operation %s failed", name)
	}
}

func ExperimentSerial(publish func()) {
	sleep := time.Duration(10)
	Log.Info().Msg("---------------------------------------")
	Log.Info().Msgf("single message,  %d min pause", sleep)

	for i:= 0; i < 10; i++ {
		publish()
	}

	Log.Info().Msgf("Producer paused %d", sleep)
	time.Sleep(sleep * time.Minute)

	for i:= 0; i < 10; i++ {
		publish()
	}

}

func ExperimentOnce(publish func()) {
	Log.Info().Msg("---------------------------------------")
	publish()

	Log.Info().Msg("-----------Done------------------------")

}
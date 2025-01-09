package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alecthomas/kong"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/lmittmann/tint"
)

var CLI struct {
	Email struct {
		MqttUsername string `help:"MQTT username." arg:""`
		MqttPassword string `help:"MQTT password." arg:""`
		MqttHostname string `help:"MQTT hostname." arg:""`
		MqttPort     int    `help:"MQTT port." default:"1883"`
		Topic        string `help:"MQTT topic to subscribe to." arg:""`
		From         string `help:"Email from." arg:""`
		To           string `help:"Email to." arg:""`
		Subject      string `help:"Email subject." arg:""`
	} `cmd:"" help:"Send email."`

	PlaySound struct {
		MqttUsername string `help:"MQTT username." arg:""`
		MqttPassword string `help:"MQTT password." arg:""`
		MqttHostname string `help:"MQTT hostname." arg:""`
		MqttPort     int    `help:"MQTT port." default:"1883"`
		Topic        string `help:"MQTT topic to subscribe to." arg:""`
	} `cmd:"" help:"Play Sound."`
}

type Payload struct {
	Encryption    bool   `json:"encryption"`
	BTHomeVersion int    `json:"BTHome_version"`
	Pid           int    `json:"pid"`
	Battery       int    `json:"Battery"`
	Illuminance   int    `json:"Illuminance"`
	Motion        int    `json:"Motion"`
	Addr          string `json:"addr"`
	Rssi          int    `json:"rssi"`
}

func SendEmail(msg MQTT.Message, from string, to string, subject string) {
	// TODO
	var payload Payload
	err := json.Unmarshal(msg.Payload(), &payload)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		return
	}

	// Use the parsed data
	slog.Debug("Parsed payload", "payload", payload)
}

// var soundHandler MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
// 	// TODO
// }

func DoEmail(ctx *kong.Context, hostname string) {
	opts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", CLI.Email.MqttHostname, CLI.Email.MqttPort))
	opts.SetClientID("rook_" + hostname)

	emailHandler := func(client MQTT.Client, msg MQTT.Message) {
		SendEmail(msg, CLI.Email.From, CLI.Email.To, CLI.Email.Subject)
	}

	opts.SetDefaultPublishHandler(emailHandler)
	opts.SetUsername(CLI.Email.MqttUsername)
	opts.SetPassword(CLI.Email.MqttPassword)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("Mqtt", "connect", token.Error())
		os.Exit(1)
	}

	token := client.Subscribe(CLI.Email.Topic, 1, nil)
	token.Wait()
	slog.Info("Subscribed", "topic", CLI.Email.Topic)
}

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Hostname", "error", err)
		os.Exit(1)
	}

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "email":
		DoEmail(ctx, hostname)
	case "playsound":
	default:
		panic(ctx.Command())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// TODO Add a little REST API to get status
	slog.Info("Waiting for messages.")
	<-c
	slog.Info("Exiting gracefully.")
}

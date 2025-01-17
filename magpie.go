package main

import (
	"899bushwick/magpie/email"
	"899bushwick/magpie/schema"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
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

	GmailUsernameFile string `help:"Gmail username." default:"gmail_username.txt"`
	GmailPasswordFile string `help:"Gmail password. Access" default:"gmail_password.txt"`
}

func ExtractEmail(msg MQTT.Message,
	from string,
	to string,
	subject string,
	gmail_username string,
	gmail_password string) {
	var update schema.NodeUpdate
	err := json.Unmarshal(msg.Payload(), &update)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		return
	}

	// Use the parsed data
	slog.Debug("Parsed payload", "update", update)
	email_body := fmt.Sprintf("Node: %s\nTimestamp: %s\nDescription: %s\n", update.Node, update.Timestamp, update.Description)
	email.Send(to, subject, email_body, gmail_username, gmail_password)
}

func DoEmail(ctx *kong.Context, hostname string) {
	opts := MQTT.NewClientOptions().AddBroker(fmt.Sprintf("tcp://%s:%d", CLI.Email.MqttHostname, CLI.Email.MqttPort))
	opts.SetClientID("rook_" + hostname)

	dat, err := os.ReadFile(CLI.GmailUsernameFile)
	if err != nil {
		panic(err)
	}
	gmail_username := strings.TrimSpace(string(dat))

	dat2, err2 := os.ReadFile(CLI.GmailPasswordFile)
	if err2 != nil {
		panic(err2)
	}
	gmail_password := strings.TrimSpace(string(dat2))

	emailHandler := func(client MQTT.Client, msg MQTT.Message) {
		go ExtractEmail(msg, CLI.Email.From, CLI.Email.To, CLI.Email.Subject, gmail_username, gmail_password)
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

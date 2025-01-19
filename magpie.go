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

var (
	motion_map    = make(map[string]int)
	openclose_map = make(map[string]int)
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
	} `cmd:"" help:"Send email."`

	PlaySound struct {
		MqttUsername string `help:"MQTT username." arg:""`
		MqttPassword string `help:"MQTT password." arg:""`
		MqttHostname string `help:"MQTT hostname." arg:""`
		MqttPort     int    `help:"MQTT port." default:"1883"`
		Topic        string `help:"MQTT topic to subscribe to." arg:""`
	} `cmd:"" help:"Play Sound."`

	Shelly struct {
		MqttUsername string `help:"MQTT username." arg:""`
		MqttPassword string `help:"MQTT password." arg:""`
		MqttHostname string `help:"MQTT hostname." arg:""`
		MqttPort     int    `help:"MQTT port." default:"1883"`
		SourceTopic  string `help:"MQTT topic to subscribe to." arg:""`
		DestTopic    string `help:"MQTT topic to subscribe to." arg:""`
		Type         string `enum:"motion,window" help:"Type of Shelly device." arg:""`
	} `cmd:"" help:"Play Sound."`

	GmailUsernameFile string `help:"Gmail username." default:"gmail_username.txt"`
	GmailPasswordFile string `help:"Gmail password. Access" default:"gmail_password.txt"`
}

func OpenCloseEmail(msg MQTT.Message,
	from string,
	to string,
	gmail_username string,
	gmail_password string) {

	var update schema.ShellyOpenClose
	err := json.Unmarshal(msg.Payload(), &update)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		return
	}

	// Use the parsed data
	slog.Debug("Parsed payload", "topic", msg.Topic())
	slog.Debug("Parsed payload", "update", update)

	send_email := true

	email_body := fmt.Sprintf("Magpie. Topic: %s\nWindow: %d\n", msg.Topic(), update.Window)

	// check previoius value
	val, ok := openclose_map[msg.Topic()]
	if ok {
		send_email = val == update.Window
	}

	openclose_map[msg.Topic()] = update.Window

	if send_email {
		var subject string
		if update.Window == 1 {
			subject = msg.Topic() + ": Open"
		} else {
			subject = msg.Topic() + ": Closed"
		}
		email.Send(to, subject, email_body, gmail_username, gmail_password)
	}
}

func DoEmail(ctx *kong.Context, hostname string) MQTT.Client {
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
		if strings.HasPrefix(msg.Topic(), "mostert/openclose/") {
			go OpenCloseEmail(msg, CLI.Email.From, CLI.Email.To, gmail_username, gmail_password)
		}
	}

	opts.SetDefaultPublishHandler(emailHandler)
	opts.SetUsername(CLI.Email.MqttUsername)
	opts.SetPassword(CLI.Email.MqttPassword)

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		slog.Error("Mqtt", "connect", CLI.Email.MqttUsername)
		slog.Error("Mqtt", "connect", CLI.Email.MqttPassword)
		slog.Error("Mqtt", "connect", token.Error())
		os.Exit(1)
	}

	token := client.Subscribe(CLI.Email.Topic, 1, nil)
	token.Wait()
	slog.Info("Subscribed", "topic", CLI.Email.Topic)

	return client
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

	// rcache.SubscribeRedis()

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "email <mqtt-username> <mqtt-password> <mqtt-hostname> <topic> <from> <to>":
		DoEmail(ctx, hostname)
	case "shelly":
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

// func DoMotion(msg MQTT.Message) {
// 	// Parse the JSON payload
// 	var payload Payload
// 	err := json.Unmarshal(msg.Payload(), &payload)
// 	if err != nil {
// 		fmt.Printf("Error parsing JSON: %s\n", err)
// 		return
// 	}

// 	// Use the parsed data
// 	slog.Debug("Parsed payload", "payload", payload)

// 	val, ok := motion_map[msg.Topic()]
// 	if (ok && val != payload.Motion) || !ok {
// 		body := "Motion detected"
// 		if payload.Motion == 0 {
// 			body = "Motion cleared"
// 		}
// 		err = sendEmail(gmail_username, msg.Topic(), body)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}

// 	motion_map[msg.Topic()] = payload.Motion
// }

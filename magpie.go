package main

import (
	"899bushwick/magpie/email"
	"899bushwick/magpie/lms"
	"899bushwick/magpie/rcache"
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

	GmailUsernameFile string `help:"Gmail username." default:"gmail_username.txt"`
	GmailPasswordFile string `help:"Gmail password. Access" default:"gmail_password.txt"`
	RedisUrl          string `help:"Redis URL." default:"redis:6379"`
}

func GarageHandler(msg MQTT.Message) {

	if strings.HasSuffix(msg.Topic(), "/rpc") {
		slog.Debug("GarageHandler", "topic", msg.Topic(), "payload", string(msg.Payload()))
		var update schema.ShellyRpcInput
		err := json.Unmarshal(msg.Payload(), &update)
		if err != nil {
			fmt.Printf("Error parsing JSON: %s\n", err)
			return
		}
		slog.Debug("Parsed payload", "update", update)

		// input state is the reed swtich (disconnected)
		var garageKey string = strings.Replace(msg.Topic(), "/events/rpc", "", 1)
		rcache.SetGarageState(garageKey, update.Params.Input.State)
	} else if strings.HasSuffix(msg.Topic(), "/online") {
		// ignore
		slog.Debug("GarageHandler", "topic", msg.Topic(), "payload", string(msg.Payload()))
	}
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

	rcache.RedisUrl = CLI.RedisUrl
	send_email := true // change detected
	play_sound := false
	sound_enabled, err := rcache.CheckEnabled()
	if err != nil {
		slog.Error("check_enabled", "error", err)
	} else {
		play_sound = sound_enabled.Enabled
	}
	slog.Debug("check_enabled", "sound_enabled", sound_enabled)

	email_body := fmt.Sprintf("Magpie. Topic: %s\nWindow: %d\n", msg.Topic(), update.Window)

	// check previoius value
	val, ok := openclose_map[msg.Topic()]
	if ok {
		send_email = val != update.Window
	}

	openclose_map[msg.Topic()] = update.Window

	if send_email {
		var subject string
		if update.Window == 1 {
			subject = msg.Topic() + ": Open"
		} else {
			subject = msg.Topic() + ": Closed"
		}
		go email.Send(to, subject, email_body, gmail_username, gmail_password)
	}

	rcache.SetOpenClose(msg.Topic(), update.Window)

	if play_sound && send_email {
		if update.Window == 1 {
			go lms.PlaySound(fmt.Sprintf("/music/%s_open.mp3", msg.Topic()))
		} else {
			go lms.PlaySound(fmt.Sprintf("/music/%s_closed.mp3", msg.Topic()))
		}
	}
}

func MotionEmail(msg MQTT.Message,
	from string,
	to string,
	gmail_username string,
	gmail_password string) {

	var update schema.ShellyMotion
	err := json.Unmarshal(msg.Payload(), &update)
	if err != nil {
		fmt.Printf("Error parsing JSON: %s\n", err)
		return
	}

	// Use the parsed data
	var topic string = msg.Topic()
	var motion int = update.Motion
	slog.Debug("Parsed payload", "topic", topic)
	slog.Debug("Parsed payload", "update", update)

	// defalt to send if no previous value
	send_email := true

	email_body := fmt.Sprintf("Magpie. Topic: %s\nMotion: %d\n", topic, motion)

	// check previoius value
	last_val, ok := motion_map[topic]
	if ok {
		send_email = last_val != motion

		email_body += fmt.Sprintf("Last value: %d\n", last_val)
	}

	motion_map[topic] = motion

	if send_email {
		var subject string
		if motion == 1 {
			subject = topic + ": Motion Detected"
		} else {
			subject = topic + ": Motion Cleared"
		}
		go email.Send(to, subject, email_body, gmail_username, gmail_password)
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
			OpenCloseEmail(msg, CLI.Email.From, CLI.Email.To, gmail_username, gmail_password)
		} else if strings.HasPrefix(msg.Topic(), "mostert/motion/") {
			MotionEmail(msg, CLI.Email.From, CLI.Email.To, gmail_username, gmail_password)
		} else if strings.HasPrefix(msg.Topic(), "mostert/garage/") {
			GarageHandler(msg)
		} else {
			slog.Warn("Unhandled topic", "topic", msg.Topic())
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

	go rcache.SubscribeRedis()

	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "email <mqtt-username> <mqtt-password> <mqtt-hostname> <topic> <from> <to>":
		DoEmail(ctx, hostname)
	default:
		panic(ctx.Command())
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	slog.Info("Waiting for messages.")
	<-c
	slog.Info("Exiting gracefully.")
}

package lms

import (
	"fmt"
	"log/slog"
	"net"
)

func PlaySound(mp3 string) {
	playerId := "d8:bc:38:e6:e8:40"
	// let volume_command = "d8:bc:38:e6:e8:40 mixer volume 90\n";

	play_command := fmt.Sprintf("%s playlist play %s fd 1\n", playerId, mp3)
	servAddr := "192.168.0.197:9090"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		slog.Error("play_sound", "ResolveTCPAddr", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		slog.Error("play_sound", "Dial", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(play_command))
	if err != nil {
		slog.Error("play_sound", "Write", err)
		return
	}

	slog.Debug("play_sound", "command", play_command)

	reply := make([]byte, 1024)

	_, err = conn.Read(reply)
	if err != nil {
		slog.Error("play_sound", "Read", err)
		return
	}
	// slog.Debug("play_sound", "reply", string(reply))
}

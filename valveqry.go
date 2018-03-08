// Package valveqry for querying a valve game server
package valveqry

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

type ServerInf struct {
	Protocol    byte
	Name        string
	Map         string
	Folder      string
	Game        string
	Id          uint16
	Players     byte
	MaxPlayers  byte
	Bots        byte
	Type        byte
	Environment byte
	Visibility  byte
	Vac         byte
	Version     string
	Edf         byte
	Port        uint16
	SteamId     uint32
	SpecPort    uint16
	TvName      string
	Keywords    string
	GameId      uint32
}

var lastCall time.Time
var recentReqs = 0
var sinfo ServerInf
var packet = make([]byte, 1024)

func fillInfo() ServerInf {
	info := bytes.SplitN(packet[6:], []byte{0}, 5)
	sinfo.Protocol = packet[5]
	sinfo.Name = string(info[0])
	sinfo.Map = string(info[1])
	sinfo.Folder = string(info[2])
	sinfo.Game = string(info[3])
	sinfo.Id = binary.BigEndian.Uint16(info[4][0:2])
	sinfo.Players = info[4][2]
	sinfo.MaxPlayers = info[4][3]
	sinfo.Bots = info[4][4]
	sinfo.Type = info[4][5]
	sinfo.Environment = info[4][6]
	sinfo.Visibility = info[4][7]
	sinfo.Vac = info[4][8]
	sinfo.Version = string(bytes.Split(info[4][9:], []byte{0})[0])

	fmt.Println(sinfo.Version)

	return sinfo
}

func GetInfo(server string) (*ServerInf, error) {

	secs := time.Now().Sub(lastCall).Seconds()
	if secs < 60 {
		if secs < 5 || recentReqs == 3 {
			return &sinfo, errors.New("Maximum requests 3 per minute 1 per 5 seconds.")
		}
	} else {
		recentReqs = 0
	}
	recentReqs += 1
	lastCall = time.Now()

	ServerAddr, err := net.ResolveUDPAddr("udp4", server)
	if err != nil {
		return &sinfo, errors.New("Couldn't resolve server.")
	}
	Conn, err := net.DialUDP("udp", nil, ServerAddr)
	if err != nil {
		return &sinfo, errors.New("Server not responding.")
	}
	defer Conn.Close()

	msg := []byte("\xFF\xFF\xFF\xFF\x54Source Engine Query\x00")
	_, err = Conn.Write(msg)
	if err != nil {
		return &sinfo, errors.New("Error sending query.")
	}
	t := time.Now()
	Conn.SetDeadline(t.Add(10 * time.Second))
	_, _, err = Conn.ReadFromUDP(packet)
	if err != nil {
		return &sinfo, errors.New("No reply from server.")
	}
	fillInfo()
	return &sinfo, nil
}

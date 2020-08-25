package core

import (
	"net"
)

type ClientHandler struct {
	IP       string
	Port     string
	Listener net.Listener
}

func BeginClientHandling(ip string, port string) (*ClientHandler, error) {
	l, err := net.Listen("tcp", ip+":"+port)

	if err != nil {
		return nil, err
	}

	ch := new(ClientHandler)
	ch.IP = ip
	ch.Port = port
	ch.Listener = l

	return ch, nil
}

package main

import (
	"bufio"
	"log"
	"midnight/pkg/core"
	"midnight/pkg/logging"
	"net"
	"strconv"
)

func main() {
	conf, err := core.LoadConfig_File()
	if err != nil {
		log.Println("Could not load server.json. Loading defaults and creating file.")
		core.SaveConfigToFile(conf)
	}

	ch, err := core.BeginClientHandling(conf.IP, strconv.FormatInt(int64(conf.Port), 10))
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
		return
	}

	s := core.StartServer(ch, conf)
	log.Println("Started server. Accepting clients.")

	for {
		conn, err := ch.Listener.Accept()
		if err != nil {
			log.Println("Could not accept client:")
			log.Println(err)
			continue
		}

		go newConnection(conn, s)
	}
}

func newConnection(conn net.Conn, server *core.Server) {
	log.Println("Connected [" + conn.RemoteAddr().String() + "]")

	c := core.Client{
		Conn:   conn,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
	}

	// Read Player Identification (0x00)
	packet, protocol, username, verify, ext, err := c.ReadPacket_PlayerIdentification()

	if err != nil {
		log.Println("Error while reading from client [" + conn.RemoteAddr().String() + "]")
		log.Println(err)
		c.Conn.Close()
		return
	}

	// Validate Player Identification
	if packet != 0x00 {
		log.Println("[" + conn.RemoteAddr().String() + "] Invalid Player Identification Packet ID. Disconnecting client.")
		c.Conn.Close()
		return
	}
	if protocol != 0x07 {
		log.Println("[" + conn.RemoteAddr().String() + "] Invalid Player Identification Protocol. Disconnecting client.")
		c.Conn.Close()
		return
	}
	if ext != 0x42 {
		log.Println("[" + conn.RemoteAddr().String() + "] Invalid Player Identification Padding Byte. Disconnecting client.")
		c.Conn.Close()
		return
	}

	if verify != "" {
		// TODO: Validate this
	}

	// Write ExtInfo/ExtEntry
	c.WritePacket_ExtInfo("Midnight", 2)
	c.WritePacket_ExtEntry("EmoteFix", 1)
	c.WritePacket_ExtEntry("LongerMessages", 1)

	// Read ExtInfo/ExtEntry
	packet, appName, extCount, err := c.ReadPacket_ExtInfo()

	if err != nil {
		log.Println("Error while reading from client [" + conn.RemoteAddr().String() + "]")
		log.Println(err)
		c.Conn.Close()
		return
	}

	if packet != 0x10 {
		log.Println("[" + conn.RemoteAddr().String() + "] Invalid ExtInfo Packet ID. Disconnecting client.")
		c.Conn.Close()
		return
	}

	logging.Log_Debugf("[%v] Client supports %v protocol extensions:", conn.RemoteAddr().String(), extCount)
	for i := int16(0); i < extCount; i++ {
		// TODO: Note the supported client extensions and internally enable those extensions for the client
		// Config will be able to be modified by admin to reject/kick users if they don't support the required extensions

		extName, version, err := c.ReadPacket_ExtEntry()
		if err != nil {
			log.Println("Error while reading from client [" + conn.RemoteAddr().String() + "]")
			log.Println(err)
			c.Conn.Close()
			return
		}
		logging.Log_Debugf("[Ext %v] '%v' v%v", i+1, extName, version)
	}

	// Send Handshake
	c.WritePacket_ServerIdentification("Midnight Station", "This is Fullerton. This is a Red Line train to 95th.", true)

	// Create player & join user to server instance
	p := core.Player{
		Cli:             c,
		Username:        username,
		IP:              c.Conn.RemoteAddr().String(),
		Client_Software: appName,
	}

	server.JoinUser(p)
}

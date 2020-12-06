package core

import (
	"crypto/rand"
	"io"
	"log"
	"midnight/pkg/util"
	"net/url"
	"strconv"
)

type Server struct {
	name        string
	port        string
	Salt        string // Salt exported for use in main.go
	public      bool
	maxUsers    int32
	VerifyLogin bool // VerifyLogin exported for use in main.go

	lvl     *Level
	players []Player

	ch *ClientHandler
}

func StartServer(ch *ClientHandler, conf *Config) *Server {
	s := new(Server)

	s.name = conf.ServerName
	s.port = strconv.FormatInt(int64(conf.Port), 10)
	s.public = conf.Public
	s.maxUsers = int32(conf.MaxUsers)
	s.VerifyLogin = conf.VerifyLogin

	if conf.Debug.OverrideSalt == true {
		s.Salt = conf.Debug.Salt
	} else {
		s.Salt = GenerateSalt()
	}

	s.lvl = ConstructLevel("main", 256, 256, 256)
	s.lvl.GenerateFlat()

	// TODO: Start heartbeat
	if s.public {
		go BeginHeartbeatLoop(s)
	}

	return s
}

func (s *Server) JoinUser(p Player) {
	// Add to player list
	// Send level

	s.players = append(s.players, p)
	s.lvl.Players = append(s.lvl.Players, p)

	p.Cli.WritePacketUtil_SendLevel(s.lvl)
	p.Cli.WritePacket_SpawnPlayer(s.lvl.SpawnPos, 0, 0, -1, p.Username)
	//p.Cli.WritePacket_PlayerTeleport(s.lvl.SpawnPos, 0, 0, -1)

	for {
		packet, err := p.Cli.ReadPacketEntry()

		if err != nil {
			s.disconnectPlayer(p, "")
			return
		}

		switch packet {
		case 0x05:
			x, y, z, mode, blockType, err := p.Cli.ReadPacket_SetBlock()

			if err != nil {
				s.disconnectPlayer(p, "")
				return
			}

			if mode == 0x00 { // Destroy
				s.lvl.ChangeBlock(0, util.Vector3i16{X: x, Y: y, Z: z})
			} else { // mode == 0x01; Create
				s.lvl.ChangeBlock(blockType, util.Vector3i16{X: x, Y: y, Z: z})
			}

		case 0x08:
			playerId, x, y, z, yaw, pitch, err := p.Cli.ReadPacket_PositionUpdate()

			if err != nil {
				s.disconnectPlayer(p, "")
				return
			}

			// TODO
			_ = playerId
			_ = x
			_ = y
			_ = z
			_ = yaw
			_ = pitch

		case 0x0D:
			longMessage, message, err := p.Cli.ReadPacket_Message()

			if err != nil {
				s.disconnectPlayer(p, "")
				return
			}

			// TODO
			_ = longMessage
			_ = message
		default:
			log.Printf("Test: %v", packet)
		}
	}
}

// Disconnects a player and reduces the number of players in the levels and the server.
// Leave disconnectMsg empty is no 0x0e packet is being sent.
func (s *Server) disconnectPlayer(p Player, disconnectMsg string) {
	for i := 0; i < len(s.players); i++ {
		if s.players[i] == p {
			s.players = append(s.players[:i], s.players[i+1:]...)
		}
	}

	for i := 0; i < len(s.lvl.Players); i++ {
		if s.lvl.Players[i] == p {
			s.lvl.Players = append(s.lvl.Players[:i], s.lvl.Players[i+1:]...)
		}
	}

	if disconnectMsg != "" {
		p.Cli.WritePacket_DisconnectPlayer(disconnectMsg)
	}

	log.Printf("Disconnected [%v]:[%v]", p.Username, p.IP)
}

// Generates a 256-byte salt
// Kinda broken right now? There's an issue out to improve it anyway so leave it as-is until this can be much better improved
func GenerateSalt() string {
	salt := make([]byte, 100)
	_, _ = io.ReadFull(rand.Reader, salt)
	return url.QueryEscape(string(salt))
}

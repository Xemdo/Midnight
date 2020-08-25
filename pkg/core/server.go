package core

import (
	"log"
	"midnight/pkg/util"
	"strconv"
)

type Server struct {
	name        string
	port        string
	salt        string
	public      bool
	maxUsers    int32
	verifyLogin bool

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

	if conf.Debug.OverrideSalt == true {
		s.salt = conf.Debug.Salt
	} else {
		// TODO: Generate salt randomly
	}

	s.lvl = ConstructLevel("main", 256, 256, 256)
	s.lvl.GenerateFlat()

	// TODO: Start heartbeat

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
			s.disconnectPlayer(p, false)
			break
		}

		switch packet {
		case 0x05:
			x, y, z, mode, blockType, err := p.Cli.ReadPacket_SetBlock()

			if err != nil {
				s.disconnectPlayer(p, false)
			}

			if mode == 0x00 { // Destroy
				s.lvl.ChangeBlock(0, util.Vector3i16{X: x, Y: y, Z: z})
			} else { // mode == 0x01; Create
				s.lvl.ChangeBlock(blockType, util.Vector3i16{X: x, Y: y, Z: z})
			}

		case 0x08:
			playerId, x, y, z, yaw, pitch, err := p.Cli.ReadPacket_PositionUpdate()

			if err != nil {
				s.disconnectPlayer(p, false)
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
				s.disconnectPlayer(p, false)
			}

			// TODO
			_ = longMessage
			_ = message
		default:
			log.Printf("Test: %v", packet)
		}
	}
}

func (s *Server) disconnectPlayer(p Player, sendPacket bool) {
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

	if sendPacket {
		// TODO: Send 0x0e
	}

	log.Printf("Disconnected [%v]:[%v]", p.Username, p.IP)
}

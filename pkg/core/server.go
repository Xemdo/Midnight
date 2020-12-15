package core

import (
	"log"
	"math/rand"
	"midnight/pkg/util"
	"regexp"
	"strconv"
	"time"
)

var colorCodeRegex *regexp.Regexp = regexp.MustCompile(`%([0-9a-fA-F])`)

type Server struct {
	name        string
	port        string
	Salt        string // Salt exported for use in main.go
	public      bool
	maxUsers    int32
	VerifyLogin bool // VerifyLogin exported for use in main.go

	lvl     *Level
	players map[int8]Player

	ch  *ClientHandler
	sch *TaskScheduler
}

func StartServer(ch *ClientHandler, conf *Config) *Server {
	s := new(Server)

	s.name = conf.ServerName
	s.port = strconv.FormatInt(int64(conf.Port), 10)
	s.public = conf.Public
	s.maxUsers = int32(conf.MaxUsers)
	s.VerifyLogin = conf.VerifyLogin
	s.players = make(map[int8]Player)

	if conf.Debug.OverrideSalt == true {
		s.Salt = conf.Debug.Salt
	} else {
		s.Salt = GenerateSalt()
	}

	s.lvl = ConstructLevel("main", 256, 256, 256)
	s.lvl.GenerateFlat()

	if s.public {
		go BeginHeartbeatLoop(s)
	}

	log.Printf("Starting task scheduler/loop")
	s.sch = new(TaskScheduler)
	go s.sch.StartServerTickLoop()

	s.createBasicTasks(conf.AnnouncePlayers)

	return s
}

func (s *Server) JoinUser(p Player) {
	// Find open player ID
	// TODO: Make this per-level instead of per-server. Right now it imposes a limit of 127 people in the server

	var playerId int8
	for i := int8(1); i < 127; i++ {
		if _, found := s.players[i]; !found {
			playerId = i
		}
	}

	p.PlayerId = playerId

	s.players[playerId] = p
	s.lvl.Players[playerId] = p

	log.Printf("%v has joined the server [%v]", p.Username, p.IP)

	// Send level
	p.Cli.WritePacketUtil_SendLevel(s.lvl)
	p.Cli.WritePacket_SpawnPlayer(s.lvl.SpawnPos, 0, 0, -1, p.Username)

	// Player packet recieve loop
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

			// Update position to all players in level
			for p := range s.players {
				_ = p
			}

		case 0x0D:
			longMessage, message, err := p.Cli.ReadPacket_Message()

			if err != nil {
				s.disconnectPlayer(p, "")
				return
			}

			// TODO: Enable Long Message
			_ = longMessage

			s.handleIncomingMessage(p, message)

		default:
			log.Printf("Test: %v", packet)
		}
	}
}

// Disconnects a player and reduces the number of players in the levels and the server.
// Leave disconnectMsg empty is no 0x0e packet is being sent.
func (s *Server) disconnectPlayer(p Player, disconnectMsg string) {
	delete(s.players, p.PlayerId)     // Remove player from server player list
	delete(s.lvl.Players, p.PlayerId) // Remove player from level player list

	if disconnectMsg != "" {
		p.Cli.WritePacket_DisconnectPlayer(disconnectMsg)
	}

	log.Printf("Disconnected [%v]:[%v]", p.Username, p.IP)
}

func (s *Server) handleIncomingMessage(sender Player, msg string) {
	formatted := "&e" + sender.Username + ": &f" + msg

	for _, p := range s.players {
		s.SendMessage(p, formatted)
	}

	log.Printf("[Chat] %v: %v", sender.Username, msg)
}

func (s *Server) SendMessage(p Player, msg string) {
	// TODO: Remove 64-char limit
	if len(msg) > 64 {
		msg = msg[0:64]
	}

	// Change color codes from % to &. E.g. %e becomes &e
	msg = colorCodeRegex.ReplaceAllString(msg, "&${1}")

	p.Cli.WritePacket_Message(-1, msg)
}

func (s *Server) SendAnnouncement(msg string) {
	for _, p := range s.players {
		s.SendMessage(p, "&e[Server] "+msg)
	}
}

func (s *Server) createBasicTasks(plTaskEnabled bool) {
	if plTaskEnabled {
		// Create player announcement task
		plTask := Task{
			Id:           "players-announce",
			ExecDelay:    300000, // 5 minutes
			DelayedStart: true,
			TaskFunc: func() {
				playerList := ""
				for _, player := range s.players {
					if playerList != "" {
						playerList += ", "
					}
					playerList += player.Username + "[" + strconv.Itoa(int(player.PlayerId)) + "]"
				}

				// TODO: Print as server message to all players
				log.Printf("Players Online: [ %v ]", playerList)
			},
		}

		s.sch.AddTask(plTask)
	}
}

var saltRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+-=[]{}\\|;':\",./<>?`~")

// Generates a 256-byte salt
func GenerateSalt() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, 256)
	for i := range b {
		b[i] = saltRunes[r.Intn(len(saltRunes))]
	}
	return string(b)
}

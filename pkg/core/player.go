package core

type Player struct {
	Cli             Client
	Username        string
	IP              string
	Client_Software string
	PlayerId        int8

	PosX, PosY, PosZ float32
	Pitch, Yaw       byte
}

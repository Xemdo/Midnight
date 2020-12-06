package core

import (
	"bufio"
	"math"
	"midnight/pkg/logging"
	"midnight/pkg/util"
	"net"
	"strings"
)

type Client struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

// Data-type read functions

func (c Client) ReadByte() (result byte, err error) {
	single, err := c.Reader.ReadByte()
	if err != nil {
		return 0xAF, err
	}

	return single, nil
}

func (c Client) ReadSByte() (result int8, err error) {
	single, err := c.Reader.ReadByte()
	if err != nil {
		return -1, err
	}

	return int8(single), nil
}

func (c Client) ReadShort() (result int16, err error) {
	b0, err := c.Reader.ReadByte()
	b1, err := c.Reader.ReadByte()

	if err != nil {
		return -1, err
	}

	return int16(b0)<<8 | int16(b1), nil
}

func (c Client) ReadInt() (result int32, err error) {
	b0, err := c.Reader.ReadByte()
	b1, err := c.Reader.ReadByte()
	b2, err := c.Reader.ReadByte()
	b3, err := c.Reader.ReadByte()

	if err != nil {
		return -1, err
	}

	return int32(b0)<<24 | int32(b1)<<16 | int32(b2)<<8 | int32(b3), nil
}

func (c Client) ReadString(trim bool) (result string, err error) {
	var raw [64]byte

	for i := 0; i < 64; i++ {
		raw[i], err = c.Reader.ReadByte()

		if err != nil {
			return "", err
		}
	}

	if trim {
		return strings.TrimSpace(string(raw[:])), nil
	} else {
		return string(raw[:]), nil
	}
}

// Packet read functions

func (c Client) ReadPacket_PlayerIdentification() (packet byte, protocol byte, username string, verify string, ext byte, err error) {
	packet, err = c.ReadByte()
	protocol, err = c.ReadByte()
	username, err = c.ReadString(true)
	verify, err = c.ReadString(true)
	ext, err = c.ReadByte()

	logging.Log_Debugf("[%v] [Read] {%v, %v, %v, %v, %v}", c.Conn.RemoteAddr(), packet, protocol, username, verify, ext)

	return packet, protocol, username, verify, ext, err
}

// 0x05 - Set Block
func (c Client) ReadPacket_SetBlock() (x int16, y int16, z int16, mode byte, blockType byte, err error) {
	x, err = c.ReadShort()
	y, err = c.ReadShort()
	z, err = c.ReadShort()
	mode, err = c.ReadByte()
	blockType, err = c.ReadByte()

	logging.Log_Debugf("[%v] [Read] {%v, %v, %v, %v, %v, %v}", c.Conn.RemoteAddr(), 0x05, x, y, z, mode, blockType)
	return x, y, z, mode, blockType, err
}

// 0x08 - Position and Orientation
func (c Client) ReadPacket_PositionUpdate() (playerId int8, x float32, y float32, z float32, yaw byte, pitch byte, err error) {
	playerId, err = c.ReadSByte()
	_x, err := c.ReadShort()
	_y, err := c.ReadShort()
	_z, err := c.ReadShort()
	yaw, err = c.ReadByte()
	pitch, err = c.ReadByte()

	x = float32(_x) / 32
	y = float32(_y) / 32
	z = float32(_z) / 32

	//fmt.Printf("Test1: [%v, %v, %v]\n", float32(int16(x*32))/32, float32(int16(y*32))/32, float32(int16(z*32))/32)
	//fmt.Printf("Test2: [%v, %v, %v]\n", x, y, z)

	//logging.Log_Debugf("[%v] [Read] {%v, %v, %v, %v, %v, %v, %v}", c.Conn.RemoteAddr(), 0x08, playerId, x, y, z, yaw, pitch)
	return playerId, x, y, z, yaw, pitch, err
}

func (c Client) ReadPacket_Message() (longMessage byte, message string, err error) {
	longMessage, err = c.ReadByte() // Unused byte (0xFF)
	message, err = c.ReadString(false)

	return longMessage, message, err
}

func (c Client) ReadPacket_ExtInfo() (packet byte, appName string, extensionCount int16, err error) {
	packet, err = c.ReadByte()
	appName, err = c.ReadString(true)
	extensionCount, err = c.ReadShort()

	logging.Log_Debugf("[%v] [Read] {%v, %v, %v}", c.Conn.RemoteAddr(), packet, appName, extensionCount)

	return packet, appName, extensionCount, err
}

func (c Client) ReadPacket_ExtEntry() (extName string, version int32, err error) {
	packet, err := c.ReadByte()
	extName, err = c.ReadString(true)
	version, err = c.ReadInt()

	logging.Log_Debugf("[%v] [Read] {%v, %v, %v}", c.Conn.RemoteAddr(), packet, extName, version)

	return extName, version, err
}

func (c Client) ReadPacketEntry() (packet byte, err error) {
	packet, err = c.ReadByte()

	return packet, err
}

// Data-type write functions

func (c Client) WriteShort(v int16) {
	var b0, b1 uint8 = uint8(v >> 8), uint8(v & 0xFF)
	c.Writer.WriteByte(b0)
	c.Writer.WriteByte(b1)
}

func (c Client) WriteInt(v int32) {
	var b0, b1, b2, b3 uint8 = uint8(v >> 24), uint8(v >> 16), uint8(v >> 8), uint8(v & 0xFF)
	c.Writer.WriteByte(b0)
	c.Writer.WriteByte(b1)
	c.Writer.WriteByte(b2)
	c.Writer.WriteByte(b3)
}

// Packet write functions

func (c Client) WritePacket_ServerIdentification(server string, motd string, op bool) {
	var userType byte = 0x00 // userType = 0x00 for normal user; 0x64 for OP
	if op {
		userType = 0x64
	}

	c.Writer.WriteByte(0x00)                          // Packet ID
	c.Writer.WriteByte(0x07)                          // Protocol Version
	c.Writer.Write(WritePacketUtil_PadString(server)) // Server name
	c.Writer.Write(WritePacketUtil_PadString(motd))   // Server MOTD
	c.Writer.WriteByte(userType)                      // User Type
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v, %v, %v}", c.Conn.RemoteAddr(), 0x00, 0x07, string(WritePacketUtil_PadString(server)), string(WritePacketUtil_PadString(motd)), userType)
}

func (c Client) WritePacket_LevelInit() {
	c.Writer.WriteByte(0x02)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v}", c.Conn.RemoteAddr(), 0x02)
}

func (c Client) WritePacket_LevelDataChunk(chunkLength int, data []byte, percentComplete byte) {
	c.Writer.WriteByte(0x03)
	c.WriteShort(int16(chunkLength))
	c.Writer.Write(data)
	c.Writer.WriteByte(percentComplete)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, <chunk data|%v>, %v}", c.Conn.RemoteAddr(), 0x03, chunkLength, len(data), percentComplete)
}

func (c Client) WritePacket_LevelFinalize(sizeX int16, sizeY int16, sizeZ int16) {
	c.Writer.WriteByte(0x04)
	c.WriteShort(sizeX)
	c.WriteShort(sizeY)
	c.WriteShort(sizeZ)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v, %v}", c.Conn.RemoteAddr(), 0x04, sizeX, sizeY, sizeZ)
}

func (c Client) WritePacket_SetBlock(block byte, x int16, y int16, z int16) {
	c.Writer.WriteByte(0x06)
	c.WriteShort(x)
	c.WriteShort(y)
	c.WriteShort(z)
	c.Writer.WriteByte(block)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v, %v, %v}", c.Conn.RemoteAddr(), 0x06, x, y, z, block)
}

func (c Client) WritePacket_SpawnPlayer(pos util.Vector3i16, yaw byte, pitch byte, playerId int8, playerName string) {
	c.Writer.WriteByte(0x07)
	c.Writer.WriteByte(byte(playerId))
	c.Writer.Write(WritePacketUtil_PadString(playerName))
	c.WriteShort(pos.X * 32)
	c.WriteShort(pos.Y * 32)
	c.WriteShort(pos.Z * 32)
	c.Writer.WriteByte(yaw)
	c.Writer.WriteByte(pitch)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v, %v, %v, %v, %v, %v}",
		c.Conn.RemoteAddr(), 0x07, playerId, string(WritePacketUtil_PadString(playerName)), pos.X, pos.Y, pos.Z, yaw, pitch)
}

func (c Client) WritePacket_PlayerTeleport(pos util.Vector3i16, yaw byte, pitch byte, playerId int8) {
	pos.X *= 32
	pos.Y *= 32
	pos.Z *= 32

	c.Writer.WriteByte(0x08)
	c.Writer.WriteByte(byte(playerId))
	c.WriteShort(pos.X)
	c.WriteShort(pos.Y)
	c.WriteShort(pos.Z)
	c.Writer.WriteByte(yaw)
	c.Writer.WriteByte(pitch)
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v, %v, %v, %v, %v}",
		c.Conn.RemoteAddr(), 0x08, playerId, pos.X, pos.Y, pos.Z, yaw, pitch)
}

func (c Client) WritePacket_Message(playerId int8, message string) {
	c.Writer.WriteByte(0x0D)
	c.Writer.WriteByte(byte(playerId))
	c.Writer.Write(WritePacketUtil_PadString(message))
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, msg[%v]}", c.Conn.RemoteAddr(), 0x0D, playerId, message)
}

func (c Client) WritePacket_DisconnectPlayer(message string) {
	c.Writer.WriteByte(0x0E)
	c.Writer.Write(WritePacketUtil_PadString(message))
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, msg[%v]}", c.Conn.RemoteAddr(), 0x0E, string(WritePacketUtil_PadString(message)))
}

func (c Client) WritePacket_ExtInfo(appName string, extensionCount int16) {
	c.Writer.WriteByte(0x10)                           // Packet ID
	c.Writer.Write(WritePacketUtil_PadString(appName)) // AppName
	c.WriteShort(extensionCount)                       // Extension Count
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v}", c.Conn.RemoteAddr(), 0x10, string(WritePacketUtil_PadString(appName)), extensionCount)
}

func (c Client) WritePacket_ExtEntry(extName string, version int32) {
	c.Writer.WriteByte(0x11)                           // Packet ID
	c.Writer.Write(WritePacketUtil_PadString(extName)) // ExtName
	c.WriteInt(version)                                // Version
	c.Writer.Flush()

	logging.Log_Debugf("[%v] [Write] {%v, %v, %v}", c.Conn.RemoteAddr(), 0x11, string(WritePacketUtil_PadString(extName)), version)
}

// Utils

func WritePacketUtil_PadString(s string) []byte {
	var raw [64]byte
	copy(raw[:], s)

	for i := len(s); i < 64; i++ {
		raw[i] = 0x20 // ASCII space padding
	}

	return raw[:]
}

func (c Client) WritePacketUtil_SendLevel(l *Level) error {
	c.WritePacket_LevelInit()

	data, err := l.Gzip()

	if err != nil {
		return err
	}

	totalChunks := int(math.Ceil(float64(len(data)) / float64(1024)))

	for i := 0; i < totalChunks; i++ {
		chunk := make([]byte, 1024)

		last := 0

		for z := 0; z < 1024; z++ {
			zz := i*1024 + z

			if zz >= len(data) { // For last chunk; Check for incomplete chunk to add padding
				chunk[z] = 0x00
				last++
			} else {
				chunk[z] = data[zz]
			}
		}

		if last != 0 {
			c.WritePacket_LevelDataChunk(last, chunk, byte(((float32(i)+1.0)/float32(totalChunks))*100))
		} else {
			c.WritePacket_LevelDataChunk(1024, chunk, byte(((float32(i)+1.0)/float32(totalChunks))*100))
		}
	}

	c.WritePacket_LevelFinalize(l.Size.X, l.Size.Y, l.Size.Z)

	return err
}

package core

import (
	"bytes"
	"compress/gzip"
	"midnight/pkg/util"
)

type Level struct {
	Name        string
	Size        util.Vector3i16
	SpawnPos    util.Vector3i16
	BlocksTotal int32
	Data        []byte
	Players     []Player
}

func ConstructLevel(name string, x int16, y int16, z int16) *Level {
	l := new(Level)
	l.Name = name
	l.Size.X = x
	l.Size.Y = y
	l.Size.Z = z

	l.BlocksTotal = int32(l.Size.X) * int32(l.Size.Y) * int32(l.Size.Z)

	l.SpawnPos = util.Vector3i16{
		X: x / 2,
		Y: (z / 2) + 5,
		Z: y / 2,
	}

	return l
}

func (l *Level) GenerateFlat() {
	l.Data = make([]byte, l.BlocksTotal)

	sizeX, sizeY, sizeZ := int(l.Size.X), int(l.Size.Y), int(l.Size.Z)

	for x := 0; x < sizeX; x++ {
		for z := 0; z < sizeZ; z++ {
			for y := 0; y < sizeY; y++ {
				var b byte

				if y > sizeY/2 {
					b = 0
				} else {
					b = 1
				}

				//l.Data[x+sizeX*(z+sizeZ*y)] = b
				l.Data[x+sizeX*(z+sizeZ*y)] = b
			}
		}
	}
}

func (l *Level) ChangeBlock(block byte, pos util.Vector3i16) {
	sizeX, sizeZ := int(l.Size.X), int(l.Size.Y)
	x, y, z := int(pos.X), int(pos.Y), int(pos.Z)
	l.Data[x+sizeX*(z+sizeZ*y)] = block

	for i := 0; i < len(l.Players); i++ {
		l.Players[i].Cli.WritePacket_SetBlock(block, pos.X, pos.Y, pos.Z)
		l.Players[i].Cli.WritePacket_Message(-1, "&fChanged block")
	}
}

// Level Utils

func (l Level) Gzip() (data []byte, err error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	blocks := []byte{
		byte(l.BlocksTotal >> 24),
		byte(l.BlocksTotal >> 16),
		byte(l.BlocksTotal >> 8),
		byte(l.BlocksTotal & 0xFF),
	}

	_, err = gz.Write(append(blocks, l.Data...))

	if err != nil {
		return nil, err
	}

	err = gz.Close()

	return buf.Bytes(), err
}

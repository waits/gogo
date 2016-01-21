package model

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type Game struct {
	Id       string
	Black    string
	White    string
	Size     uint8
	Handicap uint8
	Turn     uint16
	Board    [19][19]uint8
	Captured [2]uint8
}

// Create a game using a hash of the game parameters as an ID
func NewGame(black string, white string, size uint8) *Game {
	time := time.Now().Unix()
	uniq := []byte(strconv.FormatInt(time, 10) + white + black)
	checksum := sha256.Sum224(uniq)
	trunc := checksum[:7]
	hexid := hex.EncodeToString(trunc)

	g := &Game{Id: hexid, White: white, Black: black, Size: size, Turn: 1}
	client.Cmd("HMSET", "game:"+hexid, "black", g.Black, "white", g.White, "size", g.Size, "turn", g.Turn)
	return g
}

// Loads a game for a provided id
func LoadGame(id string) (*Game, error) {
	resp, err := client.Cmd("HGETALL", "game:"+id).Map()
	if err != nil {
		return nil, err
	}
	size, _ := strconv.Atoi(resp["size"])
	turn, _ := strconv.Atoi(resp["turn"])
	g := &Game{Id: id, Black: resp["black"], White: resp["white"], Size: uint8(size), Turn: uint16(turn)}
	return g, nil
}

func (g *Game) Up() string {
	if g.Turn % 2 == 0 {
		return g.Black
	} else {
		return g.White
	}
}

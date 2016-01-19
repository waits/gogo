package main

import "crypto/sha256"
import "encoding/hex"
import "strconv"
import "time"

type Game struct {
	Id       string
	White    string
	Black    string
	Size     uint8
	Handicap uint8
	Turn     uint8
	Board    [19][19]uint8
	Captured [2]uint8
}

// Create a game using a hash of the game parameters as an ID
func createGame(white string, black string) *Game {
	time := time.Now().Unix()
	uniq := []byte(strconv.FormatInt(time, 10) + white + black)
	checksum := sha256.Sum224(uniq)
	trunc := checksum[:7]
	hexid := hex.EncodeToString(trunc)

	return &Game{Id: hexid, White: white, Black: black, Size: 19, Turn: 0}
}

// Loads a game for a provided id
func loadGame(id string) *Game {
	game := &Game{Id: id, White: "Frank", Black: "Joe", Size: 19, Turn: 0}
	return game
}

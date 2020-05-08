package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type Message struct {
	Text     string `json:"text" firestore:"text"`
	Audience string `json:"audience" firestore:"audience"`
	Bingo    bool   `json:"bingo" firestore:"bingo"`
}

func (m *Message) Set(a string, t string, args ...interface{}) {
	m.Audience = a
	m.Text = fmt.Sprintf(t, args...)
}

// Game is the master structure for the game
type Game struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Master Master `json:"master" firestore:"-"`
	Active bool   `json:"active"`
}

// NewBoard creates a new board for a user.
func (g *Game) NewBoard(p Player) Board {
	b := Board{}
	b.Game = g.ID
	b.Player = p
	b.Load(g.Master.Phrases())
	return b

}

// JSON Returns the given Board struct as a JSON string
func (g Game) JSON() (string, error) {

	bytes, err := json.Marshal(g)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// Master is the collection of all of the people who have selected which
// element in the game
type Master struct {
	Records []Record `json:"record"`
}

// Load adds the master list of phrases to the game.
func (m *Master) Load(p []Phrase) {
	for _, v := range p {
		r := Record{}
		r.Phrase = v
		m.Records = append(m.Records, r)
	}
}

// Phrases returns the List of phrases to populate boards.
func (m Master) Phrases() []Phrase {
	result := []Phrase{}
	for _, v := range m.Records {
		result = append(result, v.Phrase)
	}
	return result
}

// Select marks a phrase as selected by one or more players
func (m *Master) Select(ph Phrase, pl Player) Record {
	r := Record{}
	for i, v := range m.Records {

		if v.Phrase.ID == ph.ID {
			if v.Players.IsMember(pl) {
				fmt.Printf("Was member, removing.  \n")
				new := v.Players.Remove(pl)
				v.Players = new

				if len(new) == 0 {
					v.Phrase.Selected = false
				}
				m.Records[i] = v
				return v
			}
			fmt.Printf("Was not member, adding.  \n")
			v.Phrase.Selected = true
			v.Players = append(v.Players, pl)
			m.Records[i] = v
			return v
		}
	}
	return r
}

// Record is a structure that keeps track of who has selected which Phrase
type Record struct {
	ID      string  `json:"id"`
	Phrase  Phrase  `json:"phrase"`
	Players Players `json:"players"`
}

// Player is a human user who is playing the game.
type Player struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
}

// Players is a slice of Player.
type Players []Player

// IsMember checks to see if a player is in the collection already
func (ps Players) IsMember(p Player) bool {
	for _, v := range ps {
		if v.Email == p.Email {
			return true
		}
	}
	return false
}

// Remove removes a particular player from the list.
func (ps *Players) Remove(p Player) Players {
	out := Players{}
	for _, v := range *ps {
		if v.Email != p.Email {
			out = append(out, v)
		}
	}
	return out
}

// Add adds a particular player from the list.
func (ps *Players) Add(p Player) {
	out := Players{}
	out = append(out, p)
	ps = &out
	return
}

// Board is an individual board that the players use to play bingo
type Board struct {
	ID      string   `json:"id"`
	Game    string   `json:"game"`
	Player  Player   `json:"player"`
	Phrases []Phrase `json:"phrases" firestore:"-"`
}

// Bingo determins if the correct sequence of items have been Selected to
// make bingo on this board.
func (b Board) Bingo() bool {
	diag1 := []string{"B1", "I2", "N3", "G4", "O5"}
	diag2 := []string{"B5", "I4", "N3", "G2", "O1"}
	counts := make(map[string]int)

	for _, v := range b.Phrases {
		if v.Selected {
			counts[v.Column]++
			counts[v.Row]++
		}

		for _, sub := range diag1 {
			if sub == v.Position() {
				counts["diag1"]++
				continue
			}
		}

		for _, sub := range diag2 {
			if sub == v.Position() {
				counts["diag2"]++
				continue
			}
		}
	}

	fmt.Printf("%+v\n", counts)

	for _, v := range counts {
		if v == 5 {
			return true
		}
	}
	return false
}

// Select records if a phrase on the board has been selected.
func (b *Board) Select(ph Phrase) Phrase {
	for i, v := range b.Phrases {
		if v.ID == ph.ID {
			if v.Selected {
				v.Selected = false
				b.Phrases[i] = v
				return v
			}
			v.Selected = true
			b.Phrases[i] = v
			return v
		}
	}
	return ph
}

// Load adds the phrases to the board and randomly orders them.
func (b *Board) Load(p []Phrase) {
	rand.Seed(randseedfunc())
	rand.Shuffle(len(p), func(i, j int) { p[i], p[j] = p[j], p[i] })

	for i, v := range p {
		v.Selected = false
		v.Column, v.Row = b.CalcColumnsRows(i + 1)
		p[i] = v
	}
	b.Phrases = p
}

func (b *Board) CalcColumnsRows(i int) (string, string) {
	column := ""
	row := ""

	switch i % 5 {
	case 1:
		column = "B"
	case 2:
		column = "I"
	case 3:
		column = "N"
	case 4:
		column = "G"
	default:
		column = "O"
	}

	row = strconv.Itoa(int(math.Round(float64((i - 1) / 5))))

	return column, row
}

// JSON Returns the given Board struct as a JSON string
func (b Board) JSON() (string, error) {

	bytes, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

func randomseed() int64 {
	return time.Now().UnixNano()
}

// Phrase represents a statement, event or other such thing that we are on the
// lookout for in this game of bingo.
type Phrase struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Selected bool   `json:"selected"`
	Row      string `json:"row"`
	Column   string `json:"column"`
}

// Position returns the combined Row and Column of the Phrase
func (p Phrase) Position() string {
	return p.Column + p.Row
}

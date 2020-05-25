package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// Message is a message that will be broadcast from the server to all players
type Message struct {
	ID        string   `json:"id" firestore:"id"`
	Text      string   `json:"text" firestore:"text"`
	Audience  []string `json:"audience" firestore:"audience"`
	Bingo     bool     `json:"bingo" firestore:"bingo"`
	Operation string   `json:"operation" firestore:"operation"`
}

// SetText sets the text of the broadcast message
func (m *Message) SetText(t string, args ...interface{}) {
	m.Text = fmt.Sprintf(t, args...)
}

// SetAudience adds the recipients to the messaage
func (m *Message) SetAudience(a ...string) {
	m.Audience = a
}

// Game is the master structure for the game
type Game struct {
	ID      string  `json:"id" firestore:"id"`
	Name    string  `json:"name" firestore:"name"`
	Active  bool    `json:"active" firestore:"active"`
	Players Players `json:"players" firestore:"-"`
	Admins  Players `json:"admins" firestore:"-"`
	master  Master  `json:"master" firestore:"-"`
}

// NewBoard creates a new board for a user.
func (g *Game) NewBoard(p Player) Board {
	b := Board{}
	b.Game = g.ID
	b.Player = p
	b.Load(g.master.Phrases())
	return b

}

// Games is a collection of game objects.
type Games []Game

// JSON marshalls the content of a slice of games to json.
func (g Games) JSON() (string, error) {
	bytes, err := json.Marshal(g)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// Report represents a report of all of the selected phraes.
type Report struct {
	Phrase  Phrase  `json:"phrase"`
	Percent float32 `json:"percent"`
	Count   int     `json:"count"`
	Total   int     `json:"total"`
}

// Reports are a slice of reports.
type Reports []Report

// IsDubious checks to see if any of the boards claiming bingo match everyone else
func (r Reports) IsDubious() bool {
	threshold := float32(.5)
	count := 2
	actual := 0

	for _, v := range r {
		if v.Percent < threshold {
			actual++
		}
		if actual > count {
			return true
		}
	}

	return false
}

// CheckBoard checks a particular board against the master records
func (g *Game) CheckBoard(b Board) Reports {

	results := Reports{}
	total := g.CountPlayers()

	for _, v := range b.Phrases {
		if v.Selected && v.Text != "FREE" {

			master := g.FindRecord(v)
			r := Report{}
			r.Phrase = v
			r.Percent = float32(len(master.Players)) / float32(total)
			r.Count = len(master.Players)
			r.Total = total
			results = append(results, r)
		}
	}

	return results
}

// FindRecord retrieves the report of a particular phrase
func (g Game) FindRecord(p Phrase) Record {
	for _, v := range g.master.Records {
		if v.Phrase.ID == p.ID {
			return v
		}
	}
	return Record{}
}

// CountPlayers returns the count of all players who have selected phrases/
func (g *Game) CountPlayers() int {
	return len(g.Players)
}

// IsAdmin determines if a player is an admin for the game.
func (g *Game) IsAdmin(p Player) bool {
	return g.Admins.IsMember(p)
}

// Select marks a phrase as selected by one or more players
func (g *Game) Select(ph Phrase, pl Player) Record {
	return g.master.Select(ph, pl)
}

// JSON marshalls the content of a game to json.
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
	Records []Record `json:"record" firestore:"records"`
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
				new := v.Players.Remove(pl)
				v.Players = new

				if len(new) == 0 {
					v.Phrase.Selected = false
				}
				m.Records[i] = v
				return v
			}
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
	ID      string  `json:"id"  firestore:"id"`
	Phrase  Phrase  `json:"phrase"  firestore:"phrase"`
	Players Players `json:"players"  firestore:"players"`
}

// Player is a human user who is playing the game.
type Player struct {
	Name  string `json:"name"  firestore:"name"`
	Email string `json:"email"  firestore:"email"`
}

// JSON Returns the given Board struct as a JSON string
func (p Player) JSON() (string, error) {

	bytes, err := json.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
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
	ID            string   `json:"id" firestore:"id"`
	Game          string   `json:"game" firestore:"game"`
	Player        Player   `json:"player" firestore:"player"`
	BingoDeclared bool     `json:"bingodeclared" firestore:"bingodeclared"`
	Phrases       []Phrase `json:"phrases" firestore:"-"`
}

// Bingo determins if the correct sequence of items have been Selected to
// make bingo on this board.
func (b *Board) Bingo() bool {
	diag1 := []string{"B0", "I1", "N2", "G3", "O4"}
	diag2 := []string{"B4", "I3", "N2", "G1", "O0"}
	counts := make(map[string]int)

	for _, v := range b.Phrases {
		if v.Selected {
			counts[v.Column]++
			counts[v.Row]++

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
	}
	b.log(fmt.Sprintf("%v", counts))
	for _, v := range counts {
		if v == 5 {
			b.log("Bingo Declared")
			b.BingoDeclared = true
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
				b.log(fmt.Sprintf("Unselected %s", v.Position()))
				v.Selected = false
				b.Phrases[i] = v
				return v
			}
			b.log(fmt.Sprintf("Selected %s", v.Position()))
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

	free := 0
	center := 12

	for i, v := range p {

		v.Selected = false

		if v.Text == "FREE" {
			free = i
			v.Selected = true
		}
		p[i] = v

	}

	p[free], p[center] = p[center], p[free]

	for i, v := range p {
		v.Column, v.Row = calcColumnsRows(i)
		v.DisplayOrder = i
		p[i] = v
	}

	b.Phrases = p
}

func calcColumnsRows(i int) (string, string) {
	column := ""
	row := ""

	switch i % 5 {
	case 1:
		column = "I"
	case 2:
		column = "N"
	case 3:
		column = "G"
	case 4:
		column = "O"
	default:
		column = "B"
	}

	row = strconv.Itoa(int(math.Round(float64((i) / 5))))

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
	ID           string `json:"id" firestore:"id"`
	Text         string `json:"text" firestore:"text"`
	Selected     bool   `json:"selected" firestore:"selected"`
	Row          string `json:"row" firestore:"-"`
	Column       string `json:"column" firestore:"-"`
	DisplayOrder int    `json:"displayorder" firestore:"displayorder"`
}

// Position returns the combined Row and Column of the Phrase
func (p Phrase) Position() string {
	return p.Column + p.Row
}

func (b Board) log(msg string) {
	if noisy {
		log.Printf("Bingo: %s\n", msg)
	}
}

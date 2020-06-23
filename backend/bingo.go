// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

// I stole this code from firestore/collref.go basically it generates the ids
// so I can use batch sets instead of adds for anything
const alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func uniqueID() string {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("agent: crypto/rand.Read error: %v", err))
	}
	for i, byt := range b {
		b[i] = alphanum[int(byt)%len(alphanum)]
	}
	return string(b)
}

// Message is a message that will be broadcast from the server to all players
type Message struct {
	ID        string   `json:"id" firestore:"id"`
	Text      string   `json:"text" firestore:"text"`
	Audience  []string `json:"audience" firestore:"audience"`
	Bingo     bool     `json:"bingo" firestore:"bingo"`
	Operation string   `json:"operation" firestore:"operation"`
	Received  bool     `json:"received" firestore:"received"`
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
	ID      string           `json:"id" firestore:"id"`
	Name    string           `json:"name" firestore:"name"`
	Active  bool             `json:"active" firestore:"active"`
	Players Players          `json:"players" firestore:"-"`
	Admins  Players          `json:"admins" firestore:"-"`
	Master  Master           `json:"master" firestore:"-"`
	Boards  map[string]Board `json:"boards" firestore:"-"`
	Created time.Time        `json:"created" firestore:"created"`
}

// NewGame initializes a new game object
func NewGame(name string, player Player, phrases []Phrase) Game {
	g := Game{}
	g.ID = uniqueID()
	g.Name = name
	g.Active = true
	g.Created = time.Now().UTC().Truncate(time.Millisecond)
	g.Boards = make(map[string]Board)
	g.Admins.Add(player)
	g.Players.Add(player)
	g.Master.Load(phrases)

	return g
}

// Obscure will obscure the email address of every email in the game other than
// the one that is input.
func (g *Game) Obscure(email string) {
	g.Players.Obscure(email)
	g.Admins.Obscure(email)

	for _, v := range g.Master.Records {
		v.Players.Obscure(email)
	}

	for i, v := range g.Boards {
		v.Obscure(email)
		g.Boards[i] = v
	}
}

// NewBoard creates a new board for a user.
func (g *Game) NewBoard(player Player) Board {
	b := InitBoard()
	b.log("Creating new board ")
	b.ID = uniqueID()
	b.Game = g.ID
	b.Player = player
	b.Load(g.Master.Phrases())
	g.Players.Add(player)
	g.Boards[b.ID] = b

	return b
}

// InitBoard creates a new board and inits Phrases
func InitBoard() Board {
	b := Board{}
	b.Phrases = make(map[string]Phrase)
	return b
}

// UpdatePhrase will change a given phrase in the master record of phrases.
func (g *Game) UpdatePhrase(phrase Phrase) {
	i, r := g.FindRecord(phrase)
	phrase.Selected = false
	r.Phrase = phrase
	r.Players = Players{}
	g.Master.Records[i] = r

	for _, b := range g.Boards {
		b.UpdatePhrase(phrase)
	}

}

// DeleteBoard removes a board from the game.
func (g *Game) DeleteBoard(board Board) {
	g.Master.RemovePlayer(board.Player)
	delete(g.Boards, board.ID)
}

// Games is a collection of game objects.
type Games []Game

// Add adds a particular game to the list.
func (gs *Games) Add(game Game) {
	*gs = append(*gs, game)
	return
}

// Sort orders Games by createdon
func (gs *Games) Sort() {
	tmp := *gs
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Created.Before(tmp[j].Created)
	})
	gs = &tmp
	return
}

// JSON marshalls the content of a slice of games to json.
func (gs Games) JSON() (string, error) {
	gs.Sort()
	bytes, err := json.Marshal(gs)
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
func (g *Game) CheckBoard(board Board) Reports {

	results := Reports{}
	total := len(g.Players)

	for _, v := range board.Phrases {
		if v.Selected && v.Text != "FREE" {

			_, record := g.FindRecord(v)
			r := Report{}
			r.Phrase = v
			r.Percent = float32(len(record.Players)) / float32(total)
			r.Count = len(record.Players)
			r.Total = total
			results = append(results, r)
		}
	}

	return results
}

// FindRecord retrieves the report of a particular phrase
func (g Game) FindRecord(phrase Phrase) (int, Record) {
	for i, v := range g.Master.Records {
		if v.Phrase.ID == phrase.ID {
			return i, v
		}
	}
	return -1, Record{}
}

// IsAdmin determines if a player is an admin for the game.
func (g *Game) IsAdmin(player Player) bool {
	return g.Admins.IsMember(player)
}

// Select marks a phrase as selected by one or more players
func (g *Game) Select(phrase Phrase, player Player) Record {

	for _, v := range g.Boards {
		if v.Player.Email == player.Email {
			v.Select(phrase)
		}
	}

	return g.Master.Select(phrase, player)
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
func (m *Master) Load(phrases []Phrase) {
	for _, v := range phrases {
		r := Record{}
		r.ID = v.ID
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
func (m *Master) Select(phrase Phrase, player Player) Record {
	r := Record{}
	for i, v := range m.Records {

		if v.Phrase.ID == phrase.ID {
			if v.Players.IsMember(player) {
				v.Players.Remove(player)

				if !phrase.Selected {
					v.Phrase.Selected = false
				}
				m.Records[i] = v
				return v
			}
			v.Phrase.Selected = phrase.Selected
			m.Records[i] = v
			m.Records[i].Players.Add(player)
			return v
		}
	}
	return r
}

// RemovePlayer removes a selected player's selection from the master list.
func (m *Master) RemovePlayer(player Player) {
	for i := range m.Records {
		m.Records[i].Players.Remove(player)
		if len(m.Records[i].Players) == 0 {
			m.Records[i].Phrase.Selected = false
		}
	}
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

// Obscure will replace the email, if it isn't the one input.
func (p *Player) Obscure(email string) {
	if p.Email != email {
		p.Email = "xxxxxx@xxxxxx.xxx"
	}
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

// Obscure will obscure the email of any email that isn't the email input.
func (ps Players) Obscure(email string) {
	for i := range ps {
		ps[i].Obscure(email)
	}
}

// IsMember checks to see if a player is in the collection already
func (ps Players) IsMember(player Player) bool {
	for _, v := range ps {
		if v.Email == player.Email {
			return true
		}
	}
	return false
}

// Remove removes a particular player from the list.
func (ps *Players) Remove(player Player) {
	new := Players{}
	for _, v := range *ps {
		if v.Email != player.Email {
			new = append(new, v)
		}
	}
	*ps = new
	return
}

// Add adds a particular player from the list.
func (ps *Players) Add(player Player) {
	for _, v := range *ps {
		if player.Email == v.Email {
			return
		}
	}
	*ps = append(*ps, player)
	return
}

// Sort orders Players by email
func (ps *Players) Sort() {
	tmp := *ps
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Email < tmp[j].Email
	})
	ps = &tmp
	return
}

// JSON Returns the given list of players struct as a JSON string
func (ps Players) JSON() (string, error) {

	bytes, err := json.Marshal(ps)
	if err != nil {
		return "", fmt.Errorf("could not marshal json for response: %s", err)
	}

	return string(bytes), nil
}

// Board is an individual board that the players use to play bingo
type Board struct {
	ID            string  `json:"id" firestore:"id"`
	Game          string  `json:"game" firestore:"game"`
	Player        Player  `json:"player" firestore:"player"`
	BingoDeclared bool    `json:"bingodeclared" firestore:"bingodeclared"`
	Phrases       Phrases `json:"phrases" firestore:"-"`
}

// Obscure obscures the email of the board's player
func (b *Board) Obscure(email string) {
	b.Player.Obscure(email)
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
	b.log(fmt.Sprintf("%+v", counts))
	for _, v := range counts {
		if v == 5 {
			b.log("Bingo Declared")
			b.BingoDeclared = true
			return true
		}
	}
	b.BingoDeclared = false
	return false
}

// Select records if a phrase on the board has been selected.
func (b *Board) Select(phrase Phrase) Phrase {
	v := b.Phrases[phrase.ID]
	v.Selected = phrase.Selected
	b.Phrases[phrase.ID] = v
	return v
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
		b.Phrases[v.ID] = v
	}
}

// UpdatePhrase change the text of a given phrases.
func (b *Board) UpdatePhrase(phrase Phrase) {

	v := b.Phrases[phrase.ID]
	v.Text = phrase.Text
	v.Selected = false
	b.Phrases[phrase.ID] = v

	return
}

// Print prints out the board for debugging
func (b *Board) Print() {

	phrases := make(Phrases, len(b.Phrases))

	for i, v := range b.Phrases {
		phrases[i] = v
	}

	sorted := phrases.ByDisplayOrder()

	fmt.Printf("|*************** %s   ****************|\n", b.ID)
	for i, v := range sorted {
		text := strings.ToLower(v.Text)
		if len(text) > 10 {
			text = text[0:9]
		}

		if v.Selected {
			text = strings.ToUpper(text)
		}

		fmt.Printf("|%s%s-%-10v|", v.Column, v.Row, text)
		if (i+1)%5 == 0 {
			fmt.Printf("\n")
		}
	}
	return
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
	Row          string `json:"row" firestore:"row"`
	Column       string `json:"column" firestore:"column"`
	DisplayOrder int    `json:"displayorder" firestore:"displayorder"`
}

// Position returns the combined Row and Column of the Phrase
func (p Phrase) Position() string {
	return p.Column + p.Row
}

func (b Board) log(msg string) {
	if noisy {
		log.Printf("Bingo     : %s\n", msg)
	}
}

// Phrases is a map of Phrase structs
type Phrases map[string]Phrase

// ByDisplayOrder returns a slice of Phrase sorted by DisplayOrder
func (ps Phrases) ByDisplayOrder() []Phrase {
	phrases := make([]Phrase, len(ps))

	for _, v := range ps {
		phrases[v.DisplayOrder] = v
	}

	return phrases
}

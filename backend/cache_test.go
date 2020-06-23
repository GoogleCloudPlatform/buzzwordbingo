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
	"sort"
	"testing"

	"github.com/gomodule/redigo/redis"
)

var mockpool = GetMockPool()

type MockPool struct {
	store   map[string]string
	queue   map[string]string
	waiting bool
}

func GetMockPool() MockPool {
	m := MockPool{}
	m.store = make(map[string]string)
	m.queue = make(map[string]string)
	return m
}

type MockAgent struct {
}

// Get Returns a Mock Connection
func (m MockPool) Get() redis.Conn {
	return MockConnection{}
}

type MockConnection struct {
}

// Close closes the connection.
func (m MockConnection) Close() error {
	return nil
}

// Err returns a non-nil value when the connection is not usable.
func (m MockConnection) Err() error {
	return nil
}

// Do sends a command to the server and returns the received reply.
func (m MockConnection) Do(commandName string, args ...interface{}) (interface{}, error) {
	result := ""
	var ok bool
	switch commandName {
	case "FLUSHALL":
		mockpool.store = make(map[string]string)
	case "SET":
		mockpool.store[args[0].(string)] = args[1].(string)
	case "GET":
		key, _ := args[0].(string)
		result, ok = mockpool.store[key]

		if !ok {
			return "", redis.ErrNil
		}

	case "EXEC":
		for i, v := range mockpool.queue {
			mockpool.store[i] = v
		}
		mockpool.waiting = false
		mockpool.queue = make(map[string]string)

	}
	return result, nil
}

// Flush flushes the output buffer to the Redis server.
func (m MockConnection) Flush() error {
	return nil
}

// Send writes the command to the client's output buffer.
func (m MockConnection) Send(commandName string, args ...interface{}) error {
	switch commandName {
	case "SET":
		mockpool.queue[args[0].(string)] = args[1].(string)
	case "DEL":
		delete(mockpool.store, args[0].(string))
	case "MULTI":
		mockpool.waiting = true
	}

	return nil
}

// Receive receives a single reply from the Redis server
func (m MockConnection) Receive() (interface{}, error) {
	return "", nil
}

func cacheTestSetup() {
	cache = &Cache{}
	cache.enabled = true
	cache.redisPool = mockpool
}

func TestCacheBoard(t *testing.T) {
	board := getTestBoard()
	player := Player{}
	player.Email = "test@example.com"
	board.Player = player

	if err := cache.SaveBoard(board); err != nil {
		t.Errorf("Cache.SaveBoard() err want %v got %s ", nil, err)
	}

	boardFromCache, err := cache.GetBoard(board.ID)

	if err != nil {
		t.Errorf("Cache.GetBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromCache) {
		t.Errorf("Cache.GetBoard() should return an unchanged board")
	}

	boardFromCacheForPlayer, err := cache.GetBoardForPlayer(board.Game, board.Player.Email)

	if err != nil {
		t.Errorf("Cache.GetBoard() err want %v got %s ", nil, err)
	}

	if !boardEquals(board, boardFromCacheForPlayer) {
		t.Errorf("Cache.GetBoard() err want %+v got %+v ", board, boardFromCache)
	}

	if err := cache.DeleteBoard(board); err != nil {
		t.Errorf("Cache.DeleteBoard() err want %v got %s ", nil, err)
	}

	boardFromCachePostDelete, err := cache.GetBoard(board.ID)
	if err != ErrCacheMiss {
		t.Errorf("Cache.GetBoard() post delete err want %v got %v ", nil, boardFromCachePostDelete)
	}
}

func TestClearCache(t *testing.T) {
	board := getTestBoard()
	player := Player{}
	player.Email = "test@example.com"
	board.Player = player
	game := NewGame("test game", player, getTestPhrases())

	if err := cache.SaveBoard(board); err != nil {
		t.Errorf("Cache.SaveBoard() err want %v got %s ", nil, err)
	}

	if err := cache.SaveGame(game); err != nil {
		t.Errorf("Cache.SaveGame() err want %v got %s ", nil, err)
	}

	if err := cache.Clear(); err != nil {
		t.Errorf("Cache.Clear() err want %v got %s ", nil, err)
	}

	boardFromCachePostClear, err := cache.GetBoard(board.ID)
	if err != ErrCacheMiss {
		t.Errorf("Cache.GetBoard() post delete err want %v got %v ", nil, boardFromCachePostClear)
	}

	gameFromCachePostDelete, err := cache.GetGame(game.ID)
	if err != ErrCacheMiss {
		t.Errorf("Cache.GetBoard() post delete err want %v got %v ", nil, gameFromCachePostDelete)
	}
}

func TestUpdatePhrase(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"
	game := NewGame("test game", player, getTestPhrases())
	phrase := getTestPhrases()[0]
	phrase.Text = "Totally new text"
	board := game.NewBoard(player)

	if err := cache.SaveBoard(board); err != nil {
		t.Errorf("Cache.SaveBoard() err want %v got %s ", nil, err)
	}

	if err := cache.SaveGame(game); err != nil {
		t.Errorf("Cache.SaveGame() err want %v got %s ", nil, err)
	}

	game.UpdatePhrase(phrase)

	if err := cache.UpdatePhrase(game, phrase); err != nil {
		t.Errorf("Cache.UpdatePhrase() err want %v got %s ", nil, err)
	}

	gameFromCache, err := cache.GetGame(game.ID)
	if err != nil {
		t.Errorf("Cache.GetGame() err want %v got %s ", nil, err)
	}

	boardFromCache, err := cache.GetBoard(board.ID)
	if err != nil {
		t.Errorf("Cache.GetBoard() err want %v got %s ", nil, err)
	}

	boardFromCacheForPlayer, err := cache.GetBoardForPlayer(board.Game, board.Player.Email)
	if err != nil {
		t.Errorf("Cache.GetBoard() err want %v got %s ", nil, err)
	}

	_, record := gameFromCache.FindRecord(phrase)
	if record.Phrase.Text != phrase.Text {
		t.Errorf("Cache.UpdatePhrase() game master phrase text want %s got %s ", phrase.Text, record.Phrase.Text)
	}

	for _, v := range gameFromCache.Boards {
		if v.Phrases[phrase.ID].Text != phrase.Text {
			t.Errorf("Cache.UpdatePhrase() game baords phrase text want %s got %s ", phrase.Text, v.Phrases[phrase.ID].Text)
		}
	}

	if boardFromCache.Phrases[phrase.ID].Text != phrase.Text {
		t.Errorf("Cache.UpdatePhrase() boardFromCache phrase text want %s got %s ", phrase.Text, boardFromCache.Phrases[phrase.ID].Text)
	}

	if boardFromCacheForPlayer.Phrases[phrase.ID].Text != phrase.Text {
		t.Errorf("Cache.UpdatePhrase() boardFromCacheForPlayer phrase text want %s got %s ", phrase.Text, boardFromCacheForPlayer.Phrases[phrase.ID].Text)
	}

}

func TestCacheGame(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"
	game := NewGame("test game", player, getTestPhrases())

	if err := cache.SaveGame(game); err != nil {
		t.Errorf("Cache.SaveGame() err want %v got %s ", nil, err)
	}

	gameFromCache, err := cache.GetGame(game.ID)

	if err != nil {
		t.Errorf("Cache.GetGame() err want %v got %s ", nil, err)
	}

	if !gameEquals(game, gameFromCache) {
		t.Errorf("Cache.GetGame() should return an unchanged game")
	}

	if err := cache.DeleteGame(game); err != nil {
		t.Errorf("Cache.DeleteGame() err want %v got %s ", nil, err)
	}

	gameFromCachePostDelete, err := cache.GetGame(game.ID)
	if err != ErrCacheMiss {
		t.Errorf("Cache.GetBoard() post delete err want %v got %v ", nil, gameFromCachePostDelete)
	}
}

func TestCacheGames(t *testing.T) {
	player := Player{}
	player.Email = "test@example.com"
	game1 := NewGame("test game", player, getTestPhrases())
	game2 := NewGame("test game2", player, getTestPhrases())

	games := Games{}
	games = append(games, game1)
	games = append(games, game2)
	key := "uniqueid"

	if err := cache.SaveGamesForKey(key, games); err != nil {
		t.Errorf("Cache.SaveGamesForKey() err want %v got %s ", nil, err)
	}

	gamesFromCache, err := cache.GetGamesForKey(key)

	if err != nil {
		t.Errorf("Cache.GetGamesForKey() err want %v got %s ", nil, err)
	}

	if !gamesEquals(games, gamesFromCache) {
		t.Errorf("Cache.GetGames() should return an unchanged set of games")
	}

	if err := cache.DeleteGamesForKey([]string{key}); err != nil {
		t.Errorf("Cache.DeleteGamesForKey() err want %v got %s ", nil, err)
	}

	gamesFromCachePostDelete, err := cache.GetGamesForKey(key)
	if err != ErrCacheMiss {
		t.Errorf("Cache.GetGamesForKey() post delete err want %v got %v ", nil, gamesFromCachePostDelete)
	}
}

func boardEquals(b1, b2 Board) bool {
	if b1.ID != b2.ID {
		return false
	}

	if b1.Game != b2.Game {
		return false
	}

	if b1.Player.Email != b2.Player.Email {
		return false
	}

	if b1.Player.Name != b2.Player.Name {
		return false
	}

	for i, v := range b1.Phrases {
		if !phraseEquals(v, b2.Phrases[i]) {
			return false
		}

	}
	return true
}

func gameEquals(g1, g2 Game) bool {

	if g1.ID != g2.ID {
		return false
	}

	if g1.Active != g2.Active {
		return false
	}

	if g1.Name != g2.Name {
		return false
	}

	if g1.Created.Unix() != g2.Created.Unix() {
		return false
	}

	for i, v := range g1.Boards {
		if !boardEquals(v, g2.Boards[i]) {
			return false
		}
	}

	if !playersEquals(g1.Admins, g2.Admins) {
		return false
	}

	if !playersEquals(g1.Players, g2.Players) {
		return false
	}

	return true
}

func gamesEquals(gs1, gs2 Games) bool {
	for i, v := range gs1 {
		if !gameEquals(v, gs2[i]) {
			return false
		}
	}
	return true
}

func playersEquals(pl1, pl2 Players) bool {
	pl1.Sort()
	pl2.Sort()

	for i, v := range pl1 {
		if !playerEquals(v, pl2[i]) {
			return false
		}
	}
	return true
}

func playerEquals(p1, p2 Player) bool {
	if p1.Email != p2.Email {
		return false
	}
	if p1.Name != p2.Name {
		return false
	}

	return true
}

func phraseEquals(p1, p2 Phrase) bool {
	if p1.ID != p2.ID {
		return false
	}

	if p1.Text != p2.Text {
		return false
	}

	if p1.Column != p2.Column {
		return false
	}

	if p1.Row != p2.Row {
		return false
	}

	if p1.DisplayOrder != p2.DisplayOrder {
		return false
	}

	return true
}

func phrasesEquals(ps1, ps2 []Phrase) bool {
	for i, v := range ps1 {
		if !phraseEquals(v, ps2[i]) {
			return false
		}
	}
	return true
}

func masterPhrasesEquals(ps1, ps2 []Phrase) bool {
	sort.Slice(ps1, func(i, j int) bool {
		return ps1[i].ID < ps1[j].ID
	})

	sort.Slice(ps2, func(i, j int) bool {
		return ps2[i].ID < ps2[j].ID
	})

	for i, v := range ps1 {
		if v.ID != ps2[i].ID {
			return false
		}

		if v.Text != ps2[i].Text {
			return false
		}
	}
	return true
}

func masterEquals(m1, m2 Master) bool {
	for i, v := range m1.Records {
		if !recordEquals(v, m2.Records[i]) {
			return false
		}
	}
	return true
}

func recordEquals(r1, r2 Record) bool {
	if r1.ID != r2.ID {
		return false
	}
	if !playersEquals(r1.Players, r2.Players) {
		return false
	}

	if !phraseEquals(r1.Phrase, r2.Phrase) {
		return false
	}

	return true
}

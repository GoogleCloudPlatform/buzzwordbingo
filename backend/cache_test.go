package main

import (
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

	cacheTestSetup()

	if err := cache.SaveBoard(board); err != nil {
		t.Errorf("Cache.SaveBoard() err want %v got %s ", nil, err)
	}

	boardFromCache, err := cache.GetBoard(board.ID)

	if err != nil {
		t.Errorf("Cache.GetBoard() err want %v got %s ", nil, err)
	}

	if !boardsEqual(board, boardFromCache) {
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

func boardsEqual(b1, b2 Board) bool {
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
		if v.ID != b2.Phrases[i].ID {
			return false
		}

		if v.Text != b2.Phrases[i].Text {
			return false
		}

		if v.Column != b2.Phrases[i].Column {
			return false
		}

		if v.Row != b2.Phrases[i].Row {
			return false
		}

		if v.DisplayOrder != b2.Phrases[i].DisplayOrder {
			return false
		}

	}

	return true
}

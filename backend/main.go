package main

func main() {

}

// Game is the master structure for the game
type Game struct {
	Master  Board   `json:"master"`
	IsBingo bool    `json:"is_bingo"`
	Boards  []Board `json:"boards"`
}

// Player is a human user who is playing the game.
type Player struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Admin bool   `json:"admin"`
}

// Board is an individual board that the players use to play bingo
type Board struct {
	Player  Player   `json:"player"`
	Phrases []Phrase `json:"phrases"`
}

// CheckBingo determins if the correct sequence of items have been clicked to
// make bingo on this board.
func (b Board) CheckBingo() bool {
	diag1 := []string{"b1", "i2", "n3", "g4", "o5"}
	diag2 := []string{"b5", "i4", "n3", "g2", "o1"}
	counts := make(map[string]int)

	// var counts = map[string]int{
	// 	"B": 0, "I": 0, "N": 0, "G": 0, "O": 0,
	// 	"1": 0, "2": 0, "3": 0, "4": 0, "5": 0,
	// 	"diag1": 0, "diag2": 0,
	// }

	for _, v := range b.Phrases {
		if v.Clicked {
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

	for _, v := range counts {
		if v == 5 {
			return true
		}
	}
	return false
}

// Phrase represents a statement, event or other such thing that we are on the
// lookout for in this game of bingo.
type Phrase struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	Clicked bool   `json:"clicked"`
	Row     string `json:"row"`
	Column  string `json:"column"`
}

func (p Phrase) Position() string {
	return p.Column + p.Row
}

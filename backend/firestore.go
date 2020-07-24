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
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// NewAgent intializes and returns a fresh agent.
func NewAgent(ctx context.Context, projectID string) (Agent, error) {
	var err error
	rand.Seed(time.Now().UTC().UnixNano())
	a := Agent{}
	a.ProjectID = projectID
	a.ctx = ctx
	a.client, err = firestore.NewClient(a.ctx, a.ProjectID)
	if err != nil {
		return a, fmt.Errorf("Failed to create client: %v", err)
	}

	admins, err := a.GetAdmins()
	if err != nil {
		return a, fmt.Errorf("error trying to check on admins: %v", err)
	}

	if len(admins) == 0 {
		a.log("Intializing admin email.")
		player := Player{"", "notrealemail"}
		err = a.AddAdmin(player)
		if err != nil {
			log.Fatal(err)
		}
	}

	phrases, err := a.GetPhrases()
	if err != nil {
		return a, fmt.Errorf("error trying to check on phrases: %v", err)
	}

	if len(phrases) < 25 {
		phrases = a.getDefaultList()

		if err := a.LoadPhrases(phrases); err != nil {
			return a, fmt.Errorf("error loading phrases: %v", err)
		}
	}

	return a, nil
}

// Agent is a go between for the main application and firestore.
type Agent struct {
	ProjectID string
	ctx       context.Context
	client    *firestore.Client
}

func (a *Agent) log(msg string) {
	if noisy {
		log.Printf("Firestore : %s\n", msg)
	}
}

////////////////////////////////////////////////////////////////////////////////
// ADMINS
////////////////////////////////////////////////////////////////////////////////

// IsAdmin tests if a give player is in the admin group by email
func (a *Agent) IsAdmin(email string) (bool, error) {

	a.log("See if user is in admin collection")
	doc, err := a.client.Collection("admins").Doc(email).Get(a.ctx)
	if err != nil {
		if strings.Contains(err.Error(), "code = NotFound") {
			return false, nil
		}
		return false, fmt.Errorf("failed to query admins: %v", err)
	}

	return doc.Exists(), nil

}

// AddAdmin adds an admin to the over all system
func (a *Agent) AddAdmin(player Player) error {
	if _, err := a.client.Collection("admins").Doc(player.Email).Set(ctx, player); err != nil {
		return fmt.Errorf("unable to add admin: %s", err)
	}
	return nil
}

// DeleteAdmin Deletes an admin to the over all system
func (a *Agent) DeleteAdmin(player Player) error {
	if _, err := a.client.Collection("admins").Doc(player.Email).Delete(ctx); err != nil {
		return fmt.Errorf("unable to delete admin: %s", err)
	}
	return nil
}

// GetAdmins fetches the master list of Admins for populating Games
func (a *Agent) GetAdmins() (Players, error) {

	p := Players{}

	a.log("Getting Phrases")
	iter := a.client.Collection("admins").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return p, fmt.Errorf("Failed to iterate: %v", err)
		}
		var player Player
		doc.DataTo(&player)
		p.Add(player)
	}

	return p, nil
}

////////////////////////////////////////////////////////////////////////////////
// PHRASES
////////////////////////////////////////////////////////////////////////////////

// GetPhrases fetches the master list of Phrases for populating Games
func (a *Agent) GetPhrases() ([]Phrase, error) {

	p := []Phrase{}

	a.log("Getting Phrases")
	iter := a.client.Collection("phrases").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return p, fmt.Errorf("Failed to iterate: %v", err)
		}
		var phrase Phrase
		dataMap := doc.Data()

		phrase.ID = dataMap["id"].(string)
		phrase.Text = dataMap["text"].(string)

		p = append(p, phrase)
	}

	return p, nil
}

// LoadPhrases does a batch load of the master phrases for the game.
func (a *Agent) LoadPhrases(phrases []Phrase) error {
	batch := a.client.Batch()

	for _, v := range phrases {
		ref := a.client.Collection("phrases").Doc(v.ID)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to load phrases to database: %v", err)
	}

	return nil
}

// UpdateMasterPhrase updates a phrase in the master collection of phrases
func (a *Agent) UpdateMasterPhrase(phrase Phrase) error {

	if _, err := a.client.Collection("phrases").Doc(phrase.ID).Set(a.ctx, phrase); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

func (a *Agent) getDefaultList() []Phrase {
	a.log("Getting the default phrase list")
	phrases := []Phrase{
		{"101", "Someone tells a dad joke", false, "", "", 0},
		{"102", "Greg references airplanes/piloting", false, "", "", 0},
		{"103", "\"We’re all in this together\"", false, "", "", 0},
		{"104", "\"the new normal\"", false, "", "", 0},
		{"105", "Someone's child/S.O. on screen", false, "", "", 0},
		{"106", "\"Goals\"", false, "", "", 0},
		{"107", "\"Increased (better, clearer) focus\"", false, "", "", 0},
		{"108", "\"These uncertain times\"", false, "", "", 0},
		{"109", "Someone’s pet on screen", false, "", "", 0},
		{"110", "\"working from home\"", false, "", "", 0},
		{"111", "Someone speaks when muted", false, "", "", 0},
		{"112", "\"Wash your hands\"", false, "", "", 0},
		{"113", "FREE", false, "", "", 0},
		{"114", "Awkward silence", false, "", "", 0},
		{"115", "Sports metaphor", false, "", "", 0},
		{"116", "Start at least 5 min late", false, "", "", 0},
		{"117", "Joke made, but no one laughs", false, "", "", 0},
		{"118", "Someone eats on screen", false, "", "", 0},
		{"119", "Answer all Dory questions", false, "", "", 0},
		{"120", "\"self care\"", false, "", "", 0},
		{"121", "\"Can you see my screen?\"", false, "", "", 0},
		{"122", "\"headcount\"", false, "", "", 0},
		{"123", "CEO's name mentioned", false, "", "", 0},
		{"124", "\"TK\"", false, "", "", 0},
		{"125", "VP's name mentioned", false, "", "", 0},
	}

	return phrases
}

////////////////////////////////////////////////////////////////////////////////
// GAMES
////////////////////////////////////////////////////////////////////////////////

// NewGame will create a new game in the database and initialize it.
func (a *Agent) NewGame(name string, player Player) (Game, error) {

	phrases, err := a.GetPhrases()
	if err != nil {
		return Game{}, fmt.Errorf("failed to get phrases: %v", err)
	}

	g := NewGame(name, player, phrases)

	batch := a.client.Batch()
	a.log(fmt.Sprintf("Creating new game, id: %s", g.ID))

	gref := a.client.Collection("games").Doc(g.ID)
	batch.Set(gref, g)

	aref := a.client.Collection("games").Doc(g.ID).Collection("admins").Doc(player.Email)
	batch.Set(aref, player)

	pref := a.client.Collection("games").Doc(g.ID).Collection("players").Doc(player.Email)
	batch.Set(pref, player)

	a.log("Adding phrases to new game")
	for _, v := range g.Master.Records {
		ref := a.client.Collection("games").Doc(g.ID).Collection("records").Doc(v.Phrase.ID)
		batch.Set(ref, v)
	}

	m := Message{}
	m.SetText("Game has begun!")
	m.SetAudience("all")

	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	mref := a.client.Collection("games").Doc(g.ID).Collection("messages").Doc(timestamp)
	batch.Set(mref, m)

	_, err = batch.Commit(a.ctx)
	if err != nil {
		return g, fmt.Errorf("failed to add records to database: %v", err)
	}

	return g, nil
}

// GetGames finds a collection of all games.
func (a *Agent) GetGames(limit int, token time.Time) (Games, error) {
	g := []Game{}

	a.log("Getting Games")
	iter := a.client.Collection("games").
		Where("active", "==", true).Limit(limit).
		OrderBy("created", firestore.Desc).
		StartAfter(token).Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("Failed to iterate: %v", err)
		}
		game := Game{}
		doc.DataTo(&game)
		game.ID = doc.Ref.ID

		game, err = a.loadGameWithRecords(game)
		if err != nil {
			return g, fmt.Errorf("failed to load records for game: %v", err)
		}

		game, err = a.loadGameWithPlayers(game)
		if err != nil {
			return g, fmt.Errorf("failed to load players for game: %v", err)
		}

		game, err = a.loadGameWithAdmins(game)
		if err != nil {
			return g, fmt.Errorf("failed to load admins for game: %v", err)
		}

		game, err = a.loadGameWithBoards(game)
		if err != nil {
			return g, fmt.Errorf("failed to load boards for game: %v", err)
		}

		g = append(g, game)
	}

	return g, nil
}

// GetGame gets a given game from the database
func (a *Agent) GetGame(gid string) (Game, error) {
	g := Game{}
	g.Boards = map[string]Board{}

	a.log("Getting existing game")
	doc, err := a.client.Collection("games").Doc(gid).Get(a.ctx)
	if err != nil {
		return g, fmt.Errorf("failed to get game: %v", err)
	}

	doc.DataTo(&g)
	g.ID = gid
	g, err = a.loadGameWithRecords(g)
	if err != nil {
		return g, fmt.Errorf("failed to load records for game: %v", err)
	}

	g, err = a.loadGameWithPlayers(g)
	if err != nil {
		return g, fmt.Errorf("failed to load players for game: %v", err)
	}

	g, err = a.loadGameWithAdmins(g)
	if err != nil {
		return g, fmt.Errorf("failed to load admins for game: %v", err)
	}

	g, err = a.loadGameWithBoards(g)
	if err != nil {
		return g, fmt.Errorf("failed to load boards for game: %v", err)
	}

	return g, nil
}

func (a *Agent) loadGameWithRecords(game Game) (Game, error) {

	a.log("Loading records from game")
	iter := a.client.Collection("games").Doc(game.ID).Collection("records").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return game, fmt.Errorf("failed getting game records: %v", err)
		}
		r := Record{}
		doc.DataTo(&r)
		game.Master.Records = append(game.Master.Records, r)
	}

	return game, nil
}

func (a *Agent) loadGameWithPlayers(game Game) (Game, error) {

	a.log("Loading players from game")
	iter := a.client.Collection("games").Doc(game.ID).Collection("players").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return game, fmt.Errorf("failed getting game records: %v", err)
		}
		p := Player{}
		doc.DataTo(&p)
		game.Players.Add(p)
	}

	return game, nil
}

func (a *Agent) loadGameWithBoards(game Game) (Game, error) {

	a.log("Loading boards from game")
	iter := a.client.Collection("games").Doc(game.ID).Collection("boards").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return game, fmt.Errorf("failed getting game boards: %v", err)
		}
		b := InitBoard()
		doc.DataTo(&b)

		b, err = a.loadBoardWithPhrases(b)
		if err != nil {
			return game, fmt.Errorf("Failed to populare board: %v", err)
		}

		game.Boards[b.ID] = b
	}

	return game, nil
}

func (a *Agent) loadGameWithAdmins(game Game) (Game, error) {
	game.Boards = map[string]Board{}

	a.log("Loading players from game")
	iter := a.client.Collection("games").Doc(game.ID).Collection("admins").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return game, fmt.Errorf("failed getting game records: %v", err)
		}
		p := Player{}
		doc.DataTo(&p)
		game.Admins.Add(p)
	}

	return game, nil
}

// SaveGame records a game to firestore.
func (a *Agent) SaveGame(game Game) error {

	oldgame, err := a.GetGame(game.ID)
	if err != nil {
		return fmt.Errorf("error getting old data for game: %s", err)
	}

	a.log("Save game")
	batch := a.client.Batch()
	gref := a.client.Collection("games").Doc(game.ID)
	batch.Set(gref, game)

	// TODO: deal with other parts of game.

	for _, v := range oldgame.Players {
		ref := a.client.Collection("games").Doc(game.ID).Collection("players").Doc(v.Email)
		batch.Delete(ref)
	}

	for _, v := range oldgame.Admins {
		ref := a.client.Collection("games").Doc(game.ID).Collection("admins").Doc(v.Email)
		batch.Delete(ref)
	}

	for _, v := range game.Players {
		ref := a.client.Collection("games").Doc(game.ID).Collection("players").Doc(v.Email)
		batch.Set(ref, v)
	}

	for _, v := range game.Admins {
		ref := a.client.Collection("games").Doc(game.ID).Collection("admins").Doc(v.Email)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to save game to database: %v", err)
	}

	return nil
}

// UpdatePhrase updates a phrase on a particular game and all boards associated with it.
func (a *Agent) UpdatePhrase(game Game, phrase Phrase) error {
	b := game.Boards

	phraseMap := map[string]interface{}{"text": phrase.Text, "selected": false}
	recordMap := map[string]interface{}{"phrase": phraseMap, "players": Players{}}

	batch := a.client.Batch()
	recoref := a.client.Collection("games").Doc(game.ID).Collection("records").Doc(phrase.ID)
	batch.Set(recoref, recordMap, firestore.MergeAll)

	for _, v := range b {
		msg := fmt.Sprintf("Updating to phrase %s on board %s on game %s", phrase.ID, v.ID, game.ID)
		a.log(msg)
		ref := a.client.Collection("games").Doc(game.ID).Collection("boards").Doc(v.ID).Collection("phrases").Doc(phrase.ID)
		batch.Set(ref, phraseMap, firestore.MergeAll)
	}

	a.log("Committing Batch")
	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

// GetBoardsForGame gets all the boards for a give game.
func (a *Agent) GetBoardsForGame(game Game) ([]Board, error) {

	b := []Board{}

	a.log("Getting boards for game")
	iter := a.client.Collection("games").Doc(game.ID).Collection("boards").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return b, fmt.Errorf("Failed to iterate: %v", err)
		}

		board := Board{}
		doc.DataTo(&board)
		board.ID = doc.Ref.ID

		board, err = a.loadBoardWithPhrases(board)
		if err != nil {
			return b, fmt.Errorf("Failed to populare board: %v", err)
		}

		b = append(b, board)

	}

	return b, nil
}

// GetGamesForKey fetches the list of all games a user in currently in.
func (a *Agent) GetGamesForKey(email string) (Games, error) {

	g := []Game{}

	refs := []*firestore.DocumentRef{}
	a.log("Getting Games for player")
	iter := a.client.CollectionGroup("players").Where("email", "==", email).Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("Failed to iterate: %v", err)
		}

		refs = append(refs, doc.Ref.Parent.Parent)
	}

	a.log("Getting Games for player")
	snapshots, err := a.client.GetAll(a.ctx, refs)
	if err != nil {
		return g, fmt.Errorf("Failed to get games: %v", err)
	}

	for _, v := range snapshots {
		game := Game{}
		v.DataTo(&game)
		game.ID = v.Ref.ID

		if !game.Active {
			continue
		}

		game, err := a.loadGameWithAdmins(game)
		if err != nil {
			return g, fmt.Errorf("failed to get admins for game: %v", err)
		}

		g = append(g, game)
	}

	return g, nil
}

func (a *Agent) PurgeOldGames() error {
	g := []Game{}

	dateCutoff := time.Now().AddDate(0, 0, -30)

	msg := fmt.Sprintf("Getting Games before %s", dateCutoff.Format("2006-01-02"))

	a.log(msg)
	iter := a.client.Collection("games").Where("created", "<", dateCutoff).Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("Failed to iterate: %v", err)
		}
		game := Game{}
		doc.DataTo(&game)
		game.ID = doc.Ref.ID

		g = append(g, game)
	}

	for _, v := range g {
		msg := fmt.Sprintf("%s - %s \n", v.Name, v.Created.Format("2006-01-02"))
		a.log(msg)
		err := a.DeleteGame(v)
		if err != nil {
			return fmt.Errorf("failure deleting %s: %s", v.Name, err)
		}
	}

	return nil
}

// DeleteGame delete a specifc game from firestore
func (a *Agent) DeleteGame(game Game) error {

	refs := []*firestore.DocumentRef{}

	game, err := a.GetGame(game.ID)
	if err != nil {
		return fmt.Errorf("loading complete game data: %v", err)
	}

	batch := a.client.Batch()
	a.log("Deleting game")
	gref := a.client.Collection("games").Doc(game.ID)
	refs = append(refs, gref)

	for _, v := range game.Admins {
		a.log("Deleting admin: " + v.Email)
		rref := a.client.Collection("games").Doc(game.ID).Collection("admins").Doc(v.Email)
		refs = append(refs, rref)
	}

	for _, v := range game.Players {
		a.log("Deleting player: " + v.Email)
		rref := a.client.Collection("games").Doc(game.ID).Collection("players").Doc(v.Email)
		refs = append(refs, rref)
	}

	for _, v := range game.Master.Records {
		a.log("Deleting record: " + v.ID)
		rref := a.client.Collection("games").Doc(game.ID).Collection("records").Doc(v.ID)
		refs = append(refs, rref)
	}

	for _, v := range game.Boards {
		a.log("Deleting boards: " + v.ID)
		rref := a.client.Collection("games").Doc(game.ID).Collection("boards").Doc(v.ID)
		refs = append(refs, rref)

		for _, subv := range v.Phrases {
			a.log("Deleting phrases: " + subv.ID)
			pref := a.client.Collection("games").Doc(game.ID).Collection("boards").Doc(v.ID).Collection("phrases").Doc(subv.ID)
			refs = append(refs, pref)
		}
	}

	a.log("removing messages from board")
	ref := a.client.Collection("games").Doc(game.ID).Collection("messages")
	for {
		// Get a batch of documents
		iter := ref.Limit(500).Documents(a.ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to clean messages from firestore: %v", err)
			}

			a.log(fmt.Sprintf("removing message %s from board", doc.Ref.ID))
			refs = append(refs, doc.Ref)
			batch.Delete(doc.Ref)
		}

		limit := 500
		for i := 0; i < len(refs); i = i + limit {
			added := 0
			batch := a.client.Batch()

			for j := i; j < limit; j++ {
				if j >= len(refs)-i {
					break
				}
				batch.Delete(refs[j])
				added++
			}

			if added == 0 {
				break
			}

			if _, err := batch.Commit(a.ctx); err != nil {
				return fmt.Errorf("failed to clean messages from firestore: %v", err)
			}

		}

		if _, err := batch.Commit(a.ctx); err != nil {
			return fmt.Errorf("failed to clean messages from firestore: %v", err)
		}
		return nil

	}

}

////////////////////////////////////////////////////////////////////////////////
// MESSAGES
////////////////////////////////////////////////////////////////////////////////

// AddMessagesToGame broadcasts a message to the game players
func (a *Agent) AddMessagesToGame(game Game, messages []Message) error {

	batch := a.client.Batch()
	for _, v := range messages {
		a.log("Adding message to game")
		timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		v.ID = timestamp
		ref := a.client.Collection("games").Doc(game.ID).Collection("messages").Doc(timestamp)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to send messages : %v", err)
	}

	return nil
}

// AcknowledgeMessage marks the message as having been received.
func (a *Agent) AcknowledgeMessage(game Game, message Message) error {

	update := map[string]interface{}{"received": true}
	if _, err := a.client.Collection("games").Doc(game.ID).Collection("messages").Doc(message.ID).Set(ctx, update, firestore.MergeAll); err != nil {
		return fmt.Errorf("unable to acknowledge message: %s", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOARDS
////////////////////////////////////////////////////////////////////////////////

// GetBoardForPlayer returns the board for a given player
func (a *Agent) GetBoardForPlayer(gid string, p Player) (Board, error) {
	b := InitBoard()

	a.log("get board for player")
	iter := a.client.Collection("games").Doc(gid).Collection("boards").Where("player.email", "==", p.Email).Documents(a.ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return b, fmt.Errorf("failed to iterate over b from firestore: %v", err)
		}
		doc.DataTo(&b)
		b.ID = doc.Ref.ID
		break
	}
	if b.ID != "" {
		var err error
		b, err = a.loadBoardWithPhrases(b)
		if err != nil {
			return b, fmt.Errorf("failed to load phrases for board: %v", err)
		}
	}

	return b, nil

}

// GetBoard retrieves a specifc board from firestore
func (a *Agent) GetBoard(bid, gid string) (Board, error) {
	b := InitBoard()

	a.log("Getting board")
	doc, err := a.client.Collection("games").Doc(gid).Collection("boards").Doc(bid).Get(a.ctx)
	if err != nil {
		return b, fmt.Errorf("failed to get board: %v", err)
	}

	doc.DataTo(&b)
	b.ID = bid
	b, err = a.loadBoardWithPhrases(b)
	if err != nil {
		return b, fmt.Errorf("failed to load phrases for board: %v", err)
	}

	return b, nil
}

func (a *Agent) loadBoardWithPhrases(board Board) (Board, error) {

	a.log("Adding phrases to existing board")
	iter := a.client.Collection("games").Doc(board.Game).Collection("boards").Doc(board.ID).Collection("phrases").OrderBy("displayorder", firestore.Asc).Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return board, fmt.Errorf("failed getting board records: %v", err)
		}
		p := Phrase{}
		doc.DataTo(&p)
		board.Phrases[p.ID] = p
	}

	return board, nil
}

// DeleteBoard delete a specifc board from firestore
func (a *Agent) DeleteBoard(board Board, game Game) error {
	batch := a.client.Batch()
	a.log("Deleting board")
	bref := a.client.Collection("games").Doc(game.ID).Collection("boards").Doc(board.ID)
	batch.Delete(bref)

	for _, v := range game.Master.Records {
		id := v.ID
		if id == "" {
			id = v.Phrase.ID
		}
		rref := a.client.Collection("games").Doc(game.ID).Collection("records").Doc(id)
		batch.Set(rref, v)
	}

	a.log("removing phrases from board")
	ref := a.client.Collection("games").Doc(game.ID).Collection("boards").Doc(board.ID).Collection("phrases")
	for {
		// Get a batch of documents
		iter := ref.Limit(100).Documents(a.ctx)

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to clean phrases from firestore: %v", err)
			}

			a.log(fmt.Sprintf("removing phrase %s from board", doc.Ref.ID))
			batch.Delete(doc.Ref)
		}
		if _, err := batch.Commit(a.ctx); err != nil {
			return fmt.Errorf("failed to clean messages from firestore: %v", err)
		}
		return nil

	}

}

// SaveBoard persists a board to firestore
func (a *Agent) SaveBoard(board Board) (Board, error) {

	a.log("Starting batch operation")
	batch := a.client.Batch()
	bref := a.client.Collection("games").Doc(board.Game).Collection("boards").Doc(board.ID)
	batch.Set(bref, board)

	pref := a.client.Collection("games").Doc(board.Game).Collection("players").Doc(board.Player.Email)
	batch.Set(pref, board.Player)

	for _, v := range board.Phrases {
		ref := a.client.Collection("games").Doc(board.Game).Collection("boards").Doc(board.ID).Collection("phrases").Doc(v.ID)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return board, fmt.Errorf("failed to add records to database: %v", err)
	}

	return board, nil

}

// SelectPhrase records clicks on the board and the game
func (a *Agent) SelectPhrase(board Board, phrase Phrase, record Record) error {

	a.log("Starting batch operation")
	batch := a.client.Batch()

	a.log("Updating phrase on board")
	bref := a.client.Collection("games").Doc(board.Game).Collection("boards").Doc(board.ID).Collection("phrases").Doc(phrase.ID)
	batch.Set(bref, phrase)

	a.log("Updating game record")
	gref := a.client.Collection("games").Doc(board.Game).Collection("records").Doc(record.Phrase.ID)
	batch.Set(gref, record)

	a.log("Updating board to bingo")
	bingoref := a.client.Collection("games").Doc(board.Game).Collection("boards").Doc(board.ID)
	update := map[string]interface{}{"bingodeclared": board.BingoDeclared}
	batch.Set(bingoref, update, firestore.MergeAll)

	a.log("Committing Batch")
	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

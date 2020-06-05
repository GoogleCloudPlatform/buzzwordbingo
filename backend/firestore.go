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
func (a *Agent) AddAdmin(p Player) error {
	if _, err := a.client.Collection("admins").Doc(p.Email).Set(ctx, p); err != nil {
		return fmt.Errorf("unable to add admin: %s", err)
	}
	return nil
}

// DeleteAdmin Deletes an admin to the over all system
func (a *Agent) DeleteAdmin(p Player) error {
	if _, err := a.client.Collection("admins").Doc(p.Email).Delete(ctx); err != nil {
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
		p = append(p, player)
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

// UpdateMasterPhrase updates a phrase in the master collection of phrases
func (a *Agent) UpdateMasterPhrase(p Phrase) error {

	if _, err := a.client.Collection("phrases").Doc(p.ID).Set(a.ctx, p); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GAMES
////////////////////////////////////////////////////////////////////////////////

// NewGame will create a new game in the database and initialize it.
func (a *Agent) NewGame(name string, p Player) (Game, error) {
	g := Game{}

	phrases, err := a.GetPhrases()
	if err != nil {
		return g, fmt.Errorf("failed to get phrases: %v", err)
	}

	g.ID = uniqueID()
	g.Admins = append(g.Admins, p)
	g.Name = name
	g.Active = true
	g.Created = time.Now()
	g.Master.Load(phrases)

	batch := a.client.Batch()
	a.log(fmt.Sprintf("Creating new game, id: %s", g.ID))

	gref := a.client.Collection("games").Doc(g.ID)
	batch.Set(gref, g)

	a.log("Adding phrases to new game")
	for _, v := range g.Master.Records {
		ref := a.client.Collection("games").Doc(g.ID).Collection("records").Doc(v.Phrase.ID)
		batch.Set(ref, v)
	}

	aref := a.client.Collection("games").Doc(g.ID).Collection("admins").Doc(p.Email)
	batch.Set(aref, p)

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
func (a *Agent) GetGames() (Games, error) {
	g := []Game{}

	a.log("Getting Games for player")
	iter := a.client.Collection("games").Where("active", "==", true).Documents(a.ctx)
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
func (a *Agent) GetGame(id string) (Game, error) {
	g := Game{}
	g.Boards = map[string]Board{}

	a.log("Getting existing game")
	doc, err := a.client.Collection("games").Doc(id).Get(a.ctx)
	if err != nil {
		return g, fmt.Errorf("failed to get game: %v", err)
	}

	doc.DataTo(&g)
	g.ID = id
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

func (a *Agent) loadGameWithRecords(g Game) (Game, error) {

	a.log("Loading records from game")
	iter := a.client.Collection("games").Doc(g.ID).Collection("records").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed getting game records: %v", err)
		}
		r := Record{}
		doc.DataTo(&r)
		g.Master.Records = append(g.Master.Records, r)
	}

	return g, nil
}

func (a *Agent) loadGameWithPlayers(g Game) (Game, error) {

	a.log("Loading players from game")
	iter := a.client.Collection("games").Doc(g.ID).Collection("players").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed getting game records: %v", err)
		}
		p := Player{}
		doc.DataTo(&p)
		g.Players = append(g.Players, p)
	}

	return g, nil
}

func (a *Agent) loadGameWithBoards(g Game) (Game, error) {

	a.log("Loading boards from game")
	iter := a.client.Collection("games").Doc(g.ID).Collection("boards").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed getting game boards: %v", err)
		}
		b := Board{}
		doc.DataTo(&b)

		b, err = a.loadBoardWithPhrases(b)
		if err != nil {
			return g, fmt.Errorf("Failed to populare board: %v", err)
		}

		g.Boards[b.ID] = b
	}

	return g, nil
}

func (a *Agent) loadGameWithAdmins(g Game) (Game, error) {
	g.Boards = map[string]Board{}

	a.log("Loading players from game")
	iter := a.client.Collection("games").Doc(g.ID).Collection("admins").Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return g, fmt.Errorf("failed getting game records: %v", err)
		}
		p := Player{}
		doc.DataTo(&p)
		g.Admins = append(g.Admins, p)
	}

	return g, nil
}

// SaveGame records a game to firestore.
func (a *Agent) SaveGame(g Game) error {

	oldgame, err := a.GetGame(g.ID)
	if err != nil {
		return fmt.Errorf("error getting old data for game: %s", err)
	}

	a.log("Save game")
	batch := a.client.Batch()
	gref := a.client.Collection("games").Doc(g.ID)
	batch.Set(gref, g)

	// TODO: deal with other parts of game.

	for _, v := range oldgame.Players {
		ref := a.client.Collection("games").Doc(g.ID).Collection("players").Doc(v.Email)
		batch.Delete(ref)
	}

	for _, v := range oldgame.Admins {
		ref := a.client.Collection("games").Doc(g.ID).Collection("admins").Doc(v.Email)
		batch.Delete(ref)
	}

	for _, v := range g.Players {
		ref := a.client.Collection("games").Doc(g.ID).Collection("players").Doc(v.Email)
		batch.Set(ref, v)
	}

	for _, v := range g.Admins {
		ref := a.client.Collection("games").Doc(g.ID).Collection("admins").Doc(v.Email)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to save game to database: %v", err)
	}

	return nil
}

// UpdatePhrase updates a phrase on a particular game and all boards associated with it.
func (a *Agent) UpdatePhrase(g Game, p Phrase) error {
	b := g.Boards

	phraseMap := map[string]interface{}{"text": p.Text, "selected": false}

	batch := a.client.Batch()
	record := Record{}
	record.Phrase = p
	recoref := a.client.Collection("games").Doc(g.ID).Collection("records").Doc(p.ID)
	batch.Set(recoref, record)

	for _, v := range b {
		msg := fmt.Sprintf("Updating to phrase %s on board %s on game %s", p.ID, v.ID, g.ID)
		a.log(msg)
		ref := a.client.Collection("games").Doc(g.ID).Collection("boards").Doc(v.ID).Collection("phrases").Doc(p.ID)
		batch.Set(ref, phraseMap, firestore.MergeAll)
	}

	a.log("Committing Batch")
	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

// GetBoardsForGame gets all the boards for a give game.
func (a *Agent) GetBoardsForGame(g Game) ([]Board, error) {

	b := []Board{}

	a.log("Getting boards for game")
	iter := a.client.Collection("games").Doc(g.ID).Collection("boards").Documents(a.ctx)
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

////////////////////////////////////////////////////////////////////////////////
// MESSAGES
////////////////////////////////////////////////////////////////////////////////

// AddMessagesToGame broadcasts a message to the game players
func (a *Agent) AddMessagesToGame(g Game, m []Message) error {

	batch := a.client.Batch()
	for _, v := range m {
		a.log("Adding message to game")
		timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
		v.ID = timestamp
		ref := a.client.Collection("games").Doc(g.ID).Collection("messages").Doc(timestamp)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to send messages : %v", err)
	}

	return nil
}

// AcknowledgeMessage marks the message as having been received.
func (a *Agent) AcknowledgeMessage(g Game, m Message) error {

	update := map[string]interface{}{"received": true}
	if _, err := a.client.Collection("games").Doc(g.ID).Collection("messages").Doc(m.ID).Set(ctx, update, firestore.MergeAll); err != nil {
		return fmt.Errorf("unable to acknowledge message: %s", err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BOARDS
////////////////////////////////////////////////////////////////////////////////

// GetBoardForPlayer returns the board for a given player
func (a *Agent) GetBoardForPlayer(id string, p Player) (Board, error) {
	b := Board{}

	a.log("get board for player")
	iter := a.client.Collection("games").Doc(id).Collection("boards").Where("game", "==", id).Where("player.email", "==", p.Email).Documents(a.ctx)

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
	b := Board{}

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

func (a *Agent) loadBoardWithPhrases(b Board) (Board, error) {

	a.log("Adding phrases to existing board")
	iter := a.client.Collection("games").Doc(b.Game).Collection("boards").Doc(b.ID).Collection("phrases").OrderBy("displayorder", firestore.Asc).Documents(a.ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return b, fmt.Errorf("failed getting board records: %v", err)
		}
		p := Phrase{}
		doc.DataTo(&p)
		b.Phrases = append(b.Phrases, p)
	}

	return b, nil
}

// DeleteBoard delete a specifc board from firestore
func (a *Agent) DeleteBoard(bid, gid string) error {
	batch := a.client.Batch()
	a.log("Deleting board")
	bref := a.client.Collection("games").Doc(gid).Collection("boards").Doc(bid)
	batch.Delete(bref)
	a.log("removing phrases from board")
	ref := a.client.Collection("games").Doc(gid).Collection("boards").Doc(bid).Collection("phrases")
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
func (a *Agent) SaveBoard(b Board) (Board, error) {

	if b.ID == "" {
		b.ID = uniqueID()
	}

	a.log("Starting batch operation")
	batch := a.client.Batch()
	bref := a.client.Collection("games").Doc(b.Game).Collection("boards").Doc(b.ID)
	batch.Set(bref, b)

	pref := a.client.Collection("games").Doc(b.Game).Collection("players").Doc(b.Player.Email)
	batch.Set(pref, b.Player)

	for _, v := range b.Phrases {
		ref := a.client.Collection("games").Doc(b.Game).Collection("boards").Doc(b.ID).Collection("phrases").Doc(v.ID)
		batch.Set(ref, v)
	}

	if _, err := batch.Commit(a.ctx); err != nil {
		return b, fmt.Errorf("failed to add records to database: %v", err)
	}

	return b, nil

}

// SelectPhrase records clicks on the board and the game
func (a *Agent) SelectPhrase(b Board, p Phrase, r Record) error {

	a.log("Starting batch operation")
	batch := a.client.Batch()

	a.log("Updating phrase on board")
	bref := a.client.Collection("games").Doc(b.Game).Collection("boards").Doc(b.ID).Collection("phrases").Doc(p.ID)
	batch.Set(bref, p)

	a.log("Updating game record")
	gref := a.client.Collection("games").Doc(b.Game).Collection("records").Doc(r.Phrase.ID)
	batch.Set(gref, r)

	a.log("Updating board to bingo")
	bingoref := a.client.Collection("games").Doc(b.Game).Collection("boards").Doc(b.ID)
	update := map[string]interface{}{"bingodeclared": b.BingoDeclared}
	batch.Set(bingoref, update, firestore.MergeAll)

	a.log("Committing Batch")
	if _, err := batch.Commit(a.ctx); err != nil {
		return fmt.Errorf("failed to update phrase: %v", err)
	}

	return nil
}

Frontend should get status info directly fron firestore
But all logic should be handled by the backend. 
    Creating collective game
    Creating Randomized board
    scoring each randomized board
    declaring bingo



Database
Entities
    Phrase
        PhraseID
        Text
        Checked (optional)
        Position (optional)
    Game
        Status
            []Phrase
            Bingo?
        []Board
    Board
        Player
        Bingo?
        []Phrase
    Player
        ID
        Name
        []Message
    Message
        Text

Backend
Create Game
    Creates Game
    Populates Master Board 

CreatePlayer
    Creates a player
    Creates a board for the player
    Adds player to board
    Adds board to game
    Returns a board id
     
Click Phrase
    Captures an individual click on one board
        Marks the phrase clicked on individual board
        Marks the phrase clicked in the game status

CalculateBingo
    Each board shoudl be able to calculate bingo on itself
    Then the game should be able to calculate if any boards have bingo

Declare bingo
    Send message to player that they have bingo
    Send message to others that someone has bingo



Frontend Flow
User arrives
Login
    Do they exist in system?
        Yes 
            Are they an admin
                Yes - 
                    Give them the master view
                No - Give them a board
        No - Create user, give them a board 
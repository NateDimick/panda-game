package engine

type ServerEventType string

const (
	LobbyUpdate  ServerEventType = "LobbyUpdate"
	GameStart    ServerEventType = "GameStart"
	GameUpdate   ServerEventType = "GameUpdate"
	GameOver     ServerEventType = "GameOver"
	ActionPrompt ServerEventType = "ActionPrompt"
	Goodbye      ServerEventType = "Goodbye" // the server has forced the connection closed
	Warning      ServerEventType = "Warning" // the last message received was bad. Warn the client to do better
)

type ClientEventType string

const (
	JoinGame        ClientEventType = "JoinGame"
	LeaveGame       ClientEventType = "LeaveGame"
	GameChat        ClientEventType = "GameChat"
	TakeAction      ClientEventType = "TakeAction"
	Reprompt        ClientEventType = "RePrompt"
	CreateGame      ClientEventType = "CreateGame"
	StartGame       ClientEventType = "StartGame"
	ChangeSettings  ClientEventType = "ChangeSettings"
	Matchmake       ClientEventType = "Matchmake"
	CancelMatchmake ClientEventType = "CancelMatchmake"
)

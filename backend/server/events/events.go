package events

import socketio "github.com/googollee/go-socket.io"

func OnConnect(s socketio.Conn) error {
	// a new user connects
	return nil
}

func OnDisconnect(s socketio.Conn, reason string) {
	//
}

func OnError(s socketio.Conn, e error) {
	//
}

func OnSearchForGame(s socketio.Conn, msg string) {
	//
}

func OnCancelSearchForGame(s socketio.Conn, msg string) {
	//
}

func OnCreateRoom(s socketio.Conn, msg string) {
	//
}

func OnChatInRoom(s socketio.Conn, msg string) {
	//
}

func OnRollWeatherDie(s socketio.Conn, msg string) {
	//
}

func OnMakeWeatherChoice(s socketio.Conn, msg string) {
	//
}

func OnTakeTurnAction(s socketio.Conn, msg string) {
	//
}

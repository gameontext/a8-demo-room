package main

import (
	"encoding/json"
	"net/http"

	"fmt"

	"os"

	"bytes"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/gameontext/a8-room/pkg/gameon"
	"github.com/gorilla/websocket"
)

var (
	SupportedVersions = []int{1}
)

type mediator struct {
	room     *room
	roomID   string
	sessions *SessionManager
}

func newMediator() *mediator {
	m := &mediator{
		room:     newRoom(),
		roomID:   os.Getenv("ROOM_ID"),
		sessions: newSessions(),
	}

	return m
}

func (m *mediator) handleHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Incoming HTTP request from %s", r.RemoteAddr)

	var upgrader websocket.Upgrader
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.WithError(err).Errorf("Error upgrading HTTP to websocket connection")
		return
	}

	logrus.Debugf("Websocket connection established with %s", conn.RemoteAddr().String())
	m.handleWebsocket(conn)
}

func (m *mediator) handleWebsocket(conn *websocket.Conn) {
	session := m.sessions.NewSession(conn)

	m.ack(session)
	go m.handleMessages(session)

	select {
	case <-session.Closed():
		conn.Close()
	}
}

func (m *mediator) handleMessages(session *Session) {
	// The loop runs forever, and is terminated only when an error occurs.
	// In such a case, attempt to close the session (in case not closed already).
	defer session.Close()

	for {
		_, bytes, err := session.Conn.ReadMessage()
		if err != nil {
			logrus.WithError(err).Errorf("Error reading websocket message")
			return
		}

		msg, err := parseMessage(bytes)
		if err != nil {
			logrus.WithError(err).Errorf("Error parsing websocket message")
			return
		}

		logrus.WithFields(messageToFields(msg)).Debugf("Websocket message received")

		// Validate the message recipient is our own room ID
		if m.roomID != "" && msg.Recipient != m.roomID {
			logrus.WithError(fmt.Errorf("recipient (%s) doesn't match expected room id (%s)", msg.Recipient, m.roomID)).
				Errorf("Invalid message received")
			return
		}

		var payload interface{}
		switch msg.Direction {
		case "roomHello":
			payload = &gameon.Hello{}
		case "roomGoodbye":
			payload = &gameon.Goodbye{}
		case "room":
			payload = &gameon.RoomCommand{}
		default:
			logrus.WithError(fmt.Errorf("unrecognized message direction: %s", msg.Direction)).
				Errorf("Invalid message received")
			return
		}

		err = json.Unmarshal(msg.Payload, payload)
		if err != nil {
			logrus.WithError(err).Errorf("Error unmarshaling message payload")
			return
		}

		switch payload := payload.(type) {
		case *gameon.Hello:
			m.handleHello(payload, session)
		case *gameon.Goodbye:
			m.handleGoodbye(payload, session)
		case *gameon.RoomCommand:
			m.handleRoomCommand(payload, session)
		default:
			logrus.WithError(fmt.Errorf("unrecognized payload type: %T", payload))
		}
	}
}

func (m *mediator) ack(session *Session) {
	logrus.Debugf("Sending ack for websocket connection with remote address %s", session.Conn.RemoteAddr().String())

	ack := gameon.Ack{
		Version: SupportedVersions,
	}
	ackBytes, _ := json.Marshal(ack)

	msg := &gameon.Message{
		Direction: "ack",
		Payload:   ackBytes,
	}

	sendMessage(msg, session)
}

func (m *mediator) handleHello(hello *gameon.Hello, session *Session) {
	session.SetUserID(hello.UserID)

	resp, err := m.room.Hello(hello)
	if err != nil {
		logrus.WithError(err).Errorf("Error executing 'hello' with room service")
		return
	}

	m.handleResponse(resp)
}

func (m *mediator) handleGoodbye(goodbye *gameon.Goodbye, session *Session) {
	defer session.Close()

	resp, err := m.room.Goodbye(goodbye)
	if err != nil {
		logrus.WithError(err).Errorf("Error executing 'goodbye' with room service")
		return
	}

	m.handleResponse(resp)
}

func (m *mediator) handleRoomCommand(command *gameon.RoomCommand, session *Session) {
	resp, err := m.room.Command(command)
	if err != nil {
		logrus.WithError(err).Errorf("Error executing command with room service")
		return
	}

	m.handleResponse(resp)
}

func (m *mediator) handleResponse(resp *gameon.MessageCollection) {
	switch len(resp.Messages) {
	case 0:
		logrus.Debugf("Response contains no messages")
	case 1:
		logrus.Debugf("Dispatching 1 response message")
	default:
		logrus.Debugf("Dispatching %d response message", len(resp.Messages))
	}

	for _, msg := range resp.Messages {
		if msg.Recipient == "*" {
			sendMessage(&msg, m.sessions.GetUserSessions()...)
		} else {
			session := m.sessions.GetUserSession(msg.Recipient)
			if session != nil {
				sendMessage(&msg, session)
			}
		}
	}
}

func sendMessage(msg *gameon.Message, sessions ...*Session) {
	logrus.WithFields(messageToFields(msg)).Debugf("Sending message")

	bytes, err := formatMessage(msg)
	if err != nil {
		logrus.WithError(err).Errorf("Error formatting message")
		return
	}

	for _, session := range sessions {
		err := session.Conn.WriteMessage(websocket.TextMessage, bytes)
		if err != nil {
			logrus.WithError(err).Errorf("Error broadcasting message")
			session.Close()
		}
	}
}

func formatMessage(msg *gameon.Message) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(msg.Direction)
	buf.WriteRune(',')

	if msg.Recipient != "" {
		buf.WriteString(msg.Recipient)
		buf.WriteRune(',')
	}

	buf.Write(msg.Payload)
	return buf.Bytes(), nil
}

func parseMessage(data []byte) (*gameon.Message, error) {
	parts := strings.SplitN(string(data), ",", 3)

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid websocket message format: %s", string(data))
	}

	msg := new(gameon.Message)
	msg.Direction = parts[0]

	if strings.HasPrefix(parts[1], "{") {
		// case 1: <direction>,{...}
		msg.Payload = data[len(parts[0])+1:]
	} else {
		// case 2: <direction>,<recipient>,{...}
		msg.Recipient = parts[1]
		msg.Payload = data[len(parts[0])+len(parts[1])+2:]
	}

	return msg, nil
}

func messageToFields(msg *gameon.Message) logrus.Fields {
	return logrus.Fields{
		"direction": msg.Direction,
		"recipient": msg.Recipient,
		"payload":   string(msg.Payload),
	}
}

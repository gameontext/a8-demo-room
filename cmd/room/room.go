package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gameontext/a8-room/pkg/gameon"
)

var exits = map[string]string{
	"N": "An old wooden door with a large arrow carved on its center",
	"S": "A heavy metal door with signs of rust",
	"W": "A gray, plain looking door",
	"E": "A door surrounded by a mysterious glow along it edges",
}

type room struct {
	profanityChecker ProfanityChecker
}

func newRoom() *room {
	return &room{
		profanityChecker: newProfanityChecker(),
	}
}

func (r *room) hello(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var hello gameon.Hello
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&hello)

	if err != nil || hello.UserID == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	location := gameon.Message{
		Direction: "player",
		Recipient: hello.UserID,
		Payload: jsonMarshal(gameon.Location{
			Type:        "location",
			Name:        "Chatter",
			FullName:    "A chat room",
			Description: "a darkly lit room, there are people here, some are walking around, some are standing in groups",
			Exits:       exits,
			Commands:    map[string]string{},
			Inventory:   []string{},
		}),
	}

	welcome := gameon.Message{
		Direction: "player",
		Recipient: "*",
		Payload: jsonMarshal(gameon.Event{
			Type: "event",
			Content: map[string]string{
				hello.UserID: "Welcome!",
				"*":          fmt.Sprintf("%s has just entered the room", hello.Username),
			},
		}),
	}

	writeResponseMessages(resp, location, welcome)
}

func (r *room) goodbye(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var goodbye gameon.Goodbye
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&goodbye)

	if err != nil || goodbye.UserID == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	farewell := gameon.Message{
		Direction: "player",
		Recipient: "*",
		Payload: jsonMarshal(gameon.Event{
			Type: "event",
			Content: map[string]string{
				goodbye.UserID: "Farewell!",
				"*":            fmt.Sprintf("%s has left the room", goodbye.Username),
			},
		}),
	}

	writeResponseMessages(resp, farewell)
}

func (r *room) room(resp http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var command gameon.RoomCommand
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&command)
	if err != nil || command.UserID == "" || command.Content == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	if strings.HasPrefix(command.Content, "/") {
		// slash command
		r.handleSlash(command, resp)
	} else {
		// chat command
		r.handleChat(command, resp)
	}
}

func (r *room) handleSlash(command gameon.RoomCommand, resp http.ResponseWriter) {
	words := strings.Fields(command.Content)
	commandName := strings.ToLower(words[0])

	var eventContent string
	switch commandName {
	case "/go":
		if len(words) < 2 {
			eventContent = "Go where?"
			break
		}

		exitID := strings.ToUpper(words[1])
		if _, ok := exits[exitID]; !ok {
			eventContent = "You probably don't wanna go there..."
			break
		}

		location := gameon.Message{
			Direction: "playerLocation",
			Recipient: command.UserID,
			Payload: jsonMarshal(gameon.PlayerLocation{
				Type:    "exit",
				Content: "You frantically run towards the exit",
				ExitID:  exitID,
			}),
		}
		writeResponseMessages(resp, location)
		return

	case "/examine":
		eventContent = "Shouldn't you be mingling?"
	case "/inventory":
		eventContent = "There is nothing here"
	case "/look":
		eventContent = "It's just a room"
	default:
		eventContent = fmt.Sprintf("Don't know how to %s", commandName[1:])
	}

	event := gameon.Message{
		Direction: "player",
		Recipient: command.UserID,
		Payload: jsonMarshal(gameon.Event{
			Type: "event",
			Content: map[string]string{
				command.UserID: eventContent,
			},
		}),
	}
	writeResponseMessages(resp, event)
}

func (r *room) handleChat(command gameon.RoomCommand, resp http.ResponseWriter) {
	var msg gameon.Message

	dirty := r.profanityChecker.Check(command.Content)
	if dirty {
		msg = gameon.Message{
			Direction: "player",
			Recipient: command.UserID,
			Payload: jsonMarshal(gameon.Event{
				Type: "event",
				Content: map[string]string{
					command.UserID: "Pardon your french!",
				},
			}),
		}
	} else {
		msg = gameon.Message{
			Direction: "player",
			Recipient: "*",
			Payload: jsonMarshal(gameon.Chat{
				Type:     "chat",
				Username: command.Username,
				Content:  command.Content,
			}),
		}
	}

	writeResponseMessages(resp, msg)
}

func writeResponseMessages(resp http.ResponseWriter, messages ...gameon.Message) {
	bytes := jsonMarshal(gameon.MessageCollection{
		Messages: messages,
	})

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	resp.Write(bytes)
}

func jsonMarshal(obj interface{}) []byte {
	bytes, _ := json.Marshal(obj)
	return bytes
}

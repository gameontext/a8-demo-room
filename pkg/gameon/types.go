package gameon

import "encoding/json"

// MessageCollection is a collection of GameOn! messages
type MessageCollection struct {
	Messages []Message `json:"messages"`
}

// Message is a generic GameOn! message, holding a direction (player, room, ...),
// a recipient (playerID, roomID, *, ...), and a message-type specific payload.
type Message struct {
	Direction string          `json:"direction,omitempty"`
	Recipient string          `json:"recipient,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// UserInfo holds the ID and username of a GameOn! client.
// It is not a complete GameOn! message on its own, but rather meant to be embedded within different message types.
type UserInfo struct {
	Username string `json:"username,omitempty"`
	UserID   string `json:"userId,omitempty"`
}

// RoomCommand is the message payload provided for a [client --> mediator --> room] room chat/slash command message.
type RoomCommand struct {
	UserInfo
	Content string `json:"content,omitempty"`
}

// Hello is the message payload provided for a [mediator --> room] hello message.
type Hello struct {
	UserInfo
	Version  int  `json:"version,omitempty"`
	Recovery bool `json:"recovery,omitempty"`
}

// Goodbye is the message payload provided for a [mediator --> room] goodbye message.
type Goodbye struct {
	UserInfo
}

// Ack is the message payload provided for a [room --> mediator] ack message.
type Ack struct {
	Version []int `json:"version,omitempty"`
}

// Location is the message payload provided for a [room --> mediator --> client] location message.
type Location struct {
	Type        string            `json:"type,omitempty"`
	Name        string            `json:"name,omitempty"`
	FullName    string            `json:"fullName,omitempty"`
	Description string            `json:"description,omitempty"`
	Exits       map[string]string `json:"exits,omitempty"`
	Commands    map[string]string `json:"commands,omitempty"`
	Inventory   []string          `json:"roomInventory,omitempty"`
}

// PlayerLocation is the message payload provided for a [room --> mediator --> client] player-location message.
type PlayerLocation struct {
	Type    string `json:"type,omitempty"`
	Content string `json:"content,omitempty"`
	ExitID  string `json:"exitId,omitempty"`
	Exit    string `json:"exit,omitempty"`
}

// Chat is the message payload provided for a [room --> mediator --> client] chat message.
type Chat struct {
	Type     string `json:"type,omitempty"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content,omitempty"`
	Bookmark string `json:"bookmark,omitempty"`
}

// Event is the message payload provided for a [room --> mediator --> client] event message.
type Event struct {
	Type     string            `json:"type,omitempty"`
	Content  map[string]string `json:"content,omitempty"`
	Bookmark string            `json:"bookmark,omitempty"`
}

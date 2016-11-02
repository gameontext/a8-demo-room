package gameon

// Headers passed with API calls in between microservices implementing the chatter room.
const (
	// UserIDHeader carries the Game On user ID.
	UserIDHeader = "X-Game-On-UserID"

	// UsernameHeader carries the Game On user name.
	UsernameHeader = "X-Game-On-Username"
)

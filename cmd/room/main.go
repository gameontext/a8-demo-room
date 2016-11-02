package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.Infof("Starting room service")

	room := newRoom()

	http.HandleFunc("/hello", room.hello)
	http.HandleFunc("/goodbye", room.goodbye)
	http.HandleFunc("/room", room.room)

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		logrus.WithError(err).Fatalf("Error running main")
	}
}

package main

import (
	"net/http"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("Starting mediator service")

	m := newMediator()

	http.HandleFunc("/", m.handleHTTP)

	err := http.ListenAndServe(":3000", nil)
	if err != nil {
		logrus.WithError(err).Fatalf("Error running main")
	}
}

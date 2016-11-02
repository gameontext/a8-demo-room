package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/gameontext/a8-room/pkg/gameon"
)

type room struct {
	httpClient *http.Client
	serverURL  string
}

func newRoom() *room {
	serverURL := os.Getenv("ROOM_SERVICE_URL")
	if serverURL == "" {
		serverURL = "http://localhost:6379/room"
	}

	return &room{
		httpClient: &http.Client{Timeout: 5 * time.Second},
		serverURL:  serverURL,
	}
}

func (r *room) Hello(hello *gameon.Hello) (*gameon.MessageCollection, error) {
	return r.doRequest("/hello", hello.UserInfo, hello)
}

func (r *room) Goodbye(goodbye *gameon.Goodbye) (*gameon.MessageCollection, error) {
	return r.doRequest("/goodbye", goodbye.UserInfo, goodbye)
}

func (r *room) Command(command *gameon.RoomCommand) (*gameon.MessageCollection, error) {
	return r.doRequest("/room", command.UserInfo, command)
}

func (r *room) doRequest(path string, userInfo gameon.UserInfo, body interface{}) (*gameon.MessageCollection, error) {
	url := r.serverURL + path

	reqBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	reqBuf := bytes.NewBuffer(reqBytes)

	req, err := http.NewRequest("POST", url, reqBuf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(gameon.UserIDHeader, userInfo.UserID)
	req.Header.Set(gameon.UsernameHeader, userInfo.Username)

	logrus.Debugf("Executing HTTP request: %s %s (%d bytes)", req.Method, req.RequestURI, req.ContentLength)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	logrus.Debugf("Received HTTP response: %d %s (%d bytes)", resp.StatusCode, resp.Status, resp.ContentLength)

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msgs gameon.MessageCollection
	err = json.Unmarshal(respBytes, &msgs)
	if err != nil {
		return nil, err
	}

	return &msgs, nil
}

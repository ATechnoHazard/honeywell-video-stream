package utils

import (
	"errors"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
)

type WebsocketResponse struct {
	StreamUrl string `json:"streamURL"`
}

func MakeWebsocket() *websocket.Conn {
	var err error
	var conn *websocket.Conn
	err = errors.New("start the socket")
	for err != nil {
		u := url.URL{Scheme: "wss", Host: "alarmcomserver.itst.mymaxprocloud.com", Path: "/accept"}
		log.SetFormatter(&log.JSONFormatter{})
		conn, _, err = websocket.DefaultDialer.Dial(u.String(), http.Header{
			"Sec-Websocket-Protocol": []string{"wamp.2.json"},
			"Origin":                 []string{"https://itst.mymaxprocloud.com"},
		})
	}
	log.Println("Websocket connection established")

	return conn
}

func MakeVidWebSocket(url string) *websocket.Conn {
	log.SetFormatter(&log.JSONFormatter{})
	conn, _, err := websocket.DefaultDialer.Dial(url, http.Header{
		"Sec-Websocket-Protocol": []string{"lws-video"},
		"Origin":                 []string{"https://itst.mymaxprocloud.com"},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Println("Video websocket connection established")
	return conn
}

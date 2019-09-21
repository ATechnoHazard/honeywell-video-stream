package utils

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
)

func MakeWebsocket() *websocket.Conn {
	u := url.URL{Scheme: "wss", Host: "alarmcomserver.ispperf.mymaxprocloud.com", Path: "/accept"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Sec-Websocket-Protocol": []string{"wamp.2.json"},
		"Origin":                 []string{"https://ispperf.mymaxprocloud.com"},
	})
	if err != nil {
		log.Panic(err)
	}

	log.Println("Websocket connection established")
	return conn

}

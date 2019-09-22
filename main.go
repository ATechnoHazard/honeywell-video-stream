package main

import (
	"encoding/json"
	"fmt"
	"github.com/ATechnoHazard/honeywell-video-stream/utils"
	"github.com/gorilla/websocket"
	"golash"
	"io"
	"log"
	"strings"
	"sync"
)

type SocketResponse struct {
	PayloadString string `json:"PayloadString"`
}

type PayloadData struct {
	StatusCode string              `json:"statusCode"`
	Id         string              `json:"id"`
	Extension  []map[string]string `json:"extension"`
}

func main() {
	cred := utils.GetCreds()

	for _, creds := range cred {
		go func() {
			body := utils.User{Model: creds}
			conn := utils.MakeWebsocket()

			authUser := utils.MakeLoginReq(body) // Initial login request
			authUser.AcceptConn = conn

			token := utils.GetReqVerToken(authUser) // Get session verification token

			cameraList := utils.GetCameraList(authUser, token) // Get list of cameras

			authToken := utils.GetAuthToken(authUser, token) // Get auth token in local storage
			SubscribeSocket(conn, creds, authToken)          // subscribe for further events to socket

			streamUrls := getStreamUrl(authUser, cameraList, token, conn) // get websocket URL to stream video

			for name, url := range streamUrls {
				go StreamVideo(name, url, creds)
			}
		}()
	}

	select {}

}

func SubscribeSocket(conn *websocket.Conn, creds utils.Creds, authToken *utils.AuthToken) {
	jsonData := "[1,\"alarmRealm\",{\"roles\":{\"caller\":{\"features\":{\"caller_identification\":true,\"progressive_call_results\":true}},\"callee\":{\"features\":{\"caller_identification\":true,\"pattern_based_registration\":true,\"shared_registration\":true,\"progressive_call_results\":true,\"registration_revocation\":true}},\"publisher\":{\"features\":{\"publisher_identification\":true,\"subscriber_blackwhite_listing\":true,\"publisher_exclusion\":true}},\"subscriber\":{\"features\":{\"publisher_identification\":true,\"pattern_based_subscription\":true,\"subscription_revocation\":true}}},\"authmethods\":[\"ticket\"],\"authid\":\"%s\"}]"
	err := conn.WriteMessage(1, []byte(fmt.Sprintf(jsonData, creds.Username)))
	if err != nil {
		log.Panic(err)
	}

	socketAuth := "[5, \"%s\", {}]"

	err = conn.WriteMessage(1, []byte(fmt.Sprintf(socketAuth, authToken.Token)))
	if err != nil {
		log.Panic(err)
	}

	var x []byte

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

	socketEvent := "[32, %v, {\"match\":\"prefix\"}, \"%v\"]"
	for _, topic := range authToken.Topics {
		err = conn.WriteMessage(1, []byte(fmt.Sprintf(socketEvent, utils.RandomNo(), topic)))
		if err != nil {
			log.Panic(err)
		}
	}

	_, x, _ = conn.ReadMessage()
	//log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	//log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	//log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	//log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	//log.Println(string(x))
}

func getStreamUrl(authUser *utils.AuthorizedUser, cameraList []utils.NodeBody, token *utils.XMLResponse, conn *websocket.Conn) map[string]string {
	streamUrls := make(map[string]string)

	for i, _ := range cameraList {
		guid := utils.CreateGuid() // create unique GUID
		utils.GetLiveStreamUrl(authUser, token, cameraList[i].Id, guid)

		y := make([]interface{}, 0)

		err := conn.ReadJSON(&y)
		if err != nil {
			log.Println(err)
		}

		sr := make([]SocketResponse, 0)
		err = json.Unmarshal([]byte(fmt.Sprintf("%v", y[4])), &sr)
		if err != nil {
			log.Println(err)
		}

		pd := new(PayloadData)
		err = json.Unmarshal([]byte(sr[0].PayloadString), pd)
		if err != nil {
			log.Println(err)
		}

		streamUrl := pd.Extension[0]["streamURL"]
		streamUrls[cameraList[i].Name] = strings.Replace(streamUrl, "rtmpts", "wss", 1)
	}

	return streamUrls
}

func StreamVideo(name, url string, creds utils.Creds) {
	vidConn := utils.MakeVidWebSocket(url)
	numBytes := 0
	mutex := sync.Mutex{}
	db := golash.Debounce(func() {
		mutex.Lock()
		defer mutex.Unlock()
		log.Printf("User %v Camera %v: %v KB/s\n", creds.Username, name, numBytes/1024)
		numBytes = 0
	}, 1000)
	db.Call()

	for {
		_, x, err := vidConn.ReadMessage()
		numBytes += len(x)
		if err == io.EOF {
			break
		}
	}
}

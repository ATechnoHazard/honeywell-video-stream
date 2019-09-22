package main

import (
	"encoding/json"
	"fmt"
	"github.com/ATechnoHazard/honeywell-video-stream/utils"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SocketResponse struct {
	PayloadString string `json:"PayloadString"`
}

type PayloadData struct {
	StatusCode string              `json:"statusCode"`
	Id         string              `json:"id"`
	Extension  []map[string]string `json:"extension"`
}

var mutex2 = sync.Mutex{}
var numTotalBytes = 0

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	cred := utils.GetCreds()
	deb := utils.Debounce(clearTotalBytes, 1000)
	deb.Call()

	for _, creds := range cred {
		go func(creds utils.Creds) {
			body := utils.User{Model: creds}
			conn := utils.MakeWebsocket()

			authUser,err  := utils.MakeLoginReq(body) // Initial login request
			if err != nil {
				return
			}
			authUser.AcceptConn = conn

			token := utils.GetReqVerToken(authUser) // Get session verification token

			cameraList := utils.GetCameraList(authUser, token) // Get list of cameras

			authToken := utils.GetAuthToken(authUser, token) // Get auth token in local storage
			SubscribeSocket(conn, creds, authToken)          // subscribe for further events to socket

			streamUrls := getStreamUrl(authUser, cameraList, token, conn) // get websocket URL to stream video
			log.Println(streamUrls)

			for name, url := range streamUrls {
				go StreamVideo(name, url, creds)
				time.Sleep(2 * time.Second)
			}
		}(creds)
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

	_, _, _ = conn.ReadMessage()

	socketEvent := "[32, %v, {\"match\":\"prefix\"}, \"%v\"]"
	for _, topic := range authToken.Topics {
		err = conn.WriteMessage(1, []byte(fmt.Sprintf(socketEvent, utils.RandomNo(), topic)))
		if err != nil {
			log.Panic(err)
		}
	}

	_, _, _ = conn.ReadMessage()

	_, _, _ = conn.ReadMessage()

	_, _, _ = conn.ReadMessage()

	_, _, _ = conn.ReadMessage()

	_, _, _ = conn.ReadMessage()
}

func getStreamUrl(authUser *utils.AuthorizedUser, cameraList []utils.NodeBody, token *utils.XMLResponse, conn *websocket.Conn) map[string]string {
	streamUrls := make(map[string]string)

	for i, _ := range cameraList {
		pd := new(PayloadData)
		var sr []SocketResponse
		for len(pd.Extension) <= 0 {

			guid := utils.CreateGuid() // create unique GUID
			utils.GetLiveStreamUrl(authUser, token, cameraList[i].Id, guid)

			y := make([]interface{}, 0)

			err := conn.ReadJSON(&y)
			if err != nil {
				log.Println(err)
			}

			err = json.Unmarshal([]byte(fmt.Sprintf("%v", y[4])), &sr)
			if err != nil {
				log.Println(err)
			}

			err = json.Unmarshal([]byte(sr[0].PayloadString), pd)
			if err != nil {
				log.Println(err)
			}
		}

		streamUrl := pd.Extension[0]["streamURL"]
		streamUrls[cameraList[i].Name] = strings.Replace(streamUrl, "rtmpts", "wss", 1)
	}

	return streamUrls
}

func StreamVideo(name, url string, creds utils.Creds) {
	log.SetFormatter(&log.JSONFormatter{})
	vidConn := utils.MakeVidWebSocket(url)
	numBytes := 0
	mutex := sync.Mutex{}
	db := utils.Debounce(func() {
		mutex.Lock()
		defer mutex.Unlock()

		log.WithFields(log.Fields{
			"user":   creds.Username,
			"camera": name,
			"speed":  strconv.Itoa(numBytes/1024) + "KB/s",
		}).Info("Recieved data")

		//log.Printf("User %v Camera %v: %v KB/s\n", creds.Username, name, numBytes/1024)
		numBytes = 0
	}, 1000)
	db.Call()

	for {
		_, x, err := vidConn.ReadMessage()
		mutex2.Lock()
		numTotalBytes += len(x)
		mutex2.Unlock()
		numBytes += len(x)
		if err == io.EOF {
			break
		}
	}
}

func clearTotalBytes() {
	mutex2.Lock()
	defer mutex2.Unlock()
	speed := numTotalBytes / 1024
	numTotalBytes = 0
	log.Println("Total bandwidth consumed:", speed)
}

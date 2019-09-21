package main

import (
	"fmt"
	"github.com/ATechnoHazard/honeywell-video-stream/utils"
	"log"
)

func main() {
	conn := utils.MakeWebsocket()
	creds := utils.GetCreds()[0]
	body := utils.User{Model: creds}

	authUser := utils.MakeLoginReq(body)
	token := utils.GetReqVerToken(authUser)

	json := "[1,\"alarmRealm\",{\"roles\":{\"caller\":{\"features\":{\"caller_identification\":true,\"progressive_call_results\":true}},\"callee\":{\"features\":{\"caller_identification\":true,\"pattern_based_registration\":true,\"shared_registration\":true,\"progressive_call_results\":true,\"registration_revocation\":true}},\"publisher\":{\"features\":{\"publisher_identification\":true,\"subscriber_blackwhite_listing\":true,\"publisher_exclusion\":true}},\"subscriber\":{\"features\":{\"publisher_identification\":true,\"pattern_based_subscription\":true,\"subscription_revocation\":true}}},\"authmethods\":[\"ticket\"],\"authid\":\"%s\"}]"

	err := conn.WriteMessage(1, []byte(fmt.Sprintf(json, creds.Username)))
	if err != nil {
		log.Panic(err)
	}

	_, x, _ := conn.ReadMessage()
	log.Println(string(x))

	rootNode := utils.GetTreeViewItem(authUser, token, nil)
	log.Println("Root node", rootNode)

	sendParams := &utils.NodeBody{
		ParentId: rootNode[0].Id,
		Id:       utils.GetNextNodeId(rootNode[0].Id),
		NodeType: "Customer",
	}

	firstNode := utils.GetTreeViewItem(authUser, token, sendParams)
	log.Println("First node", firstNode)

	paramList := make([]utils.NodeBody, len(firstNode))
	responseList := make([]utils.NodeBody, 0)
	for i := range firstNode {
		paramList[i] = utils.NodeBody{
			ParentId: firstNode[i].ParentId,
			Id:       utils.GetNextNodeId(firstNode[i].ParentId),
			NodeType: "Site",
		}

		responseList = append(responseList, utils.GetTreeViewItem(authUser, token, &paramList[i])...)
	}

	//log.Println(responseList[0].Id)

	cameraList := make([]utils.NodeBody, 0)

	for _, device := range responseList {
		if device.EntityType == "CAMERA" {
			cameraList = append(cameraList, device)
		}
	}

	log.Println(cameraList)

	authToken := utils.GetAuthToken(authUser, token)

	socketAuth := "[5, \"%s\", {}]"

	err = conn.WriteMessage(1, []byte(fmt.Sprintf(socketAuth, authToken.Token)))
	if err != nil {
		log.Panic(err)
	}

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))


	socketEvent := "[32, %v, {\"match\":\"prefix\"}, \"%v\"]"
	for _, topic := range authToken.Topics {
		err = conn.WriteMessage(1, []byte(fmt.Sprintf(socketEvent, utils.RandomNo(), topic)))
		if err != nil {
			log.Panic(err)
		}
		//log.Println(fmt.Sprintf(socketEvent, utils.RandomNo(), topic))
	}

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

	guid := utils.CreateGuid()
	utils.GetLiveStreamUrl(authUser, token, cameraList[0].Id, guid)

	_, x, _ = conn.ReadMessage()
	log.Println(string(x))

}

package main

import (
	"fmt"
	"github.com/ATechnoHazard/honeywell-video-stream/utils"
)

func main() {
	fmt.Println(utils.GenPass("Password1"))
	creds := utils.GetCreds()[0]
	body := utils.User{Model: creds}

	authUser := utils.MakeLoginReq("Login/Authenticate", body)
	utils.GetReqVerToken(authUser)
}

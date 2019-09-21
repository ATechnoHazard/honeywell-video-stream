package utils

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Creds struct {
	Username string `json:"UserName"`
	Password string `json:"Password"`
}

type AuthResponse struct {
	Error    string `json:"error"`
	Success  bool   `json:"success"`
	Redirect string `json:"redirect"`
}

type AuthorizedUser struct {
	Cookies []*http.Cookie
}

type User struct {
	Model Creds `json:"loginmodel"`
}

type XMLResponse struct {
	Name  string `xml:"name,attr"`
	Type  string `xml:"type,attr"`
	Value string `xml:"value,attr"`
}

type NodeBody struct {
	ParentId      string `json:"ParentId"`
	Id            string `json:"Id"`
	NodeType      string `json:"nodeType"`
	AccountId     string `json:"AccountId"`
	Address       int    `json:"Address"`
	EntityType    string `json:"EntityType"`
	Name          string `json:"Name"`
	StatusQueryId string `json:"StatusQueryId"`
}

type AuthToken struct {
	Topics []string `json:"Topics"`
	Token string `json:"Token"`
	AuthId string `json:"AuthenticationId"`
}

func GetCreds() []Creds {
	f, err := os.Open("credentials.csv")

	if err != nil {
		log.Fatal(err)
	}

	csvReader := csv.NewReader(bufio.NewReader(f))
	var creds []Creds

	for {
		line, err := csvReader.Read() // read credentials from csv

		if err == io.EOF { // break if EOF
			break
		} else if err != nil {
			log.Fatalln(err)
		}

		creds = append(creds, Creds{ // append creds to slice
			Username: line[0],
			Password: GenPass(line[1]),
		})
	}

	creds = append(creds[:0], creds[1:]...) // delete first element from creds
	_ = f.Close()
	return creds
}

func MakeLoginReq(body User) *AuthorizedUser {
	sendBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(sendBody))

	res, err := http.Post("https://ispperf.mymaxprocloud.com/MPC/Login/Authenticate", "application/json", bytes.NewBuffer(sendBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()

	resBody := new(AuthResponse)

	err = json.NewDecoder(res.Body).Decode(resBody)

	if err != nil {
		log.Panic(err)
	}

	if !resBody.Success {
		log.Panic("Auth failed")
	}

	//log.Println(res.Cookies()[0])
	//log.Println(resBody.Redirect)

	return &AuthorizedUser{Cookies: res.Cookies()}
}

func GetReqVerToken(user *AuthorizedUser) *XMLResponse {
	client := http.Client{}
	request, err := http.NewRequest("GET", "https://ispperf.mymaxprocloud.com/MPC/page/GetChallenge", nil)
	if err != nil {
		log.Panic(err)
	}

	request.Header.Set("X-Requested-With", "XMLHttpRequest")
	for _, cookie := range user.Cookies {
		request.AddCookie(cookie)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}

	resBody := string(body) + "</input>"

	//log.Println(resBody)

	defer res.Body.Close()

	ret := new(XMLResponse)
	err = xml.Unmarshal([]byte(resBody), ret)
	if err != nil {
		log.Panic(err)
	}

	ret.Name = ret.Name[2:]

	//log.Println(ret)

	return ret
}

func GetTreeViewItem(user *AuthorizedUser, token *XMLResponse, body *NodeBody) []NodeBody {
	sendBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	client := http.Client{}
	request, err := http.NewRequest("POST", "https://ispperf.mymaxprocloud.com/MPC/ViewerMgmt/GetTreeViewItem", bytes.NewBuffer(sendBody))
	if err != nil {
		log.Panic(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("RequestVerificationToken", token.Value)
	for _, cookie := range user.Cookies {
		request.AddCookie(cookie)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}

	ret := make([]NodeBody, 0)

	err = json.Unmarshal(resBody, &ret)
	if err != nil {
		log.Panic(err)
	}

	//log.Println(string(resBody), "raw response")

	return ret
}

func GetAuthToken(user *AuthorizedUser, token *XMLResponse) *AuthToken {

	client := http.Client{}
	request, err := http.NewRequest("GET", fmt.Sprintf("https://ispperf.mymaxprocloud.com/MPC/Plugin/GetToken"), nil)
	if err != nil {
		log.Panic(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("RequestVerificationToken", token.Value)
	for _, cookie := range user.Cookies {
		request.AddCookie(cookie)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Panic(err)
	}

	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}

	ret := new(AuthToken)
	err = json.Unmarshal(resBody, ret)
	if err != nil {
		log.Panic(err)
	}

	return ret
}

func GetLiveStreamUrl(user *AuthorizedUser, token *XMLResponse, cameraId string, guid string) {

	body, err := json.Marshal(map[string]string {
		"Id": guid,
		"cameraId": cameraId,
	})

	if err != nil {
		log.Panic(err)
	}



	client := http.Client{}
	request, err := http.NewRequest("POST", "https://ispperf.mymaxprocloud.com/MPC/ViewerMgmt/GetLiveStreamUrl", bytes.NewBuffer(body))
	if err != nil {
		log.Panic(err)
	}

	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("RequestVerificationToken", token.Value)
	for _, cookie := range user.Cookies {
		request.AddCookie(cookie)
	}

	_, err = client.Do(request)
	if err != nil {
		log.Panic(err)
	}
}

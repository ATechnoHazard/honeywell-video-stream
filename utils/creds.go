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
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Value string `xml:"value,attr"`
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

func MakeLoginReq(route string, body User) *AuthorizedUser {
	sendBody, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(sendBody))

	res, err := http.Post(fmt.Sprintf("https://ispperf.mymaxprocloud.com/MPC/%s", route), "application/json", bytes.NewBuffer(sendBody))
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()

	resBody := new(AuthResponse)

	err = json.NewDecoder(res.Body).Decode(resBody)

	if err != nil {
		panic(err)
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

	log.Println(resBody)

	defer res.Body.Close()

	ret := new(XMLResponse)
	err = xml.Unmarshal([]byte(resBody), ret)
	if err != nil {
		log.Panic(err)
	}

	ret.Name = ret.Name[2:]

	log.Println(ret)

	return ret
}

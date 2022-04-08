package main

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

//Token is the object which contains the nfmt single step authentication tokens and the auth and deauth methodes.
type RestAgent struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	Expiry       float64 `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	IpAddress    string
	UserName     string
	Password     string
	Client       *http.Client
}

//NfmtAuth methode does the NFM-T Single step authentication and fills the token variables.
func (t *RestAgent) login() error {

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%v/rest-gateway/rest/api/v1/auth/token", t.IpAddress),
		strings.NewReader("grant_type=client_credentials"),
	)

	req.Header.Add("Authorization", "Basic "+t.toBase64())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	respc := make(chan *http.Response, 1)
	errc := make(chan error, 1)
	var resp *http.Response

	go func() {
		resp, err := t.Client.Do(req)
		if err != nil {
			errc <- err
		}
		respc <- resp
	}()

	select {
	case <-time.After(7 * time.Second):
		return fmt.Errorf("can't reach the server address. request timed out: %v", t.IpAddress)
	case err := <-errc:
		return err
	case resp = <-respc:
	}

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return fmt.Errorf("authentication failure: %v %v", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(body), &t)
	log.Println("REST API login: SUCCESS!")
	return nil
}

//NfmtDeauth does the deauthentication from the NFM-T.
func (t *RestAgent) Logout() error {

	req, _ := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%v/rest-gateway/rest/api/v1/auth/revocation", t.IpAddress),
		strings.NewReader(fmt.Sprintf("token=%v&token_type_hint=token", t.AccessToken)),
	)

	req.Header.Add("Authorization", "Basic "+t.toBase64())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := t.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return fmt.Errorf("logout failure: %v %v", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	log.Println("REST API logout: SUCCESS!")
	return nil
}

//HttpGet send a Get request and returns the response in json string.
func (t *RestAgent) Get(url string, header map[string]string) (string, error) {

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%v:%v", t.IpAddress, url), nil)
	if err != nil {
		return "", err
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}

	req.Header.Add("Authorization", fmt.Sprintf("%v %v", t.TokenType, t.AccessToken))
	req.Header.Add("Content-Type", "application/json")

	res, err := t.Client.Do(req)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		res.Body.Close()
		return "", fmt.Errorf("get request failure: %v %v", res.StatusCode, http.StatusText(res.StatusCode))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		return "", fmt.Errorf("failed tp read the response body for: %v, detail: %v", url, err)
	}

	return string(body), nil
}

func (t *RestAgent) PostJson(url, payload string, header map[string]string) ([]map[string]interface{}, error) {
	var jsonStr = []byte(payload)
	var response []map[string]interface{}
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://%v:%v", t.IpAddress, url),
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}
	req.Header.Add("Authorization", fmt.Sprintf("%v %v", t.TokenType, t.AccessToken))
	req.Header.Add("Content-Type", "application/json")

	resp, err := t.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		resp.Body.Close()
		return nil, fmt.Errorf("post request failure : %v %v", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed tp read the response body for: %v, detail: %v", url, err)
	}

	json.Unmarshal([]byte(body), &response)

	return response, nil
}

//toBase64 encodes the user/pass combination to Base64.
func (t *RestAgent) toBase64() string {
	auth := t.UserName + ":" + t.Password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//HttpClientCreator creates and returns an unsecure http client object.
func createClient() *http.Client {
	return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
}

func GeneralJsonDecoder(content string) (map[string]interface{}, []map[string]interface{}) {
	if string(content[0]) == "[" {
		var result []map[string]interface{}
		json.Unmarshal([]byte(content), &result)
		return nil, result
	} else {
		var result map[string]interface{}
		json.Unmarshal([]byte(content), &result)
		return result, nil
	}
}

func Init(ipaddr, uname, passw string) (RestAgent, error) {
	token := RestAgent{
		Client:    createClient(),
		IpAddress: ipaddr,
		UserName:  uname,
		Password:  passw,
	}
	if err := token.login(); err != nil {
		return token, err
	}
	return token, nil
}

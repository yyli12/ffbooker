package booker

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"

	"github.com/yyli12/ffbooker/log"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	_cacheToken     = ""
	_tokenRefreshed = false

	emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func EnsureToken() error {
	_, tokenErr := GetToken()
	if tokenErr != nil {
		log.Error("no cached token, err: %s", tokenErr)
		email, password, inputErr := inputCredentials()
		if inputErr != nil {
			log.Error("input credentials error")
			return inputErr
		}
		loginErr := Login(email, password)
		if loginErr != nil {
			log.Error("login error, err: %s", loginErr)
			return loginErr
		}
	}
	return nil
}

func inputCredentials() (string, string, error) {
	fmt.Println("Input credential to login")

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	fmt.Print("Enter Password: ")
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", err
	}
	fmt.Println()
	password := string(passwordBytes)
	return strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func Login(email, password string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("error email format")
	}
	if len(password) <= 0 {
		return fmt.Errorf("password is missing")
	}
	client := &http.Client{}
	loginUrl := "https://api-mobile.circuithq.com/api/v1/auth/login"
	params := map[string]interface{}{
		"email":         email,
		"password":      password,
		"captcha_token": "batman",
	}
	postBodyBytes, _ := json.Marshal(params)
	req, clientErr := http.NewRequest("POST", loginUrl, bytes.NewBuffer(postBodyBytes))
	if clientErr != nil {
		return clientErr
	}
	req = addHeaders(req, "") // no token when login

	respStruct := &struct {
		Error *APIError `json:"error"`
		Data  struct {
			Token string `json:"token"`
		} `json:"data"`
	}{}
	requestErr := sendRequest(client, req, respStruct)
	if requestErr != nil {
		return requestErr
	}
	if respStruct.Error != nil {
		return respStruct.Error
	}
	_cacheToken = respStruct.Data.Token
	_tokenRefreshed = true
	log.Info("login and get token: %s", _cacheToken)
	writeTokenToFile(_cacheToken)
	return nil
}

func GetToken() (string, error) {

	if _tokenRefreshed {
		log.Info("using cache token")
		if _cacheToken == "" {
			return "", fmt.Errorf("no token")
		}
		return _cacheToken, nil
	}
	log.Info("refreshing token")
	tokenByte, readTokenErr := ioutil.ReadFile(tokenFile)
	if readTokenErr != nil {
		return "", fmt.Errorf("fail to read token from file, error: " + readTokenErr.Error())
	}
	oldToken := string(tokenByte)
	if len(oldToken) == 0 {
		return "", fmt.Errorf("no cache token in file")
	}
	newToken, exchangeErr := exchangeToken(oldToken)
	if exchangeErr != nil {
		return "", fmt.Errorf("fail to exchange token, error: " + exchangeErr.Error())
	}
	_cacheToken = newToken
	_tokenRefreshed = true
	writeTokenToFile(_cacheToken)
	return _cacheToken, nil
}

func addHeaders(r *http.Request, token string) *http.Request {
	headers := getHeaders(token)
	for k, v := range headers {
		r.Header.Add(k, v)
	}
	return r
}

func getHeaders(token string) map[string]string {
	headers := map[string]string{
		"Content-Type":      "application/json",
		"User-locale":       "sg",
		"User-Country-Code": "sg",
		"Accept":            "*/*",
		"User-Brand-Code":   "ff",
		"User-Agent":        "Fitness First Asia/1.12 (com.EvolutionWellness.App.FitnessFirst; build:68; iOS 13.6.0) Alamofire/4.8.2",
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	return headers
}

func exchangeToken(oldToken string) (string, error) {
	client := &http.Client{}
	refreshTokenUrl := "https://api-mobile.circuithq.com/api/v1/auth/token/refresh"
	req, clientErr := http.NewRequest("POST", refreshTokenUrl, nil)
	if clientErr != nil {
		return "", clientErr
	}
	req = addHeaders(req, oldToken)

	respStruct := &struct {
		Error *APIError `json:"error"`
		Data  struct {
			Token string `json:"token"`
		} `json:"data"`
	}{}
	requestErr := sendRequest(client, req, respStruct)
	if requestErr != nil {
		return "", requestErr
	}
	if respStruct.Error != nil {
		return "", respStruct.Error
	}
	return respStruct.Data.Token, nil
}

func writeTokenToFile(token string) {
	log.Info("writing token into file")
	writeTokenErr := ioutil.WriteFile(tokenFile, []byte(token), 0644)
	if writeTokenErr != nil {
		log.Error("fail to write token into file, err: %s", writeTokenErr)
	}
}

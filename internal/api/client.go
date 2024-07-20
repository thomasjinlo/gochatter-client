package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sort"
)

const domain = "https://api.gochatter.app:8443"

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

type User struct {
	Id   string
	Name string
}

type ByName []User

func (a ByName) Len() int           { return len(a) }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }

type GetUserBody struct {
	Id   string
	Name string
}

type GetUsersBody struct {
	Users []GetUserBody
}

func (c *Client) GetUsers() []User {
	client := &http.Client{}
	url := domain + "/users"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Error calling getting users %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body %s", err)
	}
	var getUsers GetUsersBody
	err = json.Unmarshal(body, &getUsers)
	if err != nil {
		log.Printf("Error parsing json %s", err)
	}
	var users []User
	for _, u := range getUsers.Users {
		users = append(users, User{Id: u.Id, Name: u.Name})
	}
	sort.Sort(ByName(users))
	return users
}

type DirectMessageRequest struct {
	SourceAccountId string
	TargetAccountId string
	Content         string
}

func (c *Client) SendDirectMessage(senderId, receiverId, content string) int {
	client := &http.Client{}
	url := domain + "/direct_message"
	dm := DirectMessageRequest{
		SourceAccountId: senderId,
		TargetAccountId: receiverId,
		Content:         content,
	}
	b, err := json.Marshal(dm)
	if err != nil {
		log.Printf("Error marshalling direct message request %s", err)
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		log.Printf("Error building direct message request %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending direct message %s", err)
	}
	return res.StatusCode
}

type LoginRequest struct {
	Username string
}

func (c *Client) Login(username string) bool {
	client := &http.Client{}
	url := domain + "/login"
	lr := LoginRequest{Username: username}
	b, err := json.Marshal(lr)
	if err != nil {
		log.Printf("Error marshalling login request %s", err)
		return false
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		log.Printf("Error building login request %s", err)
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Printf("Error login request %s", err)
	}
	return res.StatusCode == http.StatusOK
}

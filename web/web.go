package web

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// MakeRequest sends an HTTP GET request to the requested host.
func MakeRequest(host string) (string, error) {
	resp, err := http.Get(host)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf(string(b))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Client is the HTTP client
type Client struct {
	client *http.Client
	user   string
	pass   string
}

// New ...
func New(user, pass string) *Client {
	return &Client{
		client: new(http.Client),
		user:   user,
		pass:   pass,
	}
}

// UpdateIP ...
func (c *Client) UpdateIP(host string) error {
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(c.user, c.pass)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	b, _ := ioutil.ReadAll(resp.Body)
	body := string(b)
	if resp.StatusCode == 200 {
		if strings.HasPrefix(body, "good") {
			fmt.Printf("Success: %v", body)
		} else if strings.HasPrefix(body, "nochg") {
			fmt.Printf("No Fail: %", body)
		} else {
			fmt.Printf(body)
		}
		return nil
	}
	return fmt.Errorf("ERROR: %v", body)
}

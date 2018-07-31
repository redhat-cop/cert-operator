package slack

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/appscode/envconfig"
	"github.com/redhat-cop/cert-operator/pkg/notifier"
)

const kind = "slack"

var validStatusCodes = []int{200, 300}

type client struct {
	opt Options
}

type Options struct {
	//AuthToken string   `envconfig:"AUTH_TOKEN" required:"true"`
	//Channel   []string `envconfig:"CHANNEL"`
	WebhookUrl string `envconfig:"WEBHOOK_URL" required:"true"`
}

var _ notifier.Chat = &client{}

func New() (*client, error) {
	var opt Options
	err := envconfig.Process(kind, &opt)
	if err != nil {
		return nil, err
	}

	return &client{opt: opt}, nil
}

func (c *client) Kind() string {
	return kind
}

/*
We may eventually want to go with something like https://github.com/appscode/go-notify/blob/master/slack/chat.go#L60-L72
*/
func (c *client) Send(message string) error {
	s := fmt.Sprintf(`{"text": "%s"}`, message)
	var jsonStr = []byte(s)
	req, err := http.NewRequest("POST", c.opt.WebhookUrl, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !validResponse(resp.StatusCode) {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
		return errors.New("Got a bad response code")
	}

	return nil
}

func validResponse(s int) bool {
	for _, a := range validStatusCodes {
		if a == s {
			return true
		}
	}
	return false
}

package notifier

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/openshift/api/route/v1"
)

type Notifier struct {
	Name           string `json:"name"`
	Kind           string `json:"kind"`
	IntegrationURL string `json:"integration_url"`
}

type Provider struct {
	Name string
	Url  string
}

type ResultMessage struct {
	ErrorLog string
	InfoLog  string
	DebugLog string
}

func NewNotifiers() []Notifier {
	var notifiers []Notifier
	return notifiers
}

func Notify(route *v1.Route) ([]byte, error) {
	//for index, notifier := range
	return nil, nil
}

func NotifyOnce(route *v1.Route, provider Provider) ([]byte, error) {
	switch provider.Name {
	case "slack":
		return notifySlack(route, provider.Url)
	default:
		return json.Marshal(ResultMessage{"No provider configured", "", ""})
	}
	return nil, nil
}

func notifySlack(route *v1.Route, url string) ([]byte, error) {
	message := "" +
		"_Namespace_: *" + route.ObjectMeta.Namespace + "*\n" +
		"_Route Name_: *" + route.ObjectMeta.Name + "*\n"

	var jsonStr = []byte(`
		{
		    "attachments": [
		        {
		            "pretext": "*Received a _new certificate_ request*\n",
		            "text": "` + message + `",
		            "mrkdwn_in": [
		                "text",
		                "pretext"
		            ]
		        }
		    ]
		}
	`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	result := ResultMessage{
		"",
		"Slack message sent",
		"Response Status: " + resp.Status +
			"Response Headers: " +
			"Response Body: " + string(body),
	}

	return json.Marshal(result)
}

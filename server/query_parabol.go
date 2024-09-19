package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/yaronf/httpsign"
)

func getJson(res *http.Response, target interface{}) error {
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(target)
}

type MeetingTemplatesResponse struct {
	AvailableTemplates []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		IllustrationURL string `json:"illustrationUrl"`
		OrgID           string `json:"orgId"`
		TeamID          string `json:"teamId"`
	} `json:"availableTemplates"`
	Teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"teams"`
}

type Variables struct {
	Email string `json:"email"`
}

type Query struct {
	Query     string    `json:"query"`
	Variables Variables `json:"variables"`
}

func (p *Plugin) queryMeetingTemplates(email string) *MeetingTemplatesResponse {
	config := p.getConfiguration()
	privKey := []byte(config.ParabolToken)
	url := config.ParabolURL
	//privKey := []byte("123")
	//url := "http://localhost:3001/mattermost"

	// Create a signer and a wrapped HTTP client
	signer, err := httpsign.NewJWSSigner(jwa.SignatureAlgorithm("HS256"), privKey, httpsign.NewSignConfig().SignAlg(false),
		httpsign.Headers("@request-target", "Content-Digest")) // The Content-Digest header will be auto-generated
	if err != nil {
		fmt.Print("GEORG signer error", err)
		return nil
	}

	client := httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName("sig1").SetSigner(signer)) //.SetComputeDigest(true)) // sign requests, don't verify responses

	// Send an HTTP POST, get response -- signing happens behind the scenes
	query := Query{
		Query: "meetingTemplates",
		Variables: Variables{
			Email: email,
		}}
	body, _ := json.Marshal(query)
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(body)))

	if err != nil {
		fmt.Print("GEORG", err)
		return nil
	} else {
		response := new(MeetingTemplatesResponse)
		err := getJson(res, response)
		if err != nil {
			fmt.Print("GEORG", err)
			return nil
		}
		fmt.Print("GEORG", response.Teams[0])
		return response
	}

}

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"net/http"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/yaronf/httpsign"
)


func getJson(body io.ReadCloser, target interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(target)
}

type MeetingTemplatesResponse struct {
	AvailableTemplates []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Type            string `json:"type"`
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
	Email string `json:"email"`
}

type StartVariables struct {
		TeamId string `json:"teamId"`
		TemplateID  string `json:"selectedTemplateId"`
	}

type StartQuery struct {
	Query     string    `json:"query"`
	Variables StartVariables`json:"variables"`
	Email string `json:"email"`
}

type StartActivitySubmit struct {
	Type       string `json:"type"`
	CallbackID string `json:"callback_id"`
	State      string `json:"state"`
	UserID     string `json:"user_id"`
	ChannelID  string `json:"channel_id"`
	TeamID     string `json:"team_id"`
	Submission struct {
		Team     string `json:"team"`
		Template string `json:"template"`
	} `json:"submission"`
	Cancelled bool `json:"cancelled"`
}

func (p *Plugin) newSigningClient() *httpsign.Client {
	config := p.getConfiguration()
	privKey := []byte(config.ParabolToken)
	//url := config.ParabolURL
	//privKey := []byte("123")
	//url := "http://localhost:3001/mattermost"

	// Create a signer and a wrapped HTTP client
	signer, err := httpsign.NewJWSSigner(jwa.SignatureAlgorithm("HS256"), privKey, httpsign.NewSignConfig().SignAlg(false),
		httpsign.Headers("@request-target", "Content-Digest")) // The Content-Digest header will be auto-generated
	if err != nil {
		fmt.Print("GEORG signer error", err)
		return nil
	}

	client := httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName("sig1").SetSigner(signer))
	return client
}



func (p *Plugin) queryMeetingTemplates(email string) *MeetingTemplatesResponse {
	config := p.getConfiguration()
	url := config.ParabolURL
	client := p.newSigningClient()
	query := Query{
		Query: "meetingTemplates",
		Variables: Variables{
			Email: email,
		},
		Email: email,
	}
	body, _ := json.Marshal(query)
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(body)))

	if err != nil {
		fmt.Print("GEORG", err)
		return nil
	} else {
		response := new(MeetingTemplatesResponse)
		err := getJson(res.Body, response)
		if err != nil {
			fmt.Print("GEORG", err)
			return nil
		}
		fmt.Print("GEORG", response.Teams[0])
		return response
	}

}

func (p *Plugin) startActivity(w http.ResponseWriter, r *http.Request) {
	fmt.Print("GEORG r", r)
	response := new(StartActivitySubmit)
	err := getJson(r.Body, response)
	if err != nil {
		fmt.Print("GEORG err", err)
		return
	}
	user, _ := p.API.GetUser(response.UserID)

	fmt.Println("GEORG response", response)
	team := response.Submission.Team
	templateType, templateId := func ()(string, string) {
		t := strings.SplitN(response.Submission.Template, ":", 2)
		return t[0], t[1]
	}()

	fmt.Println("GEORG tea teamm", team, templateType, templateId)


	config := p.getConfiguration()
	url := config.ParabolURL
	client := p.newSigningClient()
	query := StartQuery {
		Query: "startRetrospective",
		Variables: StartVariables {
			TeamId: team,
			TemplateID: templateId,
		},
		Email: user.Email,
	}
	body, _ := json.Marshal(query)
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(body)))

	if err != nil {
		fmt.Print("GEORG", err)
		return
	} else {
		response := new(MeetingTemplatesResponse)
		err := getJson(res.Body, response)
		if err != nil {
			fmt.Print("GEORG", err)
			return
		}
		fmt.Print("GEORG", response)
		return
	}
}

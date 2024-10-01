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

type MeetingTemplatesResponse struct {
	AvailableTemplates []struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Type            string `json:"type"`
		IllustrationURL string `json:"illustrationUrl"`
		OrgID           string `json:"orgId"`
		TeamID          string `json:"teamId"`
		Scope	   string `json:"scope"`
	} `json:"availableTemplates"`
	Teams []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		OrgID string `json:"orgId"`
                RetroSettings struct {
					ID               string   `json:"id"`
					PhaseTypes       []string `json:"phaseTypes"`
					DisableAnonymity bool     `json:"disableAnonymity"`
				} `json:"retroSettings"`
				PokerSettings struct {
					ID         string   `json:"id"`
					PhaseTypes []string `json:"phaseTypes"`
				} `json:"pokerSettings"`
				ActionSettings struct {
					ID         string   `json:"id"`
					PhaseTypes []string `json:"phaseTypes"`
				} `json:"actionSettings"`
	} `json:"teams"`
}

type EmailVariables struct {
	Email string `json:"email"`
}

type Query[V any] struct {
	Query     string    `json:"query"`
	Variables V `json:"variables"`
	Email string `json:"email"`
}

type StartVariables struct {
	TeamId string `json:"teamId"`
	TemplateID  string `json:"selectedTemplateId"`
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

func newSigningClient(privKey []byte) *httpsign.Client {
	signer, err := httpsign.NewJWSSigner(jwa.SignatureAlgorithm("HS256"), privKey, httpsign.NewSignConfig().SignAlg(false),
		httpsign.Headers("@request-target", "Content-Digest"))
	if err != nil {
		fmt.Print("GEORG signer error", err)
		return nil
	}

	client := httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName("sig1").SetSigner(signer))
	return client
}

func getJson(body io.ReadCloser, target interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(target)
}

func query[R any, V any](p *Plugin, query Query[V]) *R {
	config := p.getConfiguration()
	url := config.ParabolURL
	privKey := []byte(config.ParabolToken)

	client := newSigningClient(privKey)
	body, _ := json.Marshal(query)
	res, err := client.Post(url, "application/json", bufio.NewReader(bytes.NewReader(body)))

	if err != nil {
		fmt.Print("GEORG", err)
		return nil
	} else {
		response := new(R)
		err := getJson(res.Body, response)
		if err != nil {
			fmt.Print("GEORG", err)
			return nil
		}
		return response
	}
}

func (p *Plugin) queryMeetingTemplates(email string) *MeetingTemplatesResponse {
	return query[MeetingTemplatesResponse](p, Query[EmailVariables]{
		Query: "meetingTemplates",
		Variables: EmailVariables{
			Email: email,
		},
		Email: email,
	})
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

	query[any](p, Query[StartVariables]{
		Query: "startRetrospective",
		Variables: StartVariables {
			TeamId: team,
			TemplateID: templateId,
		},
		Email: user.Email,
	})
}


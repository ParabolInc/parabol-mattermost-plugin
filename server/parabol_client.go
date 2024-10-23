package main

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/yaronf/httpsign"
)

func getJson(body io.ReadCloser, target interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(target)
}

func NewSigningClient(privKey []byte) (*httpsign.Client, error) {
	signer, err := httpsign.NewJWSSigner(jwa.SignatureAlgorithm("HS256"), privKey, httpsign.NewSignConfig().SignAlg(false),
		httpsign.Headers("@request-target", "Content-Digest"))
	if err != nil {
		fmt.Print("GEORG signer error", err)
		return nil, err
	}

	client := httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName("sig1").SetSigner(signer))
	return client, nil
}


package main

import (
	"encoding/json"
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
		httpsign.Headers("@target-uri", "Content-Digest"))
	if err != nil {
		return nil, err
	}

	client := httpsign.NewDefaultClient(httpsign.NewClientConfig().SetSignatureName("mattermost").SetSigner(signer))
	return client, nil
}

func NewVerifier(privKey []byte) (*httpsign.Verifier, error) {
	//verifier, err := httpsign.NewHMACSHA256Verifier(privKey, httpsign.NewVerifyConfig(), httpsign.Headers("@target-uri", "content-digest"))
	verifier, err := httpsign.NewJWSVerifier(jwa.SignatureAlgorithm("HS256"), privKey, httpsign.NewVerifyConfig(),
		httpsign.Headers("@target-uri", "content-digest"))
	if err != nil {
		return nil, err
	}

	return verifier, nil
}

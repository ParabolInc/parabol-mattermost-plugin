package main

import (
	"encoding/json"
	"io"
)

func getJSON(body io.ReadCloser, target any) error {
	defer func() { _ = body.Close() }()
	return json.NewDecoder(body).Decode(target)
}

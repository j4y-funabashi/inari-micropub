package web

import (
	"bytes"
	"encoding/json"

	"github.com/j4y_funabashi/inari-micropub/pkg/app"
)

type Parser struct{}

func NewParser() Parser {
	return Parser{}
}

func (p Parser) ParseUpdateRequest(bodyBytes []byte) app.UpdatePostRequest {
	out := app.UpdatePostRequest{}

	typ, props := parseUpdateAction(json.NewDecoder(bytes.NewBuffer(bodyBytes)))
	out.URL = parseUpdateURL(json.NewDecoder(bytes.NewBuffer(bodyBytes)))
	out.Type = typ
	out.Properties = props

	return out
}

func parseUpdateURL(decoder *json.Decoder) string {
	v := struct {
		URL string `json:"url"`
	}{}
	err := decoder.Decode(&v)
	if err != nil {
		return ""
	}
	return v.URL
}

func parseUpdateAction(decoder *json.Decoder) (string, map[string][]interface{}) {
	v := struct {
		Replace map[string][]interface{} `json:"replace"`
	}{}
	err := decoder.Decode(&v)
	if err != nil {
		return "", map[string][]interface{}{}
	}
	return "replace", v.Replace
}

func (p Parser) ParseMicropubPostAction(bodyBytes []byte) string {

	decoder := json.NewDecoder(bytes.NewBuffer(bodyBytes))
	e := struct {
		Action string `json:"action"`
	}{}
	err := decoder.Decode(&e)
	if err != nil {
		return "create"
	}

	if e.Action != "" {
		return e.Action
	}
	return "create"
}

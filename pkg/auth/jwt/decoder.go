package jwt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)

type Token struct {
	Subject string `json:"sub,omitempty"`
	Email   string `json:"email,omitempty"`
}

func DecodeTokenFile(tokenPath string) (token Token, err error) {
	rawBytes, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		return token, err
	}
	return decode(string(rawBytes))
}

func decode(rawToken string) (token Token, err error) {
	payload, err := decodePayloadAsRawJSON(rawToken)
	if err != nil {
		return token, err
	}
	if err := json.NewDecoder(bytes.NewReader(payload)).Decode(&token); err != nil {
		return token, err
	}
	return token, nil
}

func decodePayloadAsRawJSON(s string) ([]byte, error) {
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("wants %d segments but got %d segments", 3, len(parts))
	}
	payloadJSON, err := decodePayload(parts[1])
	if err != nil {
		return nil, fmt.Errorf("could not decode the payload: %v", err)
	}
	return payloadJSON, nil
}

func decodePayload(payload string) ([]byte, error) {
	tokenBytes, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(payload)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %v", err)
	}
	return tokenBytes, nil
}

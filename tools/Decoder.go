package tools

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
)

func EncodeToBase32(v any) (string, error) {
	var buffer bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buffer)
	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}
	_ = encoder.Close()
	return buffer.String(), nil
}

func DecodeToBase32(dest any, input string) error {
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(input))
	err := json.NewDecoder(decoder).Decode(dest)
	if err != nil {
		return err
	}

	return nil
}

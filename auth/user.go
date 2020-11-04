package auth

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"strings"
)

type user struct {
	Id            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Picture       string `json:"picture"`
}

type Map map[string]interface{}

func toMap(body []byte) (Map, error) {
	var data Map

	err := json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func toStruct(body []byte) (user, error) {
	var u user

	err := json.Unmarshal(body, &u)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (u *user) email() string {
	if u.VerifiedEmail {
		return u.Email
	}

	return ""
}

func Decoder(str string) string {
	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(str))
	b, err := ioutil.ReadAll(decoder)
	if err != nil {
		return ""
	}

	return string(b)
}

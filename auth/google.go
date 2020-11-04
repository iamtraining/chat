package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	Conf = &oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		Endpoint:     google.Endpoint,
		RedirectURL:  "http://localhost:3000/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
	}
)

const (
	URLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	url := Conf.AuthCodeURL(generateRandomState(w))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	state, _ := r.Cookie("remember_state")

	if r.FormValue("state") != state.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	data, err := getDataFromResponse(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	fmt.Fprintf(w, "user info %s\n", data)
}

func generateRandomState(w http.ResponseWriter) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)

	state := base64.URLEncoding.EncodeToString(bytes)

	cookie := &http.Cookie{
		Name:  "remember_state",
		Value: state,
	}

	http.SetCookie(w, cookie)

	return state
}

func getDataFromResponse(code string) ([]byte, error) {
	token, err := Conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, fmt.Errorf("couldnt get token")
	}

	resp, err := http.Get(URLAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("couldnt create get request")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("couldnt parse response body")
	}

	return b, nil
}

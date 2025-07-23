package oauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// TODO put this OAuth handling into its own file?
type GithubAccessTokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	Scope            string `json:"scope"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorUri         string `json:"error_uri"`
}

func RequestAccessToken(code string) (responseStruct *GithubAccessTokenResponse, error error) {
	values := url.Values{}
	values.Set("code", code)
	values.Set("client_id", os.Getenv("GITHUB_CLIENT_ID"))
	values.Set("client_secret", os.Getenv("GITHUB_CLIENT_SECRET"))

	req, err := http.NewRequest(
		http.MethodPost,
		"https://github.com/login/oauth/access_token",
		strings.NewReader(values.Encode()),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	atResponse := &GithubAccessTokenResponse{}

	if err = json.NewDecoder(res.Body).Decode(atResponse); err != nil {
		return nil, err
	}

	// expand on the scope checking / can't be bothered right now
	if atResponse.Scope != "user:email" {
		err := errors.New(fmt.Sprintf("Scope: %s does not match expected 'user:email'", atResponse.Scope))
		return nil, err
	}

	return atResponse, nil

}

type UserGetMailResponse struct {
	Email      string `json:"email"`
	Primary    bool   `json:"primary"`
	Verified   bool   `json:"verified"`
	Visibility string `json:"visibility"`
}

func GetUserMails(accessToken string) (responses []UserGetMailResponse, err error) {
	//get first public email (should be primary every time. api does not allow specific query for the primary email)
	req, _ := http.NewRequest(
		http.MethodGet,
		"https://api.github.com/user/public_emails",
		nil,
	)

	req.Header.Add("Authorization", "Bearerr "+accessToken)
	req.Header.Add("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if err = json.NewDecoder(res.Body).Decode(&responses); err != nil {
		return nil, err
	}

	return responses, nil
}

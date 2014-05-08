package oauth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-martini/martini"
)

var (
	ghProfileURL = "https://api.github.com/user"
)

// Github stores the access and refresh tokens along with the users profile.
type Github struct {
	Errors       []error
	AccessToken  string
	RefreshToken string
	Profile      GithubProfile
}

// GithubProfile stores information about the user from Github.
type GithubProfile struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Login   string `json:"login"`
	HTMLURL string `json:"html_url"`
	Email   string `json:"email"`
}


func AuthGithub(opts *OAuth2Options) martini.Handler {
	opts.AuthURL = "https://github.com/login/oauth/authorize"
	opts.TokenURL = "https://github.com/login/oauth/access_token"

	return func(r *http.Request, w http.ResponseWriter, c martini.Context) {
		transport := makeTransport(opts, r)
		cbPath := ""
		if u, err := url.Parse(transport.Config.RedirectURL); err == nil {
			cbPath = u.Path
		}
		if r.URL.Path != cbPath {
			http.Redirect(w, r, transport.Config.AuthCodeURL(""), http.StatusFound)
			return
		}
		gh := &Github{}
		defer c.Map(gh)
		code := r.FormValue("code")
		tk, err := transport.Exchange(code)
		if err != nil {
			gh.Errors = append(gh.Errors, err)
			return
		}
		gh.AccessToken = tk.AccessToken
		gh.RefreshToken = tk.RefreshToken
		resp, err := transport.Client().Get(ghProfileURL)
		if err != nil {
			gh.Errors = append(gh.Errors, err)
			return
		}
		defer resp.Body.Close()
		profile := &GithubProfile{}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			gh.Errors = append(gh.Errors, err)
			return
		}
		if err := json.Unmarshal(data, profile); err != nil {
			gh.Errors = append(gh.Errors, err)
			return
		}
		gh.Profile = *profile
		return
	}
}

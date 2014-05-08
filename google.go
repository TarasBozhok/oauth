package oauth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-martini/martini"
)

var (
	googleProfileURL = "https://www.googleapis.com/oauth2/v1/userinfo"
)

// Google stores the access and refresh tokens along with the user profile.
type Google struct {
	Errors       []error
	AccessToken  string
	RefreshToken string
	Profile      GoogleProfile
}

// GoogleProfile stores information from the users google+ profile.
type GoogleProfile struct {
	ID          string `json:"id"`
	Name 	    string `json:"name"`
	LastName    string `json:"family_name"`
	FirstName   string `json:"given_name"`
	Email       string `json:"email"`
	Avatar 		string
	GeneratedPwd string
}

type GoogleAvatar struct {
	Picture		string `json:"picture"`
}


func AuthGoogle(opts *OAuth2Options) martini.Handler {
	opts.AuthURL = "https://accounts.google.com/o/oauth2/auth"
	opts.TokenURL = "https://accounts.google.com/o/oauth2/token"

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
		goog := &Google{}
		defer c.Map(goog)
		code := r.FormValue("code")
		tk, err := transport.Exchange(code)
		if err != nil {
			goog.Errors = append(goog.Errors, err)
			return
		}
		goog.AccessToken = tk.AccessToken
		goog.RefreshToken = tk.RefreshToken
		resp, err := transport.Client().Get(googleProfileURL)
		if err != nil {
			goog.Errors = append(goog.Errors, err)
			return
		}
		defer resp.Body.Close()
		profile := &GoogleProfile{}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			goog.Errors = append(goog.Errors, err)
			return
		}
		if err := json.Unmarshal(data, profile); err != nil {
			goog.Errors = append(goog.Errors, err)
			return
		}
		avatarResp, err := http.Get("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token="+goog.AccessToken)
		if err == nil {
			defer avatarResp.Body.Close()
			Avatar := &GoogleAvatar{}
			data, err := ioutil.ReadAll(avatarResp.Body); if err == nil {
				json.Unmarshal(data, Avatar)
				profile.Avatar = Avatar.Picture
			}
		}
		goog.Profile = *profile
		return
	}
}

package oauth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/go-martini/martini"
)

var (
	fbProfileURL = "https://graph.facebook.com/me"
)

// Facebook stores the access and refresh tokens along with the users
// profile.
type Facebook struct {
	Errors       []error
	AccessToken  string
	RefreshToken string
	Profile      FacebookProfile
}

// FacebookProfile stores information about the user from facebook.
type FacebookProfile struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Name       string `json:"name"`
	LastName   string `json:"last_name"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	Gender     string `json:"gender"`
	Link       string `json:"link"`
	Email      string `json:"email"`
	Avatar 	   string
	GeneratedPwd string
}


 type FacebookAvatar struct {
	 Picture AvatarResponseData
	 ID 	 string `json:"id"`
 }

type AvatarResponseData struct {
	 Data AvatarResponseDataContent
 }

type AvatarResponseDataContent struct {
	IsSilhuette string `json:"is_silhouette"`
	Url string `json:"url"`
 }


func AuthFacebook(opts *OAuth2Options) martini.Handler {
	opts.AuthURL = "https://www.facebook.com/dialog/oauth"
	opts.TokenURL = "https://graph.facebook.com/oauth/access_token"

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
		fb := &Facebook{}
		defer c.Map(fb)
		code := r.FormValue("code")
		tk, err := transport.Exchange(code)
		if err != nil {
			fb.Errors = append(fb.Errors, err)
			return
		}
		fb.AccessToken = tk.AccessToken
		fb.RefreshToken = tk.RefreshToken
		resp, err := transport.Client().Get(fbProfileURL)
		if err != nil {
			fb.Errors = append(fb.Errors, err)
			return
		}
		defer resp.Body.Close()
		profile := &FacebookProfile{}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fb.Errors = append(fb.Errors, err)
			return
		}
		if err := json.Unmarshal(data, profile); err != nil {
			fb.Errors = append(fb.Errors, err)
			return
		}
		avatarResp, err := http.Get("https://graph.facebook.com/"+profile.ID+"?fields=picture.type(large)")
		if err == nil {
			defer avatarResp.Body.Close()
			Avatar := &FacebookAvatar{}
			data, err := ioutil.ReadAll(avatarResp.Body); if err == nil {
				json.Unmarshal(data, Avatar)
				profile.Avatar = Avatar.Picture.Data.Url
			}
		}
		fb.Profile = *profile
		return
	}
}

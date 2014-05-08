package main

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"github.com/Tossimo/oauth"
)

func main() {
	m := App()
	http.ListenAndServe("localhost:3000", m)
}

// App returns a new martini app configured with middleware.
func App() *martini.ClassicMartini {
	m := martini.Classic()

	googleOpts := &oauth.OAuth2Options{
		ClientID:     "",																								//provide your own
		ClientSecret: "",																								//provide your own
		RedirectURL:  "http://HOST:PORT/auth/callback/google",															//change HOST and PORT
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}

	fbOpts := &oauth.OAuth2Options{
		ClientID: "",																									//provide your own
		ClientSecret: "",																								//provide your own
		RedirectURL: "http://HOST:PORT/auth/callback/facebook",															//change HOST and PORT
		Scopes: []string{"email"},
	}

	// Initialize Middleware.
	m.Use(render.Renderer(render.Options{Extensions: []string{".html"}}))

	// Application Routes.

	m.Get("/login", func(r render.Render) {
		r.HTML(http.StatusOK, "login", nil)
	})

	//Login with Google
	m.Get("/auth/google", oauth.AuthGoogle(googleOpts))
	m.Get("/auth/callback/google", oauth.AuthGoogle(googleOpts), func(goog *oauth.Google, r render.Render, w http.ResponseWriter) {
		// Handle any errors.
		if len(goog.Errors) > 0 {
			http.Error(w, "Oauth failure", http.StatusInternalServerError)
			return
		}
		// Do something in a database to create or find the user by the Google profile id.
		// for now just pass the google display name
		goog.Profile.GeneratedPwd = oauth.Rand(12)
		r.HTML(http.StatusOK, "home", goog.Profile)
	})

	//Login with Facebook
	m.Get("/auth/facebook", oauth.AuthFacebook(fbOpts))
	m.Get("/auth/callback/facebook", oauth.AuthFacebook(fbOpts), func(fb *oauth.Facebook, r render.Render, w http.ResponseWriter) {
		// Handle any errors.
		if len(fb.Errors) > 0 {
			http.Error(w, "Oauth failure", http.StatusInternalServerError)
			return
		}
		// Do something in a database to create or find the user by the facebook profile id.
		//user := findOrCreateByFacebookID(fb.Profile.ID)
		fb.Profile.GeneratedPwd = oauth.Rand(12)
		r.HTML(http.StatusOK, "home", fb.Profile)
	})


	return m
}


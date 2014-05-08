package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/ottemo/ottemo-go/config"
	"log"
	"net/http"
	"github.com/Tossimo/oauth"
)

var conf *config.Config

// init - initialization and configuration
func init() {
	// read in config file
	filename, err := config.DiscoverConfig("config.json", "local.json")
	if err != nil {
		log.Fatal("No configuration file found")
	}
	conf, err = config.NewConfig(filename)
	if err != nil {
		log.Fatalf("Error parsing config.json: %s", err.Error())
	}

}

func main() {
	m := App()
	listen := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	fmt.Printf("Ottemo listening on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, m))
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

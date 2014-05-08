// File: ottemo.go
//
// Author(s): Tom Steele, James Vastbinder

// Open Source eCommerce for the discriminating retailer.
//
// Built on Go, by gophers.
//
// Ottemo makes use of the most excellent Martini framework and is committed to
// building the fastest, most scalable ecommerce solutions possible.
//
// To contribute or use Ottemo:
// 	go get github.com/ottemo/ottemo-go
//
// License and Copyright
// 	Copyright (c) 2014 Ottemo
//
// 	Permission is hereby granted, free of charge, to any person obtaining a copy
// 	of this software and associated documentation files (the "Software"), to deal
// 	in the Software without restriction, including without limitation the rights
// 	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// 	copies of the Software, and to permit persons to whom the Software is
// 	furnished to do so, subject to the following conditions:
//
// 	The above copyright notice and this permission notice shall be included in all
// 	copies or substantial portions of the Software.
//
// 	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// 	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// 	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// 	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// 	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// 	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// 	SOFTWARE.
package main

import (
	"fmt"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/secure"
	"github.com/martini-contrib/sessions"
	"github.com/ottemo/ottemo-go/auth"
	"github.com/ottemo/ottemo-go/config"
	"github.com/ottemo/ottemo-go/product"
	"github.com/ottemo/ottemo-go/visitor"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"github.com/tossimo/pwd_gen"
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
	//log.Fatal(conf)
	fmt.Printf("Ottemo listening on %s\n", listen)
	log.Fatal(http.ListenAndServe(listen, m))
}

// App returns a new martini app configured with middleware.
func App() *martini.ClassicMartini {
	m := martini.Classic()

	googleOpts := &oauth.OAuth2Options{
		ClientID:     "676130063095-m5hnkl8q4bjuij49mo4is6sdoc6uf7pi.apps.googleusercontent.com",
		ClientSecret: "wm8N23RR5UEgmrZlAaw4zW-H",
		RedirectURL:  "http://localhost:3000/auth/callback/google",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
	}

	fbOpts := &oauth.OAuth2Options{
		ClientID: "488962567871136",
		ClientSecret: "f01a43661ec22d629d5cb4ea2410d24d",
		RedirectURL: "http://localhost:3000/auth/callback/facebook",
		Scopes: []string{"email"},
	}

	store := sessions.NewCookieStore([]byte("secret"))
	m.Use(sessions.Sessions("ottemo", store))
	// Set security values in production.
	if martini.Env == martini.Prod {
		m.Use(secure.Secure(secure.Options{
			SSLRedirect: true,
		}))
	}

	// Initialize Middleware.
	m.Use(Db(conf.DbURL, conf.DbName))
	m.Use(render.Renderer(render.Options{Extensions: []string{".html"}}))
	m.Use(visitor.SessionVisitor)

	// Application Routes.
	m.Get("/", func(r render.Render) {
		r.HTML(http.StatusOK, "index", nil)
	})
	m.Get("/login", func(r render.Render) {
		r.HTML(http.StatusOK, "login", nil)
	})
	m.Post("/login", binding.Bind(auth.LocalLoginRequest{}), auth.LocalLoginHandler)
	m.Get("/register", func(r render.Render) {
		r.HTML(http.StatusOK, "register", nil)
	})
	m.Post("/register", binding.Bind(visitor.LocalVisitor{}), visitor.RegisterLocalHandler)

	//Login with Google
	m.Get("/auth/google", oauth.AuthGoogle(googleOpts))
	m.Get("/auth/callback/google", oauth.AuthGoogle(googleOpts), func(goog *oauth.Google, r render.Render, s sessions.Session, w http.ResponseWriter) {
		// Handle any errors.
		if len(goog.Errors) > 0 {
			http.Error(w, "Oauth failure", http.StatusInternalServerError)
			return
		}
		// Do something in a database to create or find the user by the Google profile id.
		// for now just pass the google display name
		goog.Profile.GeneratedPwd = pwd_gen.Rand(12)
		s.Set("userID", goog.Profile.ID)
		r.HTML(http.StatusOK, "home", goog.Profile)
	})

	//Login with Facebook
	m.Get("/auth/facebook", oauth.AuthFacebook(fbOpts))
	m.Get("/auth/callback/facebook", oauth.AuthFacebook(fbOpts), func(fb *oauth.Facebook, r render.Render, s sessions.Session, w http.ResponseWriter) {
		// Handle any errors.
		if len(fb.Errors) > 0 {
			http.Error(w, "Oauth failure", http.StatusInternalServerError)
			return
		}
		// Do something in a database to create or find the user by the facebook profile id.
		//user := findOrCreateByFacebookID(fb.Profile.ID)
		fb.Profile.GeneratedPwd = pwd_gen.Rand(12)
		s.Set("userID", fb.Profile.ID)
		r.HTML(http.StatusOK, "home", fb.Profile)
	})

	// REST API Routes
	// TODO: better error handling when resources not found

	// Visitor API CRUD
	m.Get("/api/v1/visitors", visitor.GetAllVisitors)
	m.Get("/api/v1/visitors/:id", visitor.GetLocalVisitor)
	m.Post("/api/v1/visitors", binding.Bind(visitor.LocalVisitor{}), visitor.CreateLocalVisitor)
	m.Delete("/api/v1/visitors/:id", visitor.DeleteVisitorByID)
	m.Get("/api/v1/visitors/email/:email", visitor.GetLocalVisitorByEmail)
	m.Put("/api/v1/visitors/:id/:email", visitor.UpdateVisitorEmail)
	m.Delete("/api/v1/visitors/email/:email", visitor.DeleteVisitorByEmail)
	m.Put("/api/v1/visitors/:id/:fname/:lname", visitor.UpdateVisitorName)
	// m.Post("api/v1/visitors/phone/:id", binding.Bind(visitor.PhoneNumber{}), visitor.AddVisitorPhone)

	// Admin Tool Routes
	m.Get("/admintool", func(r render.Render) {
		r.HTML(http.StatusOK, "admintool/index", nil)
	})
	m.Get("/admintool/product", func(r render.Render) {
		r.HTML(http.StatusOK, "admintool/catalog/product/index", nil)
	})
	m.Get("/admintool/product/new", func(r render.Render) {
		r.HTML(http.StatusOK, "admintool/catalog/product/index", nil)
	})
	m.Post("/admintool/product/new", binding.Bind(product.ProductRequest{}), product.NewProductHandler)
	m.Get("/admintool/product/update", func(r render.Render) {
		r.HTML(http.StatusOK, "admintool/catalog/product/index", nil)
	})
	//	m.Post("/admintool/product/update", binding.Bind(product.ProductRequest{}), product.UpdateProductHandler)

	return m
}

// Db maps a MongoDb instance to the current request context.
func Db(url, name string) martini.Handler {
	s, err := mgo.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	return func(c martini.Context) {
		clone := s.Clone()
		c.Map(clone.DB(name))
		defer clone.Close()
		c.Next()
	}
}

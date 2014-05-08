oauth
=====

oauth implementation on Go
=====
 To use this library add next routes to main file:
 
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
                goog.Profile.GeneratedPwd = oauth.Rand(12)
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
                fb.Profile.GeneratedPwd = oauth.Rand(12)
                s.Set("userID", fb.Profile.ID)
                r.HTML(http.StatusOK, "home", fb.Profile)
        })

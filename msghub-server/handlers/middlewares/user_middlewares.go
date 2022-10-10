package middlewares

import (
	jwtPkg "msghub-server/utils/jwt"
	"net/http"
)

func UserAuthorizationBeforeLogin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				handler.ServeHTTP(w, r)
			}
		}()

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("Cookie not found!")
			}
			panic("Unknown error occurred!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if claim.IsAuthenticated {
			http.Redirect(w, r, "/user/dashboard", http.StatusFound)
		} else {
			handler.ServeHTTP(w, r)
		}
	}
}

func UserAuthorizationAfterLogin(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// recovers panic
			if e := recover(); e != nil {
				http.Redirect(w, r, "/", http.StatusFound)
			}
		}()

		c, err1 := r.Cookie("userToken")
		if err1 != nil {
			if err1 == http.ErrNoCookie {
				panic("Cookie not found!")
			}
			panic("Unknown error occurred!")
		}

		claim := jwtPkg.GetValueFromJwt(c)
		if claim.IsAuthenticated == false {
			http.Redirect(w, r, "/", http.StatusFound)
		} else {

			handler.ServeHTTP(w, r)
		}
	}
}

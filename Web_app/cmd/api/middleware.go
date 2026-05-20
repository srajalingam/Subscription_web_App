package main

import "net/http"

//Auth
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := app.authenticateToken(r)
		if err != nil {
			app.invalidCredentialsResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

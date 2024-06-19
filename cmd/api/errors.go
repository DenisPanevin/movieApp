package main

import (
	"fmt"
	"net/http"
)

func (app *application) logErr(r *http.Request, err error) {
	app.logger.PrintError(err, map[string]string{
		"request_method": r.Method,
		"request_url":    r.URL.String(),
	})

}

func (app *application) errResp(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelop{"error": message}
	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.logErr(r, err)
		w.WriteHeader(500)
	}
}
func (app *application) serverErrResp(w http.ResponseWriter, r *http.Request, err error) {
	message := "serverErr from Server responce invoker"
	app.logger.PrintError(err, nil)

	app.errResp(w, r, http.StatusInternalServerError, message)
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errResp(w, r, http.StatusNotFound, message)
}

func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	app.errResp(w, r, http.StatusMethodNotAllowed, message)
}
func (app *application) appBadResp(w http.ResponseWriter, r *http.Request, err error) {
	//message := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
	//app.errResp(w, r, http.StatusMethodNotAllowed, message)
	app.errResp(w, r, http.StatusBadRequest, err.Error())
}

func (app *application) failValidationResponse(w http.ResponseWriter, r *http.Request, err map[string]string) {

	app.errResp(w, r, http.StatusUnprocessableEntity, err)
}
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
	message := "unable to update the record due to an edit conflict, please try again"
	app.errResp(w, r, http.StatusConflict, message)
}
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errResp(w, r, http.StatusTooManyRequests, message)
}
func (app *application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	message := "invalid authentication credentials"
	app.errResp(w, r, http.StatusUnauthorized, message)
}
func (app *application) invalidAuthenticationTokenResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")
	message := "invalid or missing authentication token"
	app.errResp(w, r, http.StatusUnauthorized, message)
}
func (app *application) authenticationRequiredResponse(w http.ResponseWriter, r *http.Request) {
	message := "you must be authenticated to access this resource"
	app.errResp(w, r, http.StatusUnauthorized, message)
}
func (app *application) inactiveAccountResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account must be activated to access this resource"
	app.errResp(w, r, http.StatusForbidden, message)
}
func (app *application) notPermittedResponse(w http.ResponseWriter, r *http.Request) {
	message := "your user account doesn't have the necessary permissions to access this resource"
	app.errResp(w, r, http.StatusForbidden, message)
}

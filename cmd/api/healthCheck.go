package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {

	env := envelop{
		"status": "available",
		"systemInfo": map[string]string{
			"env":     app.config.env,
			"version": version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

}

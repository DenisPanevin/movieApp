package main

import (
	"errors"
	"net/http"
	"secondBook/internal/data"
	"secondBook/internal/validator"
	"time"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.appBadResp(w, r, err)

		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {

		app.serverErrResp(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failValidationResponse(w, r, v.Errors)
		return
	}
	err = app.models.Users.Insert(user)
	if err != nil {
		println(err.Error())
		switch {

		case errors.Is(err, data.ErrDuplicateEmail):

			v.AddError("email", "a user with this email address already exists")
			app.failValidationResponse(w, r, v.Errors)
		default:

			app.serverErrResp(w, r, err)
		}
		return
	}
	token, err := app.models.Tokens.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		err = app.mailer.Send(user.Email, "User_welcome.tmpl", data)
		if err != nil {

			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

}
func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.appBadResp(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failValidationResponse(w, r, v.Errors)
		default:

			app.serverErrResp(w, r, err)
		}
		return
	}
	user.Activated = true
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrResp(w, r, err)
		}
		return
	}
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrResp(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"user": user}, nil)
	if err != nil {
		app.serverErrResp(w, r, err)
	}

}

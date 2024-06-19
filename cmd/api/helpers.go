package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"secondBook/internal/validator"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type envelop map[string]any

func (app *application) retriveId(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {

		return 0, errors.New("invalid ID params")
	}
	return id, nil
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelop, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil

}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder, and call the DisallowUnknownFields() method on it
	// before decoding. This means that if the JSON from the client now includes any
	// field which cannot be mapped to the target destination, the decoder will return
	// an error instead of just ignoring the field.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dest)

	if err != nil {

		// If there is an error during decoding, start the triage...
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError
		switch {
		// Use the errors.As() function to check whether the error has the type
		// *json.SyntaxError. If it does, then return a plain-english error message
		// which includes the location of the problem.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			// In some circumstances Decode() may also return an io.ErrUnexpectedEOF error
			// for syntax errors in the JSON. So we check for this using errors.Is() and
			// return a generic error message. There is an open issue regarding this at
			// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
			// Likewise, catch any *json.UnmarshalTypeError errors. These occur when the
			// JSON value is the wrong type for the target destination. If the error relates
			// to a specific field, then we include that in our error message to make it
			// easier for the client to debug.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)
			// An io.EOF error will be returned by Decode() if the request body is empty. We
			// check for this with errors.Is() and return a plain-english error message
			// instead.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
			// A json.InvalidUnmarshalError error will be returned if we pass something
			// that is not a non-nil pointer to Decode(). We catch this and panic,
			// rather than returning an error to our handler. At the end of this chapter
			// we'll talk about panicking versus returning errors, and discuss why it's an
			// appropriate thing to do in this specific situation.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)
			// Use the errors.As() function to check whether the error has the type
			// *http.MaxBytesError. If it does, then it means the request body exceeded our
			// size limit of 1MB and we return a clear error message.
		case errors.As(err, &maxBytesError):

			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)
		case errors.As(err, &invalidUnmarshalError):

			panic(err)
			// For anything else, return the error message as-is.
		default:

			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil

}

func (app *application) readString(qs url.Values, ks string, ds string) string {
	s := qs.Get(ks)
	if s == "" {
		return ds
	}
	return s
}
func (app *application) readCsv(qs url.Values, ks string, ds []string) []string {
	csv := qs.Get(ks)
	if csv == "" {
		return ds
	}
	return strings.Split(csv, ",")
}
func (app *application) readInt(qs url.Values, ks string, di int, v *validator.Validator) int {
	s := qs.Get(ks)

	if s == "" {
		return di
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(ks, "must be integer")
		return di
	}

	return i

}
func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {

		defer app.wg.Done()
		defer func() {

			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()

		fn()
	}()
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type envelope map[string]any

// writeJSON() helper for sending responses. This takes the destination
// http.ResponseWriter, the HTTP status code, the data to encode, and a
// header map containing any additional HTTP headers we want to include in the response.
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode to JSON, return any errors, if any.
	// In benchmarks json.MarshalIndent() takes 65% longer to run and uses around 30% more memory
	// than json.Marshal(), as well as making more heap allocations.
	// We’re talking about a few thousandths of a millisecond. If your service is operating in
	// a very resource-constrained environment, then this is worth being aware of.
	json, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// Append new line for easier reading in terminals.
	json = append(json, '\n')

	// Loop through the header map and add each header to the http.ResponseWriter header map.
	// Go doesn't throw an error if you try to range over a nil map.
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add the "Content-Type: application/json" header, then write the status code and
	// JSON response.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Use http.MaxBytesReader() to limit the size of the request body to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Init json.Decoder, and call the DisallowUnknownFields() method on it before
	// calling Decode(). If the JSON from the client now includes any field which
	// cannot be mapped to the target destination, the decoder will return an error.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body into the target destination
	err := dec.Decode(dst)
	if err != nil {
		// If err, start triage
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// Use errors.As() to check if the error is a *json.SyntaxError type.
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains badly-formed JSON at character %d", syntaxError.Offset)

		// Decode() may also return an *json.UnmarshalTypeError error.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		// *json.UnmarshalTypeError errors occur when the JSON value is the wrong type for the
		// target destination.
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type at character %d", unmarshalTypeError.Offset)

		// io.EOF errors occur if the request body is empty.
		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")

		// If the JSON contains a field which cannot be mapped to the target destination
		// then Decode() will now return an error message in the format "json: unknown
		// field "<name>"". We check for this, extract the field name from the error,
		// and interpolate it into our custom error message.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// Use the errors.As() function to check whether the error has the type
		// *http.MaxBytesError. If it does, then it means the request body exceeded our
		// size limit of 1MB and we return a clear error message.
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		// json.InvalidUnmarshalError occurs when you pass a non-nil pointer to Decode(). We catch this
		// and panic, rather than returning an error.
		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anonymous struct. If the request
	// body only contained a single JSON value this will return an io.EOF error. So if we
	// get anything else, we know that there is additional data in the request body and
	// we return our own custom error message.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// Read snip ID URL param.
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// PathValue() is new for Go 1.22 and allows us to read URL params.
	// We also convert the string to base 10 integer with a 64 bit size.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

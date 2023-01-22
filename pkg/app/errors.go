package app

// GeneralRestError is the error returned by the Mastodon API
type GeneralRestError struct {
	// ErrorInstanceID is a unique identifier for this error instance; useful for error log cross-referencing
	ErrorInstanceID string `json:"error_instance_id"`

	// ErrorMessage is the error message returned to the user
	ErrorMessage string `json:"error_message"`
}

// NoDB is returned when the no database is provided
type NoDB struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoDB) Error() string {
	if e.Msg == "" {
		e.Msg = "no database instnace provided; use WithDB() to set one"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

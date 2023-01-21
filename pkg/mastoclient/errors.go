package mastoclient

// NoAccessTokenError error
type NoAccessTokenError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoAccessTokenError) Error() string {
	if e.Msg == "" {
		e.Msg = "No access token. use WithAccessToken()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoClientKeyError error
type NoClientKeyError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoClientKeyError) Error() string {
	if e.Msg == "" {
		e.Msg = "no client key. use WithClientKey()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoClientSecretError error
type NoClientSecretError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoClientSecretError) Error() string {
	if e.Msg == "" {
		e.Msg = "no client secret. use WithClientSecret()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// NoInstanceError error
type NoInstanceError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *NoInstanceError) Error() string {
	if e.Msg == "" {
		e.Msg = "no instance. use WithInstance()"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

// PostFailedError error
type PostFailedError struct {
	Err error
	Msg string
}

// Error returns the error message
func (e *PostFailedError) Error() string {
	if e.Msg == "" {
		e.Msg = "post failed"
	}
	if e.Err != nil {
		e.Msg += ": " + e.Err.Error()
	}
	return e.Msg
}

package siridb

// Error can be returned by the siridb package.
type Error struct {
	msg string
	tp  uint8
}

// NewError returns a pointer to a new Error.
func NewError(s string, t uint8) *Error {
	return &Error{
		msg: s,
		tp:  t,
	}
}

// Error returns the error msg.
func (e *Error) Error() string { return e.msg }

// Type returns the error type.
func (e *Error) Type() uint8 { return e.tp }

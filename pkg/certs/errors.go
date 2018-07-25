package certs

// Define ErrZeroDivision
type CertError struct {
	message string
}

func NewCertError(message string) *CertError {
	return &CertError{
		message: message,
	}
}

func (e *CertError) Error() string {
	return e.message
}

// Define ErrBadHost
type ErrBadHost struct {
	message string
}

func NewErrBadHost(message string) *ErrBadHost {
	return &ErrBadHost{
		message: message,
	}
}

func (e *ErrBadHost) Error() string {
	return e.message
}

// Define ErrPrivateKey
type ErrPrivateKey struct {
	message string
}

func NewErrPrivateKey(message string) *ErrPrivateKey {
	return &ErrPrivateKey{
		message: message,
	}
}

func (e *ErrPrivateKey) Error() string {
	return e.message
}

// Define ErrFileWriteFail
type ErrFileWriteFail struct {
	message string
}

func NewErrFileWriteFail(message string) *ErrFileWriteFail {
	return &ErrFileWriteFail{
		message: message,
	}
}

func (e *ErrFileWriteFail) Error() string {
	return e.message
}

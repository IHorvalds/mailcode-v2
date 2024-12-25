package auth

type BasicAuthCredentials struct {
	User     string
	Password string
}

// NewBasicAuth returns a new BasicAuthCredentials with the given user and password.
func NewBasicAuth(user, password string) *BasicAuthCredentials {
	return &BasicAuthCredentials{
		User:     user,
		Password: password,
	}
}

// Repository for the BasicAuth
type BasicAuthRepository interface {
	CreateBasicAuthCredsTables() error
	SaveBasicAuthCreds(ba *BasicAuthCredentials) error
	DeleteBasicAuthCreds(user string) error
	GetBasicAuthCreds(user string) (*BasicAuthCredentials, error)
}

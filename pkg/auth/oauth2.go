package auth

import (
	"time"

	"github.com/markbates/goth"
)

type OAuth2Credentials struct {
	User         string
	AccessToken  string
	RefreshToken string
	ExpiryDate   time.Time
}

func NewOAuth2(user, accessToken, refreshToken string, expiryDate time.Time) *OAuth2Credentials {
	return &OAuth2Credentials{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiryDate:   expiryDate,
	}
}

func FromGothUser(u *goth.User) *OAuth2Credentials {
	return &OAuth2Credentials{
		User:         u.UserID,
		AccessToken:  u.AccessToken,
		RefreshToken: u.RefreshToken,
		ExpiryDate:   u.ExpiresAt,
	}
}

func ToGothUser(o *OAuth2Credentials) *goth.User {
	return &goth.User{
		UserID:       o.User,
		AccessToken:  o.AccessToken,
		RefreshToken: o.RefreshToken,
		ExpiresAt:    o.ExpiryDate,
	}
}

type OAuth2Repository interface {
	CreateOAuth2CredsTables() error
	SaveOAuth2Creds(ba *OAuth2Credentials) error
	DeleteOAuth2Creds(user string) error
	GetOAuth2Creds(user string) (*OAuth2Credentials, error)
}

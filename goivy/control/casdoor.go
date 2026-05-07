package control

import "context"

type CasdoorUser struct {
	ID            string
	Email         string
	DisplayName   string
	EmailVerified bool
	Disabled      bool
}

type CasdoorClient interface {
	CreateOrInviteUser(ctx context.Context, email, displayName string) (CasdoorUser, error)
	DisableUser(ctx context.Context, id string) error
	GetUserByEmail(ctx context.Context, email string) (CasdoorUser, error)
}

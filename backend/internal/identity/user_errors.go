package identity

import infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"

var (
	ErrUserNotFound = infraerrors.NotFound("USER_NOT_FOUND", "user not found")
	ErrEmailExists  = infraerrors.Conflict("EMAIL_EXISTS", "email already exists")
)

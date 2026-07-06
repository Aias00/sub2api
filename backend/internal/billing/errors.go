package billing

import infraerrors "github.com/Aias00/cloudbase/internal/pkg/errors"

var ErrInsufficientBalance = infraerrors.BadRequest("INSUFFICIENT_BALANCE", "insufficient balance")

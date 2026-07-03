package service

import billingctx "github.com/Aias00/cloudbase/internal/billing"

var ErrUsageBillingRequestIDRequired = billingctx.ErrUsageBillingRequestIDRequired
var ErrUsageBillingRequestConflict = billingctx.ErrUsageBillingRequestConflict

// UsageBillingCommand describes one billable request that must be applied at most once.
type UsageBillingCommand = billingctx.UsageBillingCommand

func HashUsageRequestPayload(payload []byte) string {
	return billingctx.HashUsageRequestPayload(payload)
}

// AccountQuotaState holds the post-increment quota state returned by the DB transaction.
// All values are post-update (i.e., already include the increment).
type AccountQuotaState = billingctx.AccountQuotaState

type UsageBillingApplyResult = billingctx.UsageBillingApplyResult

type UsageBillingRepository = billingctx.UsageBillingRepository

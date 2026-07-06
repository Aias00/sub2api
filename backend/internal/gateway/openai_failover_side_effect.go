package gateway

type OpenAIUpstreamFailoverSideEffectInput struct {
	Platform      string
	AccountID     int64
	AccountName   string
	StatusCode    int
	RequestID     string
	Message       string
	ResponseBody  []byte
	Decision      OpenAIUpstreamFailureDecision
	DefaultDetail string
}

type OpenAIUpstreamFailoverSideEffect struct {
	Platform               string
	AccountID              int64
	AccountName            string
	StatusCode             int
	RequestID              string
	Message                string
	Detail                 string
	ResponseBody           []byte
	RetryableOnSameAccount bool
}

func BuildOpenAIUpstreamFailoverSideEffect(input OpenAIUpstreamFailoverSideEffectInput) OpenAIUpstreamFailoverSideEffect {
	detail := input.Decision.Detail
	if detail == "" {
		detail = input.DefaultDetail
	}
	return OpenAIUpstreamFailoverSideEffect{
		Platform:               input.Platform,
		AccountID:              input.AccountID,
		AccountName:            input.AccountName,
		StatusCode:             input.StatusCode,
		RequestID:              input.RequestID,
		Message:                input.Message,
		Detail:                 detail,
		ResponseBody:           input.ResponseBody,
		RetryableOnSameAccount: input.Decision.RetryableOnSameAccount,
	}
}

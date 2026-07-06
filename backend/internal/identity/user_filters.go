package identity

// UserListFilters contains all filter options for listing users.
type UserListFilters struct {
	Status    string
	Role      string
	Search    string
	GroupName string

	// APIKeyGroupID filters users who own at least one non-soft-deleted API key
	// bound to this group (api_keys.group_id). 0 = no filter.
	APIKeyGroupID int64
	Attributes    map[int64]string

	// IncludeSubscriptions controls whether list queries should load active subscriptions.
	// nil means not specified.
	IncludeSubscriptions *bool

	// IncludeDeleted bypasses soft-delete filtering for admin audit/search paths.
	IncludeDeleted bool
}

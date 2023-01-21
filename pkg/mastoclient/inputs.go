package mastoclient

// AsyncGetAccountStatusesInput is a struct for getting statuses asynchronously
type AsyncGetAccountStatusesInput struct {
	ID      string
	SinceID *string
	Ch      chan AsyncStatuses
}

// AsyncGetFollowersInput is a struct for getting followers asynchronously
type AsyncGetFollowersInput struct {
	ID      string
	SinceID *string
	Ch      chan AsyncFollowers
}

// AsyncGetFollowingInput is a struct for getting followings asynchronously
type AsyncGetFollowingInput struct {
	ID      string
	SinceID *string
	Ch      chan AsyncFollowing
}

// AsyncGetNotificationsInput is a struct for getting notifications asynchronously
type AsyncGetNotificationsInput struct {
	MaxID   *string
	SinceID *string
	MinID   *string
	Ch      chan AsyncNotices
}

type RegisterAppInput struct {
	ClientName  string
	InstanceURL string
	RedirectURI string
	Scopes      []string
	Website     string
}

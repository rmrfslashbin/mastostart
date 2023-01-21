package mastoclient

import "github.com/mattn/go-mastodon"

// AsyncAccount is a struct for returning ollowers asynchronously
type AsyncFollowers struct {
	Followers  []*mastodon.Account
	Pagination *mastodon.Pagination
	Err        error
}

// AsyncAccount is a struct for returning followings asynchronously
type AsyncFollowing struct {
	Following  []*mastodon.Account
	Pagination *mastodon.Pagination
	Err        error
}

// AsyncAccount is a struct for returning notices asynchronously
type AsyncNotices struct {
	Notices []*mastodon.Notification
	Err     error
}

// AsyncStatuses is a struct for returning statuses asynchronously
type AsyncStatuses struct {
	Statuses []*mastodon.Status
	Err      error
}

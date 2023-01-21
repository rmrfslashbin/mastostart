package mastoclient

import (
	"context"
	"log"

	"github.com/mattn/go-mastodon"
)

// AsyncGetAccountStatuses gets account statuses asynchronously
func (c *Config) AsyncGetAccountStatuses(input *AsyncGetAccountStatusesInput) {
	// TODO: Something in the pagination is broken
	client, err := c.preflight()
	if err != nil {
		input.Ch <- AsyncStatuses{
			Statuses: nil,
			Err:      err,
		}
	}

	var pg mastodon.Pagination
	pg.SinceID = "0"
	if input.SinceID != nil {
		c.log.Debug().
			Str("since_id", *input.SinceID).
			Msg("setting statuses since_id")
		pg.SinceID = mastodon.ID(*input.SinceID)
	}
	total := 0
	for {
		statuses, err := client.GetAccountStatuses(context.Background(), mastodon.ID(input.ID), &pg)
		if err != nil {
			log.Fatal(err)
		}
		input.Ch <- AsyncStatuses{
			Statuses: statuses,
			Err:      nil,
		}
		total += len(statuses)
		if pg.MaxID == "" {
			c.log.Debug().
				Int("statuses", total).
				Msg("finished fetching statuses")
			break
		}
		c.log.Debug().
			Int("statuses", total).
			Msg("got statuses- witing 5 seconds")
		pg.SinceID = ""
		pg.MinID = ""
		//time.Sleep(5 * time.Second)
	}
	close(input.Ch)
}

// AsyncGetFollowers gets followers asynchronously
func (c *Config) AsyncGetFollowers(input *AsyncGetFollowersInput) {
	client, err := c.preflight()
	if err != nil {
		input.Ch <- AsyncFollowers{
			Followers: nil,
			Err:       err,
		}
	}

	var pg mastodon.Pagination
	pg.SinceID = "0"
	if input.SinceID != nil {
		c.log.Debug().
			Str("since_id", *input.SinceID).
			Msg("setting followers since_id")
		pg.SinceID = mastodon.ID(*input.SinceID)
	}
	total := 0
	for {
		fs, err := client.GetAccountFollowers(context.Background(), mastodon.ID(input.ID), &pg)
		if err != nil {
			input.Ch <- AsyncFollowers{
				Followers: nil,
				Err:       err,
			}
		}
		input.Ch <- AsyncFollowers{
			Followers:  fs,
			Pagination: &pg,
			Err:        nil,
		}
		c.log.Info().
			Str("max_id", string(pg.MaxID)).
			Str("since_id", string(pg.SinceID)).
			Str("min_id", string(pg.MinID)).
			Msg("got followers")
		total += len(fs)
		if pg.MaxID == "" {
			c.log.Debug().
				Int("followers", total).
				Str("for", input.ID).
				Msg("finished fetching followers")
			break
		}
		c.log.Info().
			Int("followers", total).
			Str("for", input.ID).
			Msg("got followers- witing 5 seconds")
		pg.SinceID = ""
		//time.Sleep(5 * time.Second)
	}
	// Send pagination info
	input.Ch <- AsyncFollowers{
		Followers:  nil,
		Pagination: &pg,
		Err:        nil,
	}
	close(input.Ch)
}

// AsyncGetFollowing gets following asynchronously
func (c *Config) AsyncGetFollowing(input *AsyncGetFollowingInput) {
	client, err := c.preflight()
	if err != nil {
		input.Ch <- AsyncFollowing{
			Following: nil,
			Err:       err,
		}
	}

	var pg mastodon.Pagination
	pg.SinceID = "0"
	if input.SinceID != nil {
		c.log.Debug().
			Str("since_id", *input.SinceID).
			Msg("setting following since_id")
		pg.SinceID = mastodon.ID(*input.SinceID)
	}
	total := 0
	for {
		fs, err := client.GetAccountFollowing(context.Background(), mastodon.ID(input.ID), &pg)
		if err != nil {
			log.Fatal(err)
		}
		input.Ch <- AsyncFollowing{
			Following:  fs,
			Pagination: &pg,
			Err:        nil,
		}
		total += len(fs)
		if pg.MaxID == "" {
			c.log.Debug().
				Int("following", total).
				Msg("finished fetching following")
			break
		}
		c.log.Info().
			Int("following", total).
			Msg("got following- witing 5 seconds")
		pg.SinceID = ""
		//time.Sleep(5 * time.Second)
	}
	// Send pagination info
	input.Ch <- AsyncFollowing{
		Following:  nil,
		Pagination: &pg,
		Err:        nil,
	}
	close(input.Ch)
}

// AsyncGetNotifications gets notifications asynchronously
func (c *Config) AsyncGetNotifications(input *AsyncGetNotificationsInput) {
	client, err := c.preflight()
	if err != nil {
		input.Ch <- AsyncNotices{
			Notices: nil,
			Err:     err,
		}
		close(input.Ch)
		return
	}

	var pg mastodon.Pagination
	pg.SinceID = "0"

	if input.MaxID != nil {
		c.log.Debug().
			Str("max_id", *input.MaxID).
			Msg("setting notifications max_id")
		pg.MaxID = mastodon.ID(*input.MaxID)
	}
	if input.SinceID != nil {
		c.log.Debug().
			Str("since_id", *input.SinceID).
			Msg("setting notifications since_id")
		pg.SinceID = mastodon.ID(*input.SinceID)
	}
	if input.MinID != nil {
		c.log.Debug().
			Str("min_id", *input.MinID).
			Msg("setting notifications min_id")
		pg.MinID = mastodon.ID(*input.MinID)
	}

	total := 0
	for {
		// Get notifications
		notices, err := client.GetNotifications(context.Background(), &pg)
		if err != nil {
			input.Ch <- AsyncNotices{
				Notices: nil,
				Err:     err,
			}
			close(input.Ch)
			return
		}
		input.Ch <- AsyncNotices{
			Notices: notices,
			Err:     nil,
		}
		total += len(notices)

		if pg.MinID == "" {
			c.log.Debug().
				Int("notifications", total).
				Msg("finished fetching notifications")
			break
		}
		c.log.Info().
			Int("notifications", total).
			Msg("got notifications- witing 5 seconds")

		pg.SinceID = ""
		pg.MinID = ""
		//time.Sleep(5 * time.Second)
	}
	close(input.Ch)
}

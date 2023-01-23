package database

// AppCredentials represents an app credentials item in the database.
type AppCredentials struct {
	/* Exaple return from Mastodon
	{
		"id": "13",
		"name": "Test Application 8421",
		"website": "https://myapp.example",
		"redirect_uri": "http://localhost:8421",
		"client_id": "sample_client_id",
		"client_secret": "sample_client_secret",
		"vapid_key": "sample_vapid_key"
	}
	*/

	InstanceURL  string `json:"instance_url"`
	ID           string `json:"id"`
	Name         string `json:"name"`
	Website      string `json:"website"`
	RedirectURI  string `json:"redirect_uri"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthURI      string `json:"vapid_key"`
}

// ConfigItem represents a config item in the database.
type ConfigItem struct {
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
}

// List represents a list item in the database.
type List struct {
	// ListID is the Mastodon (numeric) list ID.
	ListID string `json:"list_id"`

	// ListTitle is the title of the list.
	ListTitle string `json:"list_title"`

	// OwerUserID is the Mastodon (numeric) user ID of the list owner.
	OwerUserID string `json:"owner_user_id"`

	// PSK is the pre-shared key for the list.
	// This is used to share the list with other users.
	// Optional. One will be generated if not provided.
	PSK string `json:"psk"`

	// Public is a boolean indicating if the list is public or not.
	Public bool `json:"public"`
}

// ListMember represents a list member item in the database.
type ListMember struct {
	// ListID is the Mastodon (numeric) list ID.
	ListID string `json:"list_id"`

	// UserIDs are the Mastodon (numeric) user IDs of the list members.
	UserIDs []string `json:"user_id"`
}

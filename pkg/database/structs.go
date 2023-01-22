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

// UserCredentials represents a user credentials item in the database.
type UserCredentials struct {
	InstanceURL string `json:"instance_url"`
	UserID      string `json:"user_id"`
	Username    string `json:"username"`
	CreatedAt   string `json:"created_at"`
}

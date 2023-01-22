package app

import (
	"errors"
	"net/url"
	"strings"

	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
)

// createAppCreds creates an app on the instance and returns the credentials
func (cfg *Config) createAppCreds(instanceURL *url.URL) (*database.AppCredentials, error) {
	// Get redirect_uri from database
	redirectURI, err := cfg.db.GetConfig("redirect_uri")
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::cfg.db.GetConfig('redirect_uri')").
			Msg("error fetching 'redirect_uri' key/value pair from ddb")
		return nil, errors.New(guid.String() + ": error fetching 'redirect_uri' key/value pair from ddb")
	}

	if redirectURI == nil {
		guid := xid.New()
		cfg.log.Error().
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::redirectURI == nil").
			Msg("'redirect_uri' key/value pair is nil (not found in database). Maybe run setup?")
		return nil, errors.New(guid.String() + ": 'redirect_uri' key/value pair is nil (not found in database). Maybe run setup?")
	}

	// Get app_name from database
	appName, err := cfg.db.GetConfig("app_name")
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::cfg.db.GetConfig('app_name')").
			Msg("error fetching 'app_name' key/value pair from ddb")
		return nil, errors.New(guid.String() + ": error fetching 'app_name' key/value pair from ddb")
	}
	if appName == nil {
		guid := xid.New()
		cfg.log.Error().
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::appName == nil").
			Msg("'appName' key/value pair is nil (not found in database). Maybe run setup?")
		return nil, errors.New(guid.String() + ": 'appName' key/value pair is nil (not found in database). Maybe run setup?")
	}

	// Get website from database
	website, err := cfg.db.GetConfig("website")
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::cfg.db.GetConfig('website')").
			Msg("error fetching 'website' key/value pair from ddb")
		return nil, errors.New(guid.String() + ": error fetching 'website' key/value pair from ddb")
	}
	if website == nil {
		guid := xid.New()
		cfg.log.Error().
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::website == nil").
			Msg("'website' key/value pair is nil (not found in database). Maybe run setup?")
		return nil, errors.New(guid.String() + ": 'website' key/value pair is nil (not found in database). Maybe run setup?")
	}

	// Get website from database
	scopeConfig, err := cfg.db.GetConfig("scopes")
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::cfg.db.GetConfig('scopes')").
			Msg("error fetching 'scopes' key/value pair from ddb")
		return nil, errors.New(guid.String() + ": error fetching 'scopes' key/value pair from ddb")
	}
	if scopeConfig == nil {
		guid := xid.New()
		cfg.log.Error().
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::scopeConfig == nil").
			Msg("'scopeConfig' key/value pair is nil (not found in database). Maybe run setup?")
		return nil, errors.New(guid.String() + ": 'scopeConfig' key/value pair is nil (not found in database). Maybe run setup?")
	}

	// Split and clean up the scopes
	scopes := strings.Split(scopeConfig.ConfigValue, ",")
	for i, scope := range scopes {
		scopes[i] = strings.ToLower(strings.TrimSpace(scope))
	}

	// Construct the redirect URI
	redirectURIStr := redirectURI.ConfigValue + "?instance_url=" + instanceURL.String()

	// Register the app with the instance
	app, err := mastoclient.RegisterApp(&mastoclient.RegisterAppInput{
		ClientName:  appName.ConfigValue,
		InstanceURL: instanceURL.String(),
		RedirectURI: redirectURIStr,
		Scopes:      scopes,
		Website:     website.ConfigValue,
	})
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::mastoclient.RegisterApp()").
			Str("clientName", appName.ConfigValue).
			Str("instanceURL", instanceURL.String()).
			Str("redirectURI", redirectURIStr).
			Str("website", website.ConfigValue).
			Msg("error registering app")
		return nil, errors.New(guid.String() + ": error registering app")
	}

	newApp := &database.AppCredentials{
		InstanceURL:  instanceURL.Host,
		ID:           string(app.ID),
		Name:         appName.ConfigValue,
		Website:      website.ConfigValue,
		RedirectURI:  redirectURIStr,
		ClientID:     app.ClientID,
		ClientSecret: app.ClientSecret,
		AuthURI:      app.AuthURI,
	}

	// Save the app credentials in the database
	if err := cfg.db.PutAppCredentials(newApp); err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "createAppCreds::cfg.db.PutAppCredentials(newApp)").
			Msg("error putting app in ddb")
		return nil, errors.New(guid.String() + ": error putting app in ddb")
	}

	// Log success
	cfg.log.Info().
		Str("appID", string(app.ID)).
		Str("appName", appName.ConfigValue).
		Str("authURI", app.AuthURI).
		Str("clientID", app.ClientID).
		Str("instanceURL", instanceURL.String()).
		Str("redirectURI", redirectURIStr).
		Str("website", website.ConfigValue).
		Str("scopes", strings.Join(scopes, ",")).
		Msg("created mastodon app credentials")

	// Return the app credentials
	return newApp, nil
}

package app

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// apiInstanceInfo returns information about the Mastodon instance
func (cfg *Config) apiInstanceInfo(c *fiber.Ctx) error {
	flight, err := cfg.preflight(
		&PreflightInput{
			jwtToken: c.Locals("user").(*jwt.Token),
		},
	)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiInstanceInfo::cfg.preflight()").
			Msg("prefilight failed")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	instance, err := flight.Client.GetInstanceInfo()
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiInstanceInfo::flight.Client.GetInstanceInfo()").
			Str("UserID", string(*flight.Userid)).
			Msg("failed to get instance info")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	stats, err := flight.Client.GetInstanceStats()
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiInstanceInfo::flight.Client.GetInstanceStats()").
			Str("UserID", string(*flight.Userid)).
			Msg("failed to get instance stats")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	type output struct {
		Instance *mastodon.Instance         `json:"instance"`
		Stats    []*mastodon.WeeklyActivity `json:"stats"`
	}

	return c.JSON(&output{
		Instance: instance,
		Stats:    stats,
	})
}

func (cfg *Config) preflight(in *PreflightInput) (*PreflightOutput, error) {
	output := &PreflightOutput{}

	// Get the JWT claims
	claims := in.jwtToken.Claims.(jwt.MapClaims)

	// Get user's Mastodon access token from the JWT claims
	accessToken := claims["access_token"].(string)

	// Subject is the fully qualified URL to the user's account
	subject := claims["sub"].(string)
	output.FQUsername = &subject

	// userid is the user's numeric ID in the Mastodon instance
	userid := mastodon.ID(claims["jti"].(string))
	output.Userid = &userid

	// subjectURL is a fully qualified URL to the user's account
	subjectURL, err := url.Parse(subject)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "preflight::url.Parse(subject)").
			Str("subject", subject).
			Msg("unable to parse subject as URL from JWT claims")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return nil, errors.New(string(e))
	}

	// username is the user's username in the Mastodon instance.
	// Most Mastodon API calls require the UserID and not the username
	// but we can parse it out and use it if needed.
	username := strings.TrimPrefix(subjectURL.Path, "/@")
	output.Username = &username

	// Construct the instance URL
	instanceURL := "https://" + subjectURL.Host
	output.InstanceURL = &instanceURL

	// Get the app credentials from the database
	appCreds, err := cfg.db.GetAppCredentials(subjectURL.Host)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "preflight::cfg.db.GetAppCredentials(subjectURL.Host)").
			Str("instanceURL", subjectURL.Host).
			Msg("unable to get app credentials from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return nil, errors.New(string(e))
	}

	if appCreds == nil {
		guid := xid.New()
		log.Error().
			Str("errRef", guid.String()).
			Str("function", "preflight::cfg.db.GetAppCredentials(subjectURL.Host)").
			Str("instanceURL", subjectURL.Host).
			Msg("unable to get app credentials from database: appCreds is nil")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return nil, errors.New(string(e))
	}

	// Create a new mastoclient instance
	mc, err := mastoclient.New(
		mastoclient.WithInstance(&instanceURL),               // Mastodon instance URL from the JWT subject claim
		mastoclient.WithClientkey(&appCreds.ClientID),        // Mastodon app client ID from the database
		mastoclient.WithClientSecret(&appCreds.ClientSecret), // Mastodon app client secret from the database
		mastoclient.WithAccessToken(&accessToken),            // Mastodon user access token from the JWT claims
		mastoclient.WithLogger(cfg.log),                      // You know, for logging
	)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("errRef", guid.String()).
			Str("function", "preflight::mastoclient.New()").
			Str("instanceURL", instanceURL).
			Str("ClientID", appCreds.ClientID).
			Str("ClientSecret", appCreds.ClientSecret).
			Str("errRef", guid.String()).
			Msg("Unable to create mastoclient")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return nil, errors.New(string(e))
	}
	output.Client = mc

	return output, nil
}

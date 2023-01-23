package app

import (
	"encoding/json"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// authVerify is the handler for the /auth/verify endpoint
func (cfg *Config) authVerify(c *fiber.Ctx) error {
	// This function is a PoC to show how to grab the user's data from the JWT
	// and then transact on the Mastodon instance with the user's access token.
	// Most of this code is consolidated into the preflight() function.

	// Get the jwtToken from the JWT
	jwtToken := c.Locals("user").(*jwt.Token)

	// Get the JWT claims
	claims := jwtToken.Claims.(jwt.MapClaims)

	// Get user's Mastodon access token from the JWT claims
	accessToken := claims["access_token"].(string)

	// Subject is the fully qualified URL to the user's account
	subject := claims["sub"].(string)

	// userid is the user's numeric ID in the Mastodon instance
	userid := mastodon.ID(claims["jti"].(string))

	// subjectURL is a fully qualified URL to the user's account
	subjectURL, err := url.Parse(subject)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::url.Parse(subject)").
			Str("subject", subject).
			Msg("unable to parse subject as URL from JWT claims")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// username is the user's username in the Mastodon instance.
	// Most Mastodon API calls require the UserID and not the username
	// but we can parse it out and use it if needed.
	//username := strings.TrimPrefix(subjectURL.Path, "/@")

	// Construct the instance URL
	instanceURL := "https://" + subjectURL.Host

	// Get the app credentials from the database
	appCreds, err := cfg.db.GetAppCredentials(subjectURL.Host)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::cfg.db.GetAppCredentials(subjectURL.Host)").
			Str("instanceURL", subjectURL.Host).
			Msg("unable to get app credentials from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	if appCreds == nil {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::cfg.db.GetAppCredentials(subjectURL.Host)").
			Str("instanceURL", subjectURL.Host).
			Msg("unable to get app credentials from database: appCreds is nil")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
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
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::mastoclient.New()").
			Str("instanceURL", instanceURL).
			Str("ClientID", appCreds.ClientID).
			Str("ClientSecret", appCreds.ClientSecret).
			Str("errRef", guid.String()).
			Msg("Unable to create mastoclient")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Get the user's Mastodon profile.
	// The Me() funtion assumes the identity of the user based on the access token
	me, err := mc.Me()
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::mc.me()").
			Msg("Unable to get user details from mastodon")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "unable to fetch user profile. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrPreconditionRequired.Code).SendString(string(e))
	}

	// Get the user's last status from Mastodon
	lastStatus, err := mc.GetLastStatus(&userid)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::mc.GetLastStatus(&userid)").
			Msg("Unable to get user's last status from Mastodon")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "unable to fetch user's last status from Mastodon. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrPreconditionRequired.Code).SendString(string(e))
	}

	// Return the user's profile and last status
	return c.JSON(&AuthVerifyReturn{
		Account:    me,
		LastStatus: lastStatus,
	})
}

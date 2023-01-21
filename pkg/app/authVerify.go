package app

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func (cfg *Config) authVerify(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	accessToken := claims["access_token"].(string)
	subject := claims["sub"].(string)
	//userid := claims["jti"].(string)

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

	//username := strings.TrimPrefix(subjectURL.Path, "/@")
	instanceURL := fmt.Sprintf("https://%s", subjectURL.Host)

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
	mastodon, err := mastoclient.New(
		mastoclient.WithInstance(&instanceURL),
		mastoclient.WithClientkey(&appCreds.ClientID),
		mastoclient.WithClientSecret(&appCreds.ClientSecret),
		mastoclient.WithAccessToken(&accessToken),
		mastoclient.WithLogger(cfg.log),
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

	me, err := mastodon.Me()
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authVerify::mastodon.me()").
			Msg("Unable to get user details from mastodon")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "unable to fetch user profile. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrPreconditionRequired.Code).SendString(string(e))
	}

	return c.JSON(&me)
}

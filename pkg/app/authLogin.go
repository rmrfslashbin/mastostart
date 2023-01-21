package app

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func (cfg *Config) authLogin(c *fiber.Ctx) error {
	// get the username and instance URL from the query params
	username := c.Query("username")
	if username == "" {
		guid := xid.New()
		cfg.log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authLogin::c.Query('username')").
			Msg("missing 'username' query param")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "missing 'username' query param",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	rawInstanceURL := c.Query("instance_url")
	if rawInstanceURL == "" {
		guid := xid.New()
		cfg.log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authLogin::c.Query('instance_url')").
			Msg("missing 'instance_url' query param")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "missing 'instance_url' query param",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	instanceURL, err := url.Parse(rawInstanceURL)
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authLogin::url.Parse(rawInstanceURL)").
			Msg("error parsing instance_url")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "unable to parse instance_url",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	// Get the instance permit list
	permitInstances, err := cfg.db.GetConfig("permit_instances")
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::cfg.db.GetConfig('permit_instances')").
			Str("errRef", guid.String()).
			Msg("Unable get permit_instances from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Check if the instance is in the permit list
	if permitInstances != nil {
		permitList := make(map[string]struct{})
		for _, instance := range strings.Split(permitInstances.ConfigValue, ",") {
			permitList[strings.ToLower(strings.TrimSpace(instance))] = struct{}{}
		}

		if len(permitList) > 0 {
			if _, ok := permitList[strings.ToLower(instanceURL.Host)]; !ok {
				guid := xid.New()
				cfg.log.Error().
					Str("method", c.Method()).
					Str("originalURL", c.OriginalURL()).
					Str("errRef", guid.String()).
					Str("function", "authLogin::CheckPermitInstanceList").
					Str("instanceURL", instanceURL.Host).
					Msg("instance not in permit list")
				e, _ := json.Marshal(&GeneralRestError{
					ErrorInstanceID: guid.String(),
					ErrorMessage:    "instance not in permit list",
				})
				return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
			}
		}
	}

	// Get/Setup App credentials
	var appCreds *database.AppCredentials
	appCreds, err = cfg.db.GetAppCredentials(instanceURL.Host)
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authLogin::cfg.db.GetAppCredentials(instanceURL.Host)").
			Msg("error fetching app creds from ddb")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// If app creds don't exist, create them
	if appCreds == nil {
		createdAppCreds, appCredsErr := cfg.createAppCreds(instanceURL)
		if appCredsErr != nil {
			guid := xid.New()
			cfg.log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("errRef", guid.String()).
				Str("function", "authLogin::cfg.createAppCreds(instanceURL.Host)").
				Msg("error creating app creds")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}
		appCreds = createdAppCreds
	}

	// Return the signed JWT
	return c.JSON(fiber.Map{"authuri": appCreds.AuthURI})

}

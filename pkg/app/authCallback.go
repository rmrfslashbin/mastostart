package app

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// authCallback is the handler for the /auth/callback endpoint
func (cfg *Config) authCallback(c *fiber.Ctx) error {
	// Fetch the code query param
	code := c.Query("code")
	if code == "" {
		guid := xid.New()
		cfg.log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::c.Query('code')").
			Msg("missing 'code' query param")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "missing 'code' query param",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	// Fetch the instance_url query param
	rawInstanceURL := c.Query("instance_url")
	if rawInstanceURL == "" {
		guid := xid.New()
		cfg.log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::c.Query('instance_url')").
			Msg("missing 'instance_url' query param")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "missing 'instance_url' query param",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	// Parse the instance_url
	instanceURL, err := url.Parse(rawInstanceURL)
	if err != nil {
		guid := xid.New()
		cfg.log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::url.Parse(rawInstanceURL)").
			Msg("error parsing instance_url")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "unable to parse instance_url",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	// Get the app name
	appNameConfig, err := cfg.db.GetConfig("app_name")
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::cfg.db.GetConfig('app_name')").
			Str("errRef", guid.String()).
			Msg("Unable get app_name from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Set a defailt app name if it's not set
	appName := strings.TrimSpace(appNameConfig.ConfigValue)
	if appName == "" {
		appName = "mastostart"
	}

	permitted, err := cfg.checkPermitInstanceList(instanceURL)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::cfg.checkPermitInstanceList(instanceURL)").
			Str("errRef", guid.String()).
			Msg("Unable get do permit instance list check")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	if !*permitted {
		guid := xid.New()
		cfg.log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::CheckPermitInstanceList").
			Str("instanceURL", instanceURL.Host).
			Msg("instance not in permit list")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "instance not in permit list",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	// Get the app credentials from the database
	appCreds, err := cfg.db.GetAppCredentials(instanceURL.Host)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::cfg.db.GetAppCredentials()").
			Msg("unable to get app credentials from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// If no app is set up, someone is doing something they shouldn't
	if appCreds == nil {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::cfg.db.GetAppCredentials()").
			Msg("unable to get app credentials from database: appCreds is nil")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Get the full URL of the instance
	instanceUrlStr := instanceURL.String()

	// Create a new mastoclient instance
	mastodon, err := mastoclient.New(
		mastoclient.WithAccessToken(nil), // access client is not needed for this step
		mastoclient.WithClientkey(&appCreds.ClientID),
		mastoclient.WithClientSecret(&appCreds.ClientSecret),
		mastoclient.WithInstance(&instanceUrlStr),
		mastoclient.WithLogger(cfg.log),
	)

	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::mastoclient.New()").
			Str("instanceURL", instanceURL.Host).
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

	// Using the OAuth2 code, get the access token
	accessToken, err := mastodon.GetAuthTokenFromCode(&code, &appCreds.RedirectURI)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::mastodon.GetAuthTokenFromCode()").
			Str("code", code).
			Str("redirect_uri", appCreds.RedirectURI).
			Str("errRef", guid.String()).
			Msg("Unable to get access token")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	if accessToken == nil {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::mastodon.GetAuthTokenFromCode()").
			Str("code", code).
			Str("redirect_uri", appCreds.RedirectURI).
			Str("errRef", guid.String()).
			Msg("Unable to get access token: accessToken is nil")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Get the user's profile from Mastodon using their access token
	mastodon.SetAccessToken(accessToken)
	me, err := mastodon.Me()
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "authCallback::mastodon.me()").
			Msg("Unable to get user details from mastodon")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Get RSA private key for signing JWT
	privateKey, err := cfg.getRSAPrivateKey()
	if err != nil {
		if err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("function", "authCallback::cfg.getRSAPrivateKey()").
				Str("errRef", guid.String()).
				Msg("Unable fetch RSA private key")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}
	}

	// Create the JWT claims
	claims := JWTClaims{
		*accessToken, // Encode the user's Mastodon access token
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 1 week
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    appName,
			Subject:   me.URL,        // Fully qualified URL representing the user
			ID:        string(me.ID), // Mastodon (numeric) user ID
		},
	}

	// Create the JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedJWT, err := token.SignedString(privateKey)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::token.SignedString(privateKey)").
			Str("errRef", guid.String()).
			Msg("Unable to sign JWT with RSA private key")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Return the signed JWT
	return c.JSON(
		fiber.Map{
			"token": signedJWT,
			"type":  "Bearer",
		},
	)
}

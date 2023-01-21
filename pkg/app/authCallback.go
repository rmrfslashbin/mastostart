package app

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func (cfg *Config) authCallback(c *fiber.Ctx) error {
	// Fetch the code and instance_url query params
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

	instanceUrlStr := instanceURL.String()

	// Create a new mastoclient instance
	mastodon, err := mastoclient.New(
		mastoclient.WithInstance(&instanceUrlStr),
		mastoclient.WithClientkey(&appCreds.ClientID),
		mastoclient.WithClientSecret(&appCreds.ClientSecret),
		mastoclient.WithAccessToken(nil), // access client is not needed for this step
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

	// Get the access token from Mastodon
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

	/*
		// Store the user's credentials in the database
		createdAt := time.Now().UTC().Format(time.RFC3339)
		if err := cfg.db.PutUserCredentials(&database.UserCredentials{
			InstanceURL: instanceURL,
			UserID:      string(me.ID),
			Username:    me.Username,
			CreatedAt:   createdAt,
		}); err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("function", "authRedirect::cfg.db.PutUserCredentials()").
				Str("instanceURL", instanceURL).
				Str("userID", string(me.ID)).
				Str("accessToken", *accessToken).
				Str("username", me.Username).
				Str("createdAt", createdAt).
				Str("errRef", guid.String()).
				Msg("Unable to put user credentials")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}
	*/

	// Get the PEM encoded JWT RSA private signing key from the database
	jwtSingingKeyEncoded, err := cfg.db.GetConfig("jwt_signing_key")
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::cfg.db.GetConfig('jwt_signing_key')").
			Str("errRef", guid.String()).
			Msg("Unable get jwt_signing_key from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}
	if jwtSingingKeyEncoded == nil {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::if jwtSingingKeyEncoded == nil").
			Str("errRef", guid.String()).
			Msg("jwt_signing_key from database is nil")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Decode the PEM formatted RSA private signing key
	block, _ := pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::pem.Decode([]byte(jwtSingingKeyEncoded.ConfigValue))").
			Str("errRef", guid.String()).
			Msg("Unable to decode jwt_signing_key PEM from database")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	// Parse the PEM encoded RSA private signing key
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("function", "authCallback::x509.ParsePKCS1PrivateKey(block.Bytes)").
			Str("errRef", guid.String()).
			Msg("Unable to parse PEM to RSA private key")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	//userURL := me.URL
	//userID := me.ID
	//instanceURL

	// Create the JWT claims
	claims := JWTClaims{
		*accessToken,
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * 7 * time.Hour)), // 1 week
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "mastostart",
			Subject:   me.URL,
			ID:        string(me.ID),
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
	return c.JSON(fiber.Map{"token": signedJWT})
}

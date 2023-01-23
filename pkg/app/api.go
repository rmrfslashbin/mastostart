package app

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mattn/go-mastodon"
	"github.com/rmrfslashbin/mastostart/pkg/database"
	"github.com/rmrfslashbin/mastostart/pkg/mastoclient"
	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

func (cfg *Config) apiAccountsInList(c *fiber.Ctx) error {
	listID := mastodon.ID(strings.TrimSpace(c.Params("listID")))

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
			Str("function", "apiAccountsInList::cfg.preflight()").
			Msg("prefilight failed")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	listSlice, err := flight.Client.MyLists(&listID)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiMyLists::flight.Client.MyLists(&listID)").
			Str("listID", string(listID)).
			Str("UserID", string(*flight.Userid)).
			Msg("failed to get list")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	if len(listSlice) == 0 {
		guid := xid.New()
		log.Error().
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiMyLists::len(list) == 0").
			Str("listID", string(listID)).
			Str("UserID", string(*flight.Userid)).
			Msg("fetched list but lenght is 0")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "no list found (or unable to access) with that id",
		})
		return c.Status(fiber.ErrBadRequest.Code).SendString(string(e))
	}

	accounts, err := flight.Client.GetAccountsInList(&listID)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("listID", string(listID)).
			Str("function", "apiAccountsInList::flight.Client.GetAccountsInList(&listID)").
			Msg("unable to get list of accounts in list")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	list := listSlice[0]

	saved := false
	public := false
	psk := ""
	if strings.ToLower(c.Query("save", "false")) == "true" {
		saved = true
		if strings.ToLower(c.Query("public", "false")) == "true" {
			public = true
		}

		randLen := 48
		b := make([]byte, randLen)
		_, err := rand.Read(b)
		if err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("errRef", guid.String()).
				Str("function", "apiAccountsInList::rand.Read(b)").
				Msg("failed getting random bytes for PSK")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}

		psk = base64.StdEncoding.EncodeToString(b)[0:32]
		instanceURL, _ := url.Parse(*flight.InstanceURL)

		if err = cfg.db.PutList(&database.List{
			Instance:    instanceURL.Host,
			ListID:      string(listID),
			ListTitle:   list.Title,
			OwnerUserID: string(*flight.Userid),
			Public:      public,
			PSK:         psk,
		}); err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("errRef", guid.String()).
				Str("listID", string(listID)).
				Str("owenerUserID", string(*flight.Userid)).
				Str("listTitle", list.Title).
				Str("function", "apiAccountsInList::cfg.db.PutList()").
				Msg("failed to save list to database")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}

		userIDs := make([]string, len(accounts))
		for i, account := range accounts {
			userIDs[i] = string(account.ID)
		}

		if err = cfg.db.PutAccountsInList(&database.ListMember{
			ListID:  string(listID),
			UserIDs: userIDs,
		}); err != nil {
			guid := xid.New()
			log.Error().
				Err(err).
				Str("method", c.Method()).
				Str("originalURL", c.OriginalURL()).
				Str("errRef", guid.String()).
				Str("listID", string(listID)).
				Str("owenerUserID", string(*flight.Userid)).
				Str("listTitle", list.Title).
				Str("function", "apiAccountsInList::cfg.db.PutAccountsInList()").
				Msg("failed to save list to database")
			e, _ := json.Marshal(&GeneralRestError{
				ErrorInstanceID: guid.String(),
				ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
			})
			return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
		}
	}

	return c.JSON(fiber.Map{
		"saved":    saved,
		"public":   public,
		"listID":   string(listID),
		"listName": list.Title,
		"ownerID":  string(*flight.Userid),
		"psk":      psk,
		"accounts": accounts,
	})
}

// apiMyLists is the handler for the /api/myLists endpoint
func (cfg *Config) apiMyLists(c *fiber.Ctx) error {
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
			Str("function", "apiMyLists::cfg.preflight()").
			Msg("prefilight failed")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}

	lists, err := flight.Client.MyLists(nil)
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("method", c.Method()).
			Str("originalURL", c.OriginalURL()).
			Str("errRef", guid.String()).
			Str("function", "apiMyLists::flight.Client.MyLists(nil)").
			Msg("unable to get list of lists")
		e, _ := json.Marshal(&GeneralRestError{
			ErrorInstanceID: guid.String(),
			ErrorMessage:    "server side failure. please report the error_instance_id to the admin",
		})
		return c.Status(fiber.ErrInternalServerError.Code).SendString(string(e))
	}
	return c.JSON(lists)
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

package app

import (
	"errors"
	"net/url"
	"strings"

	"github.com/rs/xid"
	"github.com/rs/zerolog/log"
)

// checkPermitInstanceList checks if the instance is in the permit list
func (cfg *Config) checkPermitInstanceList(instanceURL *url.URL) (*bool, error) {
	var permitted bool

	// Get the instance permit list
	permitInstances, err := cfg.db.GetConfig("permit_instances")
	// Fail if there's an error- this doesn't mean the instance isn't permitted, it means we can't check
	if err != nil {
		guid := xid.New()
		log.Error().
			Err(err).
			Str("function", "checkPermitInstanceList::cfg.db.GetConfig('permit_instances')").
			Str("errRef", guid.String()).
			Msg("unable get permit_instances from database")
		return nil, errors.New(guid.String() + ": unable get permit_instances from database")
	}

	// Default to permitted
	permitted = true

	// Check if the instance is in the permit list
	if permitInstances != nil {
		// Create a map of the instances in the permit list
		permitList := make(map[string]struct{})

		// Normalize the instance name to lowercase and trimp the spaces
		for _, instance := range strings.Split(permitInstances.ConfigValue, ",") {
			permitList[strings.ToLower(strings.TrimSpace(instance))] = struct{}{}
		}

		// If there is a permit list, check if the instance is in the list
		if len(permitList) > 0 {
			if _, ok := permitList[strings.ToLower(instanceURL.Host)]; !ok {
				// Permit list exists, but the instance isn't in the list
				permitted = false
				return &permitted, nil
			}
		}
	}

	// Instance is on the permit list --or-- no permit list exists, allow all instances
	return &permitted, nil
}

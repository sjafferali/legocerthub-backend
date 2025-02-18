package challenges

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"time"
)

var (
	errChallengeRetriesExhausted = errors.New("challenge failed (out of retries)")
	errChallengeTypeNotFound     = errors.New("intended challenge type not found")
)

// Solve accepts a slice of challenges from an authorization and solves the specific challenge
// specified by the method. Valid or invalid status is returned.  An error is returned if can't resolve
// a valid or invalid state.
func (service *Service) Solve(identifier acme.Identifier, challenges []acme.Challenge, method Method, key acme.AccountKey, isStaging bool) (status string, err error) {
	var challenge acme.Challenge
	found := false

	// range to the correct challenge to solve based on Type
	for i := range challenges {
		if challenges[i].Type == method.ChallengeType {
			found = true
			challenge = challenges[i]
		}
	}
	if !found {
		return "", errChallengeTypeNotFound
	}

	// provision the needed resource for validation and defer deprovisioning
	err = service.Provision(identifier, method, key, challenge.Token)
	// do error check after Deprovision to ensure any records that were created
	// get cleaned up, even if Provisioning errored.

	defer func() {
		err := service.Deprovision(identifier, method, key, challenge.Token)
		if err != nil {
			service.logger.Error(err)
		}
	}()

	// Provision error check
	if err != nil {
		return "", err
	}

	// Below this point is to inform ACME the challenge is ready to be validated
	// by the server and to subsequently monitor the challenge to be moved to the
	// valid or invalid state.

	// make pointer for the correct acme.Service (to avoid repeat of if/else)
	var acmeService *acme.Service
	if isStaging {
		acmeService = service.acmeStaging
	} else {
		acmeService = service.acmeProd
	}

	// inform ACME that the challenge is ready
	_, err = acmeService.ValidateChallenge(challenge.Url, key)
	if err != nil {
		return "", err
	}

	// monitor for processing to complete (max 5 tries, 20 seconds apart each)
	for i := 1; i <= 5; i++ {
		// sleep to allow ACME time to process
		// cancel/error if shutdown is called
		select {
		case <-service.shutdownContext.Done():
			// cancel/error if shutting down
			return "", errors.New("cloudflare dns provisioning canceled due to shutdown")

		case <-time.After(20 * time.Second):
			// sleep and retry
		}

		// get challenge and check for error or final Statuses
		challenge, err = acmeService.GetChallenge(challenge.Url, key)
		if err != nil {
			return "", err
		}

		// return Status if it has reached a final status
		if challenge.Status == "valid" {
			return challenge.Status, nil
		} else if challenge.Status == "invalid" {
			service.logger.Debug(challenge.Error)
			return challenge.Status, nil
		}
		// else repeat loop
	}

	// loop ended without reaching valid or invalid Status
	return "", errChallengeRetriesExhausted
}

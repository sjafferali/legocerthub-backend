package authorizations

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"sync"
)

var errAuthPending = errors.New("one or more auths are still in 'pending' status")

// FulfillAuths attempts to validate each of the auth URLs in the slice of auth URLs. It returns 'valid' Status if all auths were
// determined to be 'valid'. It returns 'invalid' if any of the auths were determined to be in any state other than valid or pending.
// It returns an error if any of the auth Statuses could not be determined or if any are still in pending.
func (service *Service) FulfillAuths(authUrls []string, method challenges.Method, key acme.AccountKey, isStaging bool) (status string, err error) {
	// aysnc checking the authz for validity
	var wg sync.WaitGroup
	wgSize := len(authUrls)

	wg.Add(wgSize)
	wgStatuses := make(chan string, wgSize)
	wgErrors := make(chan error, wgSize)

	// fulfill each auth concurrently
	// TODO: Add context to cancel everything if any auth fails / invalid?
	for i := range authUrls {
		go func(authUrl string, method challenges.Method, key acme.AccountKey, isStaging bool) {
			defer wg.Done()
			status, err = service.fulfillAuth(authUrl, method, key, isStaging)
			wgStatuses <- status
			wgErrors <- err
		}(authUrls[i], method, key, isStaging)
	}

	// wait for all auths to do their thing
	wg.Wait()

	// close channel
	close(wgStatuses)
	close(wgErrors)

	// check for errors, returns first encountered error only
	for err = range wgErrors {
		if err != nil {
			return "", err
		}
	}

	// check if all auths are valid, if not return invalid
	for status = range wgStatuses {
		if status == "pending" {
			return "", errAuthPending
		} else if status != "valid" {
			return "invalid", nil
		}
	}

	// if looped through all statuses and confirmed all 'valid'
	return "valid", nil
}

// fulfillAuth attempts to validate an auth URL using the specified method. It will either respond from cache
// or call an authWorker.  An error is returned if the auth status could not be determined.
func (service *Service) fulfillAuth(authUrl string, method challenges.Method, key acme.AccountKey, isStaging bool) (status string, err error) {
	// add authUrl to working and call a worker, if the authUrl is already being worked,
	// block and return the cached result. If the cached result is an error, try to work
	// the auth again.
	for {
		exists, signal := service.working.add(authUrl)
		// if doesn't exist (not working) break from loop and call worker
		if !exists {
			break
		}

		// block until this auth's work (on other thread) is complete
		<-signal

		// read results of the other thread from the cache
		status, err = service.cache.read(authUrl)

		// if no error, return the status from cache
		if err == nil {
			return status, nil
		}

		// if there was an error in cache, loop repeats and authUrl is added to working again
	}

	// defer removing auth once it has been worked
	defer func(authUrl string, service *Service) {
		err := service.working.remove(authUrl)
		if err != nil {
			service.logger.Error(err)
		}
	}(authUrl, service)

	// work the auth
	status, err = service.authWorker(authUrl, method, key, isStaging)

	// cache result &
	// error check
	if err != nil {
		service.cache.add(authUrl, "", err)
		return "", err
	}

	service.cache.add(authUrl, status, nil)
	return status, nil
}

// authWorker returns the Status of an authorization URL. If the authorization Status is currently 'pending', authWorker attempts to
// move the authorization to the 'valid' Status.  An error is returned if the Status can't be determined.
func (service *Service) authWorker(authUrl string, method challenges.Method, key acme.AccountKey, isStaging bool) (status string, err error) {
	var auth acme.Authorization

	// PaG the authorization
	if isStaging {
		auth, err = service.acmeStaging.GetAuth(authUrl, key)
	} else {
		auth, err = service.acmeStaging.GetAuth(authUrl, key)
	}
	if err != nil {
		return "", err
	}

	// return the authoization Status
	switch auth.Status {
	// try to solve a challenge if auth is pending
	case "pending":
		auth.Status, err = service.challenges.Solve(auth.Identifier, auth.Challenges, method, key, isStaging)
		// return error if couldn't solve
		if err != nil {
			return "", err
		}
		// if no error, auth Status should now be "valid" or "invalid"
		fallthrough

	// if the auth is in any of these Statuses, break to return Status
	case "valid", "invalid", "deactivated", "expired", "revoked":
		// break

	// if the Status is unknown or otherwise unmatched, error
	default:
		return "", errors.New("unknown authorization status")
	}

	return auth.Status, nil
}

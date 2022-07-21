package acme_accounts

import (
	"encoding/json"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/validation"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// NameDescPayload is the struct for editing an existing account name and desc
type NameDescPayload struct {
	ID          *int    `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// PutNameDescAccount is a handler that sets the name and description of an account
// within storage
func (service *Service) PutNameDescAccount(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// payload decoding
	var payload NameDescPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// validation
	// id
	err = service.isIdExistingMatch(idParam, payload.ID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// name (optional)
	if payload.Name != nil {
		err = service.isNameValid(payload.ID, payload.Name)
		if err != nil {
			service.logger.Debug(err)
			return output.ErrValidationFailed
		}
	}
	///

	// save account name and desc to storage, which also returns the account id with new
	// name and description
	err = service.storage.PutNameDescAccount(payload)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      idParam,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

// ChangeEmailPayload is the struct for updating an account's email address
type ChangeEmailPayload struct {
	ID    *int    `json:"id"`
	Email *string `json:"email"`
}

// ChangeEmail() is a handler that updates an ACME account with the specified
// email address and saves the updated address to storage
func (service *Service) ChangeEmail(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// decode payload
	var payload ChangeEmailPayload
	err = json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	/// validation
	// id
	err = service.isIdExistingMatch(idParam, payload.ID)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	// email (update cannot be to blank)
	err = validation.IsEmailValid(payload.Email)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}
	///

	// fetch the relevant account
	account, err := service.storage.GetOneAccountById(idParam, true)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// get AccountKey
	accountKey, err := account.AccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// make ACME update email payload
	acmePayload := acme.UpdateAccountPayload{
		Contact: emailToContact(*payload.Email),
	}

	// send the email update to ACME
	var acmeResponse acme.AcmeAccountResponse
	if *account.IsStaging {
		acmeResponse, err = service.acmeStaging.UpdateAccount(acmePayload, accountKey)
	} else {
		acmeResponse, err = service.acmeProd.UpdateAccount(acmePayload, accountKey)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to account
	err = service.storage.PutLEAccountResponse(idParam, acmeResponse)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status:  http.StatusOK,
		Message: "updated",
		ID:      idParam,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

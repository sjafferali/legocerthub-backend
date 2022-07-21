package certificates

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/output"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// OrderCert sends the account information to the ACME new-order endpoint
// which creates a new order for the certificate
// endpoint: /api/v1/certificates/:id/order
func (service *Service) OrderCert(w http.ResponseWriter, r *http.Request) (err error) {
	idParamStr := httprouter.ParamsFromContext(r.Context()).ByName("id")

	// convert id param to an integer
	idParam, err := strconv.Atoi(idParamStr)
	if err != nil {
		service.logger.Debug(err)
		return output.ErrValidationFailed
	}

	// fetch the relevant cert
	cert, err := service.storage.GetOneCertById(idParam)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// fetch the relevant account (needed to get sensitive info to make Account Key)
	account, err := service.storage.GetOneAccountById(*cert.AcmeAccount.ID, true)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// no need to validate, can try to order any cert in storage

	// get account key
	key, err := account.AccountKey()
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	//send the new-account to ACME
	var acmeResponse acme.OrderResponse
	if *cert.AcmeAccount.IsStaging {
		acmeResponse, err = service.acmeStaging.NewOrder(cert.newOrderPayload(), key)
	} else {
		acmeResponse, err = service.acmeProd.NewOrder(cert.newOrderPayload(), key)
	}
	if err != nil {
		service.logger.Error(err)
		return output.ErrInternal
	}

	// save ACME response to order storage
	newOrderId, err := service.storage.PostNewOrder(cert, acmeResponse)
	if err != nil {
		service.logger.Error(err)
		return output.ErrStorageGeneric
	}

	// return response to client
	response := output.JsonResponse{
		Status: http.StatusCreated,
		// Message: "order created", // TODO?
		Message: acmeResponse,
		ID:      newOrderId,
	}

	_, err = service.output.WriteJSON(w, response.Status, response, "response")
	if err != nil {
		service.logger.Error(err)
		return output.ErrWriteJsonFailed
	}

	return nil
}

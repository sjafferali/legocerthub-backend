package orders

import (
	"errors"
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/domain/authorizations"
	"legocerthub-backend/pkg/domain/certificates"
	"legocerthub-backend/pkg/output"
	"legocerthub-backend/pkg/pagination_sort"
	"time"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary orders service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetLogger() *zap.SugaredLogger
	GetOutputter() *output.Service
	GetOrderStorage() Storage
	GetAcmeProdService() *acme.Service
	GetAcmeStagingService() *acme.Service
	GetCertificatesService() *certificates.Service
	GetAuthsService() *authorizations.Service
}

// Storage interface for storage functions
type Storage interface {
	// orders
	GetOneOrder(orderId int) (order Order, err error)
	GetOrdersByCert(certId int) (orders []Order, err error)

	PostNewOrder(payload NewOrderAcmePayload) (newId int, err error)

	PutOrderAcme(payload UpdateAcmeOrderPayload) (err error)
	PutOrderInvalid(orderId int) (err error)
	UpdateFinalizedKey(orderId int, keyId int) (err error)
	UpdateOrderCert(orderId int, CertPayload CertPayload) (err error)
	RevokeOrder(orderId int) (err error)

	GetAllValidCurrentOrders(q pagination_sort.Query, maxTimeRemaining *time.Duration) (orders []Order, totalRows int, err error)
	GetNewestIncompleteCertOrderId(certId int) (orderId int, err error)

	// certs
	UpdateCertUpdatedTime(certId int) (err error)
}

// Keys service struct
type Service struct {
	logger         *zap.SugaredLogger
	output         *output.Service
	storage        Storage
	acmeProd       *acme.Service
	acmeStaging    *acme.Service
	certificates   *certificates.Service
	authorizations *authorizations.Service
	inProcess      *inProcess
	highJobs       chan orderJob
	lowJobs        chan orderJob
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// output service
	service.output = app.GetOutputter()
	if service.output == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetOrderStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	// acme services
	service.acmeProd = app.GetAcmeProdService()
	if service.acmeProd == nil {
		return nil, errServiceComponent
	}
	service.acmeStaging = app.GetAcmeStagingService()
	if service.acmeStaging == nil {
		return nil, errServiceComponent
	}

	// certificates
	service.certificates = app.GetCertificatesService()
	if service.certificates == nil {
		return nil, errServiceComponent
	}

	// authorization service
	service.authorizations = app.GetAuthsService()
	if service.authorizations == nil {
		return nil, errServiceComponent
	}

	// initialize inProcess (tracker)
	service.inProcess = newInProcess()

	// workers
	// make job channels for order workers
	service.highJobs = make(chan orderJob)
	service.lowJobs = make(chan orderJob)
	workerCount := 3

	// make workers
	for i := 0; i < workerCount; i++ {
		go service.makeOrderWorker(i, service.highJobs, service.lowJobs)
	}

	// start cert refresh goroutine
	service.backgroundCertRefresher()

	return service, nil
}

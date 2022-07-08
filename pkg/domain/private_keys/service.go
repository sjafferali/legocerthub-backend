package private_keys

import (
	"errors"

	"go.uber.org/zap"
)

var errServiceComponent = errors.New("necessary key service component is missing")

// App interface is for connecting to the main app
type App interface {
	GetKeyStorage() Storage
	GetLogger() *zap.SugaredLogger
}

// Storage interface for storage functions
type Storage interface {
	GetAllKeys() ([]Key, error)
	GetOneKeyById(int) (Key, error)
	GetOneKeyByName(string) (Key, error)

	PostNewKey(NewPayload) (keyId int, err error)
	PutNameDescKey(NameDescPayload) error
	DeleteKey(int) error

	GetAvailableKeys() ([]Key, error)
}

// Keys service struct
type Service struct {
	logger  *zap.SugaredLogger
	storage Storage
}

// NewService creates a new private_key service
func NewService(app App) (*Service, error) {
	service := new(Service)

	// logger
	service.logger = app.GetLogger()
	if service.logger == nil {
		return nil, errServiceComponent
	}

	// storage
	service.storage = app.GetKeyStorage()
	if service.storage == nil {
		return nil, errServiceComponent
	}

	return service, nil
}

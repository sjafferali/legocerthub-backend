package app

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// routes creates the application's router and adds the routes. It also
// inserts the CORS middleware before returning the routes
func (app *Application) routes() http.Handler {
	app.router = httprouter.New()

	// app - auth - insecure
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/login", app.auth.Login)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/refresh", app.auth.Refresh)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/auth/logout", app.auth.Logout)
	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/auth/changepassword", app.auth.ChangePassword)

	// app
	app.makeHandle(http.MethodHead, apiUrlPath+"/status", app.statusHandler)
	app.makeHandle(http.MethodGet, apiUrlPath+"/status", app.statusHandler)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/log", app.viewCurrentLogHandler)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/logs", app.downloadLogsHandler)

	// updater
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/app/new-version", app.updater.GetNewVersionInfo)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/app/new-version", app.updater.CheckForNewVersion)

	// private_keys
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys", app.keys.GetAllKeys)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id", app.keys.GetOneKey)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/privatekeys/:id/download", app.keys.DownloadOneKey)

	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/privatekeys", app.keys.PostNewKey)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.StageNewApiKey)
	app.makeHandle(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id/apikey", app.keys.RemoveOldApiKey)

	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/privatekeys/:id", app.keys.PutKeyUpdate)

	app.makeHandle(http.MethodDelete, apiUrlPath+"/v1/privatekeys/:id", app.keys.DeleteKey)

	// acme_accounts
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/acmeaccounts", app.accounts.GetAllAccounts)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.GetOneAccount)

	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts", app.accounts.PostNewAccount)

	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.PutNameDescAccount)
	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/email", app.accounts.ChangeEmail)
	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/acmeaccounts/:id/key-change", app.accounts.RolloverKey)

	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/new-account", app.accounts.NewAcmeAccount)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/acmeaccounts/:id/deactivate", app.accounts.Deactivate)

	app.makeHandle(http.MethodDelete, apiUrlPath+"/v1/acmeaccounts/:id", app.accounts.DeleteAccount)

	// certificates
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/certificates", app.certificates.GetAllCerts)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid", app.certificates.GetOneCert)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/download", app.certificates.DownloadOneCert)

	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/certificates", app.certificates.PostNewCert)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.StageNewApiKey)
	app.makeHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid/apikey", app.certificates.RemoveOldApiKey)

	app.makeHandle(http.MethodPut, apiUrlPath+"/v1/certificates/:certid", app.certificates.PutDetailsCert)

	app.makeHandle(http.MethodDelete, apiUrlPath+"/v1/certificates/:certid", app.certificates.DeleteCert)

	// orders (for certificates)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/orders/currentvalid", app.orders.GetAllValidCurrentOrders)
	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.GetCertOrders)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders", app.orders.NewOrder)

	app.makeHandle(http.MethodGet, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/download", app.orders.DownloadOneOrder)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid", app.orders.FulfillExistingOrder)
	app.makeHandle(http.MethodPost, apiUrlPath+"/v1/certificates/:certid/orders/:orderid/revoke", app.orders.RevokeOrder)

	// download keys and certs
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatekeys/:name", app.download.DownloadKeyViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certificates/:name", app.download.DownloadCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatecerts/:name", app.download.DownloadPrivateCertViaHeader)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certrootchains/:name", app.download.DownloadCertRootChainViaHeader)

	// download keys and certs - via URL routes
	// include
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatekeys/:name/*apiKey", app.download.DownloadKeyViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certificates/:name/*apiKey", app.download.DownloadCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/privatecerts/:name/*apiKey", app.download.DownloadPrivateCertViaUrl)
	app.makeDownloadHandle(http.MethodGet, apiUrlPath+"/v1/download/certrootchains/:name/*apiKey", app.download.DownloadCertRootChainViaUrl)

	// frontend (if enabled)
	if *app.config.ServeFrontend {
		// log availability
		app.logger.Infof("frontend hosting enabled and available at: %s", frontendUrlPath)

		// configure environment file
		app.setFrontendEnv()

		// redirect root to frontend app
		app.makeHandle(http.MethodGet, "/", redirectToFrontendHandler)

		// add file server route for frontend
		app.makeHandle(http.MethodGet, frontendUrlPath+"/*anything", app.frontendHandler)
	}

	// invalid route
	app.router.NotFound = app.makeHandler(app.notFoundHandler)
	app.router.MethodNotAllowed = app.makeHandler(app.notFoundHandler)

	return app.enableCORS(app.router)
}

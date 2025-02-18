package certificates

import (
	"legocerthub-backend/pkg/acme"
	"legocerthub-backend/pkg/challenges"
	"legocerthub-backend/pkg/domain/acme_accounts"
	"legocerthub-backend/pkg/domain/private_keys"
)

// Certificate is a single certificate with all of its fields
type Certificate struct {
	ID                 int
	Name               string
	Description        string
	CertificateKey     private_keys.Key
	CertificateAccount acme_accounts.Account
	Subject            string
	SubjectAltNames    []string
	ChallengeMethod    challenges.Method
	Organization       string
	OrganizationalUnit string
	Country            string
	State              string
	City               string
	CreatedAt          int
	UpdatedAt          int
	ApiKey             string
	ApiKeyNew          string
	ApiKeyViaUrl       bool
}

// certificateSummaryResponse is a JSON response containing only
// fields desired for the summary
type certificateSummaryResponse struct {
	ID                 int                               `json:"id"`
	Name               string                            `json:"name"`
	Description        string                            `json:"description"`
	CertificateKey     certificateKeySummaryResponse     `json:"private_key"`
	CertificateAccount certificateAccountSummaryResponse `json:"acme_account"`
	Subject            string                            `json:"subject"`
	SubjectAltNames    []string                          `json:"subject_alts"`
	ChallengeMethod    challenges.Method                 `json:"challenge_method"`
	ApiKeyViaUrl       bool                              `json:"api_key_via_url"`
}

type certificateKeySummaryResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type certificateAccountSummaryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	IsStaging bool   `json:"is_staging"`
}

func (cert Certificate) summaryResponse() certificateSummaryResponse {
	return certificateSummaryResponse{
		ID:          cert.ID,
		Name:        cert.Name,
		Description: cert.Description,
		CertificateKey: certificateKeySummaryResponse{
			ID:   cert.CertificateKey.ID,
			Name: cert.CertificateKey.Name,
		},
		CertificateAccount: certificateAccountSummaryResponse{
			ID:        cert.CertificateAccount.ID,
			Name:      cert.CertificateAccount.Name,
			IsStaging: cert.CertificateAccount.IsStaging,
		},
		Subject:         cert.Subject,
		SubjectAltNames: cert.SubjectAltNames,
		ChallengeMethod: cert.ChallengeMethod,
		ApiKeyViaUrl:    cert.ApiKeyViaUrl,
	}
}

// certificateDetailedResponse is a JSON response containing all
// fields that can be returned as JSON
type certificateDetailedResponse struct {
	certificateSummaryResponse
	Organization       string `json:"organization"`
	OrganizationalUnit string `json:"organizational_unit"`
	Country            string `json:"country"`
	State              string `json:"state"`
	City               string `json:"city"`
	CreatedAt          int    `json:"created_at"`
	UpdatedAt          int    `json:"updated_at"`
	ApiKey             string `json:"api_key"`
	ApiKeyNew          string `json:"api_key_new,omitempty"`
}

func (cert Certificate) detailedResponse(withSensitive bool) certificateDetailedResponse {
	// option to redact sensitive info
	apiKey := cert.ApiKey
	apiKeyNew := cert.ApiKeyNew
	if !withSensitive {
		apiKey = "[redacted]"
		// redact if not empty
		if apiKeyNew != "" {
			apiKeyNew = "[redacted]"
		}
	}

	return certificateDetailedResponse{
		certificateSummaryResponse: cert.summaryResponse(),
		Organization:               cert.Organization,
		OrganizationalUnit:         cert.OrganizationalUnit,
		Country:                    cert.Country,
		State:                      cert.State,
		City:                       cert.City,
		CreatedAt:                  cert.CreatedAt,
		UpdatedAt:                  cert.UpdatedAt,
		ApiKey:                     apiKey,
		ApiKeyNew:                  apiKeyNew,
	}
}

// NewOrderPayload creates the appropriate newOrder payload for ACME
func (cert *Certificate) NewOrderPayload() acme.NewOrderPayload {
	var identifiers []acme.Identifier

	// subject is always required and should be first
	// dns is the only supported type and is hardcoded
	identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: cert.Subject})

	// add alt names if they exist
	if cert.SubjectAltNames != nil {
		for _, name := range cert.SubjectAltNames {
			identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: name})
		}
	}

	return acme.NewOrderPayload{
		Identifiers: identifiers,
	}
}

// new account info
// used to return info about valid options when making a new account
type newCertOptions struct {
	AvailableKeys             []private_keys.KeySummaryResponse      `json:"private_keys"`
	UsableAccounts            []acme_accounts.AccountSummaryResponse `json:"acme_accounts"`
	AvailableChallengeMethods []challenges.Method                    `json:"challenge_methods"`
}

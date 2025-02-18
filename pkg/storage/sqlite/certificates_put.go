package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/certificates"
	"time"
)

// PutDetailsCert saves details about the cert that can be updated at any time. It only updates
// the details which are provided
func (store *Storage) PutDetailsCert(payload certificates.DetailsUpdatePayload) (err error) {
	// database update
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
		UPDATE
			certificates
		SET
			name = case when $1 is null then name else $1 end,
			description = case when $2 is null then description else $2 end,
			private_key_id = case when $3 is null then private_key_id else $3 end,
			challenge_method = case when $4 is null then challenge_method else $4 end,
			subject_alts = case when $5 is null then subject_alts else $5 end,
			csr_org = case when $6 is null then csr_org else $6 end,
			csr_ou = case when $7 is null then csr_ou else $7 end,
			csr_country = case when $8 is null then csr_country else $8 end,
			csr_state = case when $9 is null then csr_state else $9 end,
			csr_city = case when $10 is null then csr_city else $10 end,
			api_key_via_url = case when $12 is null then api_key_via_url else $12 end,
			updated_at = $11
		WHERE
			id = $13
		`

	_, err = store.Db.ExecContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.PrivateKeyId,
		payload.ChallengeMethodValue,
		makeCommaJoinedString(payload.SubjectAltNames),
		payload.Organization,
		payload.OrganizationalUnit,
		payload.Country,
		payload.State,
		payload.City,
		payload.ApiKeyViaUrl,
		payload.UpdatedAt,
		payload.ID,
	)

	if err != nil {
		return err
	}

	return nil
}

// UpdateCertUpdatedTime sets the specified order's updated_at to now
func (store *Storage) UpdateCertUpdatedTime(certId int) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		updated_at = $1
	WHERE
		id = $2
	`

	_, err = store.Db.ExecContext(ctx, query,
		time.Now().Unix(),
		certId,
	)
	if err != nil {
		return err
	}

	// TODO: Handle 0 rows updated.

	return nil
}

// PutCertNewApiKey sets a cert's new api key and updates the updated at time
func (store *Storage) PutCertNewApiKey(certId int, newApiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		api_key_new = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.Db.ExecContext(ctx, query,
		newApiKey,
		updateTimeUnix,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}

// PutCertApiKey sets a cert's api key and updates the updated at time
func (store *Storage) PutCertApiKey(certId int, apiKey string, updateTimeUnix int) (err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	UPDATE
		certificates
	SET
		api_key = $1,
		updated_at = $2
	WHERE
		id = $3
	`

	_, err = store.Db.ExecContext(ctx, query,
		apiKey,
		updateTimeUnix,
		certId,
	)

	if err != nil {
		return err
	}

	return nil
}

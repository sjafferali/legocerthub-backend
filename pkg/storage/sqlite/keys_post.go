package sqlite

import (
	"context"
	"legocerthub-backend/pkg/domain/private_keys"
)

// PostNewKey saves the KeyExtended to the db as a new key
func (store *Storage) PostNewKey(payload private_keys.NewPayload) (id int, err error) {
	// database action
	ctx, cancel := context.WithTimeout(context.Background(), store.Timeout)
	defer cancel()

	query := `
	INSERT INTO private_keys (name, description, algorithm, pem, api_key, api_key_disabled, api_key_via_url, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id
	`

	// insert and scan the new id
	err = store.Db.QueryRowContext(ctx, query,
		payload.Name,
		payload.Description,
		payload.AlgorithmValue,
		payload.PemContent,
		payload.ApiKey,
		payload.ApiKeyDisabled,
		payload.ApiKeyViaUrl,
		payload.CreatedAt,
		payload.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return -2, err
	}

	return id, nil
}

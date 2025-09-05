-- name: RetrieveAuthCredentials :one
SELECT key_id, client_id, priv_key_pem FROM axm_names WHERE name = ?;

-- name: RetrieveClientAssertion :one
SELECT ca_token, ca_validity_sec, ca_expiry_unix, client_id FROM axm_names WHERE name = ? FOR UPDATE;

-- name: UpdateClientAssertion :exec
UPDATE axm_names SET ca_token = ?, ca_validity_sec = ?, ca_expiry_unix = ? WHERE name = ?;
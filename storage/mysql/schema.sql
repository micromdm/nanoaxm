CREATE TABLE axm_names (
    name VARCHAR(255) NOT NULL,

    key_id       VARCHAR(255) NOT NULL,
    client_id    VARCHAR(255) NOT NULL,
    priv_key_pem TEXT         NOT NULL,

    ca_token        TEXT NULL,
    ca_validity_sec INT  NULL, -- validity in seconds
    ca_expiry_unix  INT  NULL, -- unix timestamp

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    PRIMARY KEY (name)
);
#!/bin/sh

set -xeou

# using fork github.com/jessepeterson/go-jsonschema (micromdm branch)

go-jsonschema \
	-p abm \
	--capitalization ID,URL,MEID,EID,IMEI,WiFi,MAC,SCIM,MDM,HREF,JSON \
	-t \
	--tags json \
	--only-models \
	--disable-omitzero \
	--output schema.go \
	apple-device-services/abm/schemas/*.json

# Generate the HTTP endpoint methods in this (abm) package.
gen_openapi_endpoints.py \
	--spec apple-device-services/abm/schemas/OpenAPI.json \
	--package abm \
	--receiver 'c *Client' \
	--receiver-var c \
	--auth-param axmName \
	--schema-pkg abm \
	--capitalization ID,URL,MEID,EID,IMEI,WiFi,MAC,SCIM,MDM,HREF,JSON \
	--words OrgDevice,MDMServer,OrgDeviceActivity,AppleCare,OrganizationalUnit,UserGroup \
	--base-url https://api-business.apple.com \
	--base-url-symbol BaseURL \
	--do-func Do \
	--has-out-status \
	--has-err-statuses \
	--out endpoints.go

gofmt -w endpoints.go

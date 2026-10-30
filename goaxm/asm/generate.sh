#!/bin/sh

set -xeou

# using fork github.com/jessepeterson/go-jsonschema (micromdm branch)

go-jsonschema \
	-p asm \
	--capitalization ID,URL,MEID,EID,IMEI,WiFi,MAC,SCIM,MDM,HREF,JSON \
	-t \
	--tags json \
	--only-models \
	--disable-omitzero \
	--output schema.go \
	apple-device-services/asm/schemas/*.json

# Generate the HTTP endpoint methods in this (asm) package.
gen_openapi_endpoints.py \
	--spec apple-device-services/asm/schemas/OpenAPI.json \
	--package asm \
	--receiver 'c *Client' \
	--receiver-var c \
	--auth-param axmName \
	--schema-pkg asm \
	--capitalization ID,URL,MEID,EID,IMEI,WiFi,MAC,SCIM,MDM,HREF,JSON \
	--words OrgDevice,MDMServer,OrgDeviceActivity,AppleCare \
	--base-url https://api-school.apple.com \
	--base-url-symbol BaseURL \
	--do-func Do \
	--has-out-status \
	--has-err-statuses \
	--out endpoints.go

gofmt -w endpoints.go

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
	apple-device-services/asm/schemas/*.json


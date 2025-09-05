#!/bin/sh

# Arguments:
#
# $1 - Client ID: identifier from Apple AxM portal
# $2 - Key ID: identifier from Apple AxM portal
# $3 - Private key: path to file downloaded from AxM portal

# Required envvars:
#
# $AXM_NAME - AXM name
# $BASE_URL - base URL to which paths are appended
# $API_KEY  - API key for service

URL="$BASE_URL/authcreds"

if [ "x$API_USER" = "x" ]; then
    API_USER="nanoaxm"
fi

curl \
    $CURL_OPTS \
    -u "$API_USER:$API_KEY" \
    -X POST \
    -F "axm_name=$AXM_NAME" \
    -F "client_id=$1" \
    -F "key_id=$2" \
    -F "private_key=@$3" \
    "$URL"

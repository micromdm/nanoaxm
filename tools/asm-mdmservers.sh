#!/bin/sh

# Get a list of device management services in an organization.
# See https://developer.apple.com/documentation/appleschoolmanagerapi/get-mdm-servers

# Required envvars:
#
# $AXM_NAME - AXM name
# $BASE_URL - base URL to which paths are appended
# $API_KEY  - API key for service

URL="$BASE_URL/proxy/school/$AXM_NAME/v1/mdmServers"

if [ "x$API_USER" = "x" ]; then
    API_USER="nanoaxm"
fi

curl \
    $CURL_OPTS \
    -u "$API_USER:$API_KEY" \
    "$URL"

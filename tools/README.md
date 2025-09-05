# NanoAXM tools and scripts

A set of example shell scripts and tools for working with the server proxy and configuration APIs. For more information how and why you might use these scripts please see the [Operations Guide](../docs/operations-guide.md).

## Requirements

These scripts require a couple tools to be in your shell path:

* [curl](https://curl.se/)
* [jq](https://stedolan.github.io/jq/)
* Bourne-ish shell interpreter.

## Setup

For these scripts to work you have to have a few environment variables set first. You could embed these into their own file and use `source` to set them if you'd like to re-use them.

```bash
# the URL of the running nanoaxm server
export BASE_URL='http://[::1]:9005'
# should match the -api switch of the nanoaxm server
export API_KEY=supersecret
# the AxM name (instance) you want to use
export AXM_NAME=myAxmToken1
```

Be cautious to unset these variables or exit the shell when you're done so as not to leave API keys hanging around in environment variables. Also beware the API key is provided to `curl` on the command line and will likely be visible in process lists.

## Example

First setup the environment variables per above then the scripts can be executed:

```bash
# Get a list of device management services in an organization.
% ./tools/abm-mdmservers.sh 
{
  "data" : [ ... ],
  "links" : { ... },
  "meta" : { ... }
}
```

## Troubleshooting

If there's problems you can optionally set the `$CURL_OPTS` envvar as well to display more debugging information from the curl call:

```bash
# enable curl verbose output
export CURL_OPTS=-v
```

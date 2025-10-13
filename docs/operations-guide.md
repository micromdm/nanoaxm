# NanoAXM Operations Guide

This is a brief overview of the various tools and utilities for working with NanoAXM.

## AxM names

NanoAXM supports configuring multiple API OAuth credentials (API Account in Apple's terms). These different credentials are referenced by an arbitrary name string that you specify called an "AxM name." This string is used to both configure the API credentials as well to reference these configurations for actually talking to the Apple AxM API endpoints.

> [!WARNING]
> Because the name string is used pervasively in URL API paths you probably want to avoid names that include things like forward-slashes "/", spaces, or anything else really that might have trouble in URLs.

## Server

The server included in NanoAXM serves two main purposes:

1. Setup & configuration of the AxM name(s) — that is, the locally-named instances that correspond API users in the Apple Business Manager (ABM) or Apple School Manager (ASM)portals. Configuration includes uploading the Client ID, Key ID, and downloaded private key to the server for storage and later use. See the "API endpoints" section below for more.
1. Accessing the actual AxM APIs using a transparently-authenticating reverse proxy. After you've configured the authentication tokens using the above APIs the server provides a reverse proxy to talk to the Apple AxM endpoints. You don't have to worry about session management or token authentication: this's taken care of for you. All you need to do is use a special URL path and normal API (HTTP Basic) authentication and you can talk to the AxM APIs unfiltered. See the "Reverse proxy" section below for more.

### Command line flags

Command line flags can be specified using command line arguments or environment variables (in NanoAXM versions later than v0.1.0). Flags take precedence over environment variables, which take precedence over default values. Environment variables are denoted in square brackets below (e.g., [HELLO]), and default values are shown in parentheses (e.g., (default "world")). If an environment variable is currently set then the help output will add "is set" as an indicator.

#### -h, -help

Built-in flag that prints all other available flags, environment variables, and defaults.

#### -api string

* API key for API endpoints [NANOAXM_API]

Required. API authentication in the NanoAXM server is simply HTTP Basic authentication using "nanoaxm" as the username and the API key (from this flag) as the password.

#### -debug

* log debug messages [NANOAXM_DEBUG]

Enable additional debug logging.

#### -listen string

* HTTP listen address [NANOAXM_LISTEN] (default ":9005")

Specifies the network listen address (interface and port number) for the server to listen on.

#### -storage, -storage-dsn, & -storage-options

* -storage string
  * storage backend [NANOAXM_STORAGE] (default "file")
* -storage-dsn string
  * storage backend data source name [NANOAXM_STORAGE_DSN]
* -storage-options string
  * storage backend options [NANOAXM_STORAGE_OPTIONS]

The `-storage`, `-storage-dsn`, and `-storage-options` flags together configure the storage backend. `-storage` specifies the name of the backend type while `-storage-dsn` specifies the backend data source name (e.g. the connection string). The optional `-storage-options` flag specifies options for the backend if it supports them. If no `-storage` backend is specified then `file` is used as a default.

##### file storage backend

* `-storage file`

Configure the `file` storage backend. This backend manages AxM credentials and configuration data within plain filesystem files and directories using a key-value storage system. It has zero dependencies, no options, and should run out of the box with no other depenencies. The `-storage-dsn` flag specifies the filesystem directory for the database. If no `storage-dsn` is specified then `db` is used as a default.

*Example:* `-storage file -storage-dsn /path/to/my/db`

##### in-memory storage backend

* `-storage inmem`

Configure the `inmem` in-memory storage backend. This backend manages AxM credentials and configuration data entirely in *volatile* memory. There are no options and the DSN is ignored.

> [!CAUTION]
> All data is lost when the server process has exited.

*Example:* `-storage inmem`

##### mysql storage backend

* `-storage mysql`

Configures the MySQL storage backend. The `-dsn` flag should be in the [format the SQL driver expects](https://github.com/go-sql-driver/mysql#dsn-data-source-name). MySQL 8.0.19 or later is required.

> [!TIP]
> Be sure to create the storage tables with the [schema.sql](../storage/mysql/schema.sql) file first.

*Example:* `-storage mysql -dsn nanoaxm:nanoaxm/myaxmdb`

#### -version

* print version and exit

Print version and exit.

### API endpoints

API endpoints for the NanoAXM server.

A brief overview of the endpoints is provided here. For detailed API semantics please see the [OpenAPI documentation for NanoAXM](https://www.jessepeterson.space/swagger/nanoaxm.html). The OpenAPI source YAML is a part of this project.

> [!TIP]
> You aren't required to use these APIs directly — NanoAXM provides a set of tools and scripts for working with some of these endpoints — see the "Tools and scripts" section, below.

#### Version

* Endpoint: `GET /version`

Returns a JSON response with the version of the running NanoAXM server.

#### Authentication Credentials

* Endpoint: `GET /authcreds`
* Endpoint: `POST /authcreds`

Creates or updates the OAuth 2 authentication credentials for the provided AxM name. When requesting using `GET`, an HTML form is presented. When using `POST` data is submitted as typical HTTP multi-part form data.

### Reverse proxy

In addition to individually handling some of various Apple AxM API endpoints in its `goaxm` library NanoAXM provides a transparently-authenticating HTTP reverse proxy to the Apple AxM servers. This allows us to simply provide the server with the Apple AxM endpoint, the NanoAXM "AxM name," and the API key, and we can talk to any of the Apple AxM endpoint APIs. The server will authenticate to the Apple AxM server and keep track of session management transparently behind the scenes. To be clear: this means you do not have to use the OAuth 2 HTTP headers to authenticate nor to manage and update them with each request. NanoAXM does this for you.

The proxy URLs are accessible as: `/proxy/business/{name}/endpoint` or `/proxy/school/{name}/endpoint` where `/endpoint` is the Apple AxM API endpoint you want to access. In the case of `/proxy/business` the proxy will automatically translate this URL to `https://https://api-business.apple.com/endpoint` and use `{name}` for retrieving the AxM authentication OAuth 2 tokens. Note that in some cases, for some endpoints, various HTTP headers are added or removed:

* For any proxy request the API authentication header is removed before passing to the underlying Apple AxM server.

> [!TIP]
> For simple cases you don't need to use this proxy directly — NanoAXM provides a set of tools and scripts for working with some of the AXM endpoints — see the "Tools and scripts" section, below.

#### Example usage

This example is taken directly out of the `./tools/abm-mdmservers.sh` helper script under the "Tools and scripts" section, below, but we'll duplicate it for illustrative purposes:

```bash
% curl -v -u nanoaxm:supersecret 'http://[::1]:9005/proxy/business/myAxmToken1/v1/mdmServers'
*   Trying [::1]:9005...
* Connected to ::1 (::1) port 9005
* Server auth using Basic with user 'nanoaxm'
> GET /proxy/business/myAxmToken1/v1/mdmServers HTTP/1.1
> Host: [::1]:9005
> Authorization: Basic bmFub2F4bTpzdXBlcnNlY3JldA==
> User-Agent: curl/8.7.1
> Accept: */*
> 
* Request completely sent off
< HTTP/1.1 200 OK
< Apple-Originating-System: UnknownOriginatingSystem
< Apple-Seq: 0.0
< Apple-Timing-App: 199 ms
< Apple-Tk: false
< B3: ...
< Content-Length: 1132
< Content-Type: application/json
< Cross-Origin-Opener-Policy: same-origin
< Date: Fri, 29 Aug 2025 06:12:16 GMT
< Server: Apple
< Strict-Transport-Security: max-age=31536000; includeSubdomains
< X-Apple-Jingle-Correlation-Key: ...
< X-Apple-Request-Uuid: ...
< X-B3-Parentspanid: ...
< X-B3-Spanid: ...
< X-B3-Traceid: ...
< X-Content-Type-Options: nosniff
< X-Daiquiri-Instance: ...
< X-Frame-Options: SAMEORIGIN
< X-Responding-Instance: ...
< X-Unicorn-Originated: true
< X-Xss-Protection: 1; mode=block
< 
{
  "data" : [ ... ],
  "links" : { ... },
  "meta" : { ... }
}
* Connection #0 to host ::1 left intact
```

This request URL path was "translated" from `GET /proxy/business/myAxmToken1/v1/mdmServers` to `GET /v1/mdmServers` at the `https://api-business.apple.com` URL and authenticated using the `myAxmToken1` AxM name (assuming it was already configured, of course). Note that no OAuth 2 exchange happened, that was entirely handled by NanoAXM.

## Tools and scripts

The NanoAXM project includes some tools and scripts that use the above APIs in the server for performing some typical API tasks. These are basically just shell scripts that utilize `curl` and `jq` to drive the server API and/or Apple AxM API endpoints. Naturally those tools are requiremented for the scripts to work. These tools and scripts also have their own documentation under the `./tools` directory of the project as noted below.

Generally, the scripts are split into three types indicated by the script prefix:

* Scripts starting with `cfg-` configure the NanoAXM server.
* Scripts starting with `abm-` use the Reverse Proxy described above to perform operations with the Apple Business Manager API.
* Scripts starting with `asm-` use the Reverse Proxy described above to perform operations with the Apple School Manager API.

### Scripts

These scripts require setting up a few environment variables before use. Please see the [tools](../tools) for more documentation. But generally you'll need to set these environment variables for the scripts to work:

```bash
# the URL of the running nanoaxm server
export BASE_URL='http://[::1]:9005'
# should match the -api flag of the nanoaxm server
export API_KEY=supersecret
# the AxM name (instance) you want to use
export AXM_NAME=myAxmToken1
```

The [Quickstart Guide](quickstart.md) also documents some usage of these scripts, too.

#### cfg-authcreds.sh

For the AxM name (in the environment vraible `$AXM_NAME`) this script uploads the authentication credentials from the ABM or ASM portal  The `curl` call will upload the credentials and print out a plain text message informing its success.

This script has three required positional arguments:
- First: the Client ID: an identifier from the Apple AxM portal
- Second: the Key ID: an identifier from the Apple AxM portal
- Third: the private key: a path to the file downloaded from the AxM portal

##### Example usage

```bash
% ./tools/cfg-authcreds.sh BUSINESSAPI.AAA111 BBB222 /path/to/private.key
Saved authentication credentials for AXM name: myAxmToken1 (Client ID BUSINESSAPI.AAA111)
```

#### abm-mdmservers.sh

For the AxM name (in the environment vraible `$AXM_NAME`) this script gets a list of device management services in an organization. 

##### Example usage

```bash
% ./tools/abm-mdmservers.sh 
{
  "data" : [ ... ],
  "links" : { ... },
  "meta" : { ... }
}
```

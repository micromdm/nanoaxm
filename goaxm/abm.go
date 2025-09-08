package goaxm

import (
	"context"
	"net/url"

	"github.com/micromdm/nanoaxm/goaxm/abm"
)

// ABMv1MDMServers calls the Apple Business Manager API "v1" to
// Get a list of device management services in an organization.
// Query parameters may be provided in v, otherwise nil.
// See https://developer.apple.com/documentation/applebusinessmanagerapi/get-mdm-servers
func (c *Client) ABMv1MDMServers(ctx context.Context, axmName string, v url.Values) (*abm.MdmServersResponseJson, error) {
	var q string
	if v != nil {
		q = "?" + v.Encode()
	}
	out := new(abm.MdmServersResponseJson)
	return out, c.Do(ctx, axmName, "GET", "https://api-business.apple.com/v1/mdmServers"+q, nil, out)
}

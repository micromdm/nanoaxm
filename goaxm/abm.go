package goaxm

import (
	"context"

	"github.com/micromdm/nanoaxm/goaxm/abm"
)

// ABMv1MDMServers calls the Apple Business Manager API "v1" to
// Get a list of device management services in an organization.
// See https://developer.apple.com/documentation/applebusinessmanagerapi/get-mdm-servers
func (c *Client) ABMv1MDMServers(ctx context.Context, axmName string) (*abm.MdmServersResponseJson, error) {
	out := new(abm.MdmServersResponseJson)
	return out, c.Do(ctx, axmName, "GET", "https://api-business.apple.com/v1/mdmServers", nil, out)
}

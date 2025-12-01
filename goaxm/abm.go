package goaxm

import (
	"context"
	"net/http"
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
	return out, c.Do(ctx, axmName, "GET", "https://api-business.apple.com/v1/mdmServers"+q, nil, out, 0, nil)
}

func (c *Client) ABMv1OrgDeviceActivities(ctx context.Context, axmName string, req *abm.OrgDeviceActivityCreateRequestJson) (*abm.OrgDeviceActivityResponseJson, error) {
	out := new(abm.OrgDeviceActivityResponseJson)
	return out, c.Do(ctx, axmName, http.MethodPost, "https://api-business.apple.com/v1/orgDeviceActivities", req, out, 201, []int{400, 401, 403, 409, 422, 429})
}

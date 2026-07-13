package iracing

import (
	"context"
	"net/url"
	"strconv"
	"strings"
)

// Info returns the profile of the currently authenticated member (used to
// resolve the caller's own cust_id after login).
func (c *Client) Info(ctx context.Context) (Member, error) {
	var m Member
	if err := c.get(ctx, "/data/member/info", nil, &m); err != nil {
		return Member{}, err
	}
	return m, nil
}

// Members fetches one or more member profiles including licenses (iRating/SR).
func (c *Client) Members(ctx context.Context, custIDs ...int) ([]Member, error) {
	ids := make([]string, len(custIDs))
	for i, id := range custIDs {
		ids[i] = strconv.Itoa(id)
	}
	params := url.Values{}
	params.Set("cust_ids", strings.Join(ids, ","))
	params.Set("include_licenses", "true")

	var out struct {
		Members []Member `json:"members"`
	}
	if err := c.get(ctx, "/data/member/get", params, &out); err != nil {
		return nil, err
	}
	return out.Members, nil
}

// Member is a convenience wrapper returning a single profile.
func (c *Client) Member(ctx context.Context, custID int) (Member, error) {
	members, err := c.Members(ctx, custID)
	if err != nil {
		return Member{}, err
	}
	if len(members) == 0 {
		return Member{}, ErrNotFound
	}
	return members[0], nil
}

// Career returns per-category career stats for a member.
func (c *Client) Career(ctx context.Context, custID int) ([]CareerStat, error) {
	params := url.Values{}
	params.Set("cust_id", strconv.Itoa(custID))

	var out struct {
		Stats []CareerStat `json:"stats"`
	}
	if err := c.get(ctx, "/data/stats/member_career", params, &out); err != nil {
		return nil, err
	}
	return out.Stats, nil
}

// RecentRaces returns a member's recent finished races.
func (c *Client) RecentRaces(ctx context.Context, custID int) ([]RecentRace, error) {
	params := url.Values{}
	params.Set("cust_id", strconv.Itoa(custID))

	var out struct {
		Races []RecentRace `json:"races"`
	}
	if err := c.get(ctx, "/data/stats/member_recent_races", params, &out); err != nil {
		return nil, err
	}
	return out.Races, nil
}

// SearchDrivers looks up drivers by (partial) name. The lookup endpoint returns
// a bare JSON array rather than an object.
func (c *Client) SearchDrivers(ctx context.Context, term string) ([]DriverSearchResult, error) {
	params := url.Values{}
	params.Set("search_term", term)

	var out []DriverSearchResult
	if err := c.get(ctx, "/data/lookup/drivers", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

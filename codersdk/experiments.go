package codersdk

import (
	"context"
	"encoding/json"
	"net/http"
)

var (
	// ExperimentVSCodeLocal enables a workspace button to launch VSCode
	// and connect using the local VSCode extension.
	ExperimentVSCodeLocal = "vscode_local"
	// ExperimentsAll includes all known experiments.
	ExperimentsAll = []string{
		ExperimentVSCodeLocal,
	}
)

// ExperimentalConfig is the set of experiments that are enabled.
type ExperimentsResponse []string

func (c *Client) Experiments(ctx context.Context) (ExperimentsResponse, error) {
	res, err := c.Request(ctx, http.MethodGet, "/api/v2/experiments", nil)
	if err != nil {
		return ExperimentsResponse{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ExperimentsResponse{}, readBodyAsError(res)
	}
	var exp ExperimentsResponse
	return exp, json.NewDecoder(res.Body).Decode(&exp)
}

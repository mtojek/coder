package coderd

import (
	"net/http"

	"github.com/coder/coder/coderd/httpapi"
)

// @Summary Get experimental features
// @ID get-experimental-features
// @Security CoderSessionToken
// @Produce json
// @Tags General
// @Success 200 {object} codersdk.ExperimentsResponse
// @Router /experiments [get]
func (api *API) handleExperimentsGet(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	httpapi.Write(ctx, rw, http.StatusOK, api.DeploymentConfig.Experimental.Value)
}

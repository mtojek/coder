package coderd

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/xerrors"

	"github.com/coder/coder/coderd/database"
	"github.com/coder/coder/coderd/httpapi"
	"github.com/coder/coder/coderd/httpmw"
	"github.com/coder/coder/coderd/rbac"
	"github.com/coder/coder/codersdk"
)

// @Summary Get organization by ID
// @ID get-organization-by-id
// @Security CoderSessionToken
// @Produce json
// @Tags Organizations
// @Param organization path string true "Organization ID" format(uuid)
// @Success 200 {object} codersdk.Organization
// @Router /organizations/{organization} [get]
func (api *API) organization(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	organization := httpmw.OrganizationParam(r)

	if !api.Authorize(r, rbac.ActionRead, rbac.ResourceOrganization.
		InOrg(organization.ID)) {
		httpapi.ResourceNotFound(rw)
		return
	}

	httpapi.Write(ctx, rw, http.StatusOK, convertOrganization(organization))
}

// @Summary Create organization
// @ID create-organization
// @Security CoderSessionToken
// @Accept json
// @Produce json
// @Tags Organizations
// @Param request body codersdk.CreateOrganizationRequest true "Create organization request"
// @Success 201 {object} codersdk.Organization
// @Router /organizations [post]
func (api *API) postOrganizations(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	apiKey := httpmw.APIKey(r)
	// Create organization uses the organization resource without an OrgID.
	// This means you need the site wide permission to make a new organization.
	if !api.Authorize(r, rbac.ActionCreate, rbac.ResourceOrganization) {
		httpapi.Forbidden(rw)
		return
	}

	var req codersdk.CreateOrganizationRequest
	if !httpapi.Read(ctx, rw, r, &req) {
		return
	}

	_, err := api.Database.GetOrganizationByName(ctx, req.Name)
	if err == nil {
		httpapi.Write(ctx, rw, http.StatusConflict, codersdk.Response{
			Message: "Organization already exists with that name.",
		})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: fmt.Sprintf("Internal error fetching organization %q.", req.Name),
			Detail:  err.Error(),
		})
		return
	}

	var organization database.Organization
	err = api.Database.InTx(func(tx database.Store) error {
		organization, err = tx.InsertOrganization(ctx, database.InsertOrganizationParams{
			ID:        uuid.New(),
			Name:      req.Name,
			CreatedAt: database.Now(),
			UpdatedAt: database.Now(),
		})
		if err != nil {
			return xerrors.Errorf("create organization: %w", err)
		}
		_, err = tx.InsertOrganizationMember(ctx, database.InsertOrganizationMemberParams{
			OrganizationID: organization.ID,
			UserID:         apiKey.UserID,
			CreatedAt:      database.Now(),
			UpdatedAt:      database.Now(),
			Roles: []string{
				// TODO: When organizations are allowed to be created, we should
				// come back to determining the default role of the person who
				// creates the org. Until that happens, all users in an organization
				// should be just regular members.
				rbac.RoleOrgMember(organization.ID),
			},
		})
		if err != nil {
			return xerrors.Errorf("create organization admin: %w", err)
		}

		_, err = tx.InsertAllUsersGroup(ctx, organization.ID)
		if err != nil {
			return xerrors.Errorf("create %q group: %w", database.AllUsersGroup, err)
		}
		return nil
	}, nil)
	if err != nil {
		httpapi.Write(ctx, rw, http.StatusInternalServerError, codersdk.Response{
			Message: "Internal error inserting organization member.",
			Detail:  err.Error(),
		})
		return
	}

	httpapi.Write(ctx, rw, http.StatusCreated, convertOrganization(organization))
}

// convertOrganization consumes the database representation and outputs an API friendly representation.
func convertOrganization(organization database.Organization) codersdk.Organization {
	return codersdk.Organization{
		ID:        organization.ID,
		Name:      organization.Name,
		CreatedAt: organization.CreatedAt,
		UpdatedAt: organization.UpdatedAt,
	}
}

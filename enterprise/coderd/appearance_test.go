package coderd_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/coderd/coderdtest"
	"github.com/coder/coder/codersdk"
	"github.com/coder/coder/enterprise/coderd/coderdenttest"
	"github.com/coder/coder/enterprise/coderd/license"
	"github.com/coder/coder/testutil"
)

func TestServiceBanners(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()

	adminClient := coderdenttest.New(t, &coderdenttest.Options{})

	adminUser := coderdtest.CreateFirstUser(t, adminClient)

	// Even without a license, the banner should return as disabled.
	sb, err := adminClient.Appearance(ctx)
	require.NoError(t, err)
	require.False(t, sb.ServiceBanner.Enabled)

	coderdenttest.AddLicense(t, adminClient, coderdenttest.LicenseOptions{
		Features: license.Features{
			codersdk.FeatureAppearance: 1,
		},
	})

	// Default state
	sb, err = adminClient.Appearance(ctx)
	require.NoError(t, err)
	require.False(t, sb.ServiceBanner.Enabled)

	basicUserClient, _ := coderdtest.CreateAnotherUser(t, adminClient, adminUser.OrganizationID)

	sb.SupportLinks = nil // clean "support links" as they can't be modified using API

	// Regular user should be unable to set the banner
	sb.ServiceBanner.Enabled = true
	err = basicUserClient.UpdateAppearance(ctx, sb)
	require.Error(t, err)
	var sdkError *codersdk.Error
	require.True(t, errors.As(err, &sdkError))
	require.Equal(t, http.StatusForbidden, sdkError.StatusCode())

	// But an admin can
	wantBanner := sb
	wantBanner.ServiceBanner.Enabled = true
	wantBanner.ServiceBanner.Message = "Hey"
	wantBanner.ServiceBanner.BackgroundColor = "#00FF00"
	err = adminClient.UpdateAppearance(ctx, wantBanner)
	require.NoError(t, err)
	gotBanner, err := adminClient.Appearance(ctx)
	require.NoError(t, err)
	gotBanner.SupportLinks = nil // clean "support links" before comparison
	require.Equal(t, wantBanner, gotBanner)

	// But even an admin can't give a bad color
	wantBanner.ServiceBanner.BackgroundColor = "#bad color"
	err = adminClient.UpdateAppearance(ctx, wantBanner)
	require.Error(t, err)
}

func TestCustomSupportLinks(t *testing.T) {
	t.Parallel()

	supportLinks := []codersdk.LinkConfig{
		{
			Name:   "First link",
			Target: "http://first-link-1",
			Icon:   "chat",
		},
		{
			Name:   "Second link",
			Target: "http://second-link-2",
			Icon:   "bug",
		},
	}
	cfg := coderdtest.DeploymentConfig(t)
	cfg.Support = new(codersdk.SupportConfig)
	cfg.Support.Links = &codersdk.DeploymentConfigField[[]codersdk.LinkConfig]{
		Value: supportLinks,
	}

	client := coderdenttest.New(t, &coderdenttest.Options{
		Options: &coderdtest.Options{
			DeploymentConfig: cfg,
		},
	})
	coderdtest.CreateFirstUser(t, client)
	coderdenttest.AddLicense(t, client, coderdenttest.LicenseOptions{
		Features: license.Features{
			codersdk.FeatureAppearance: 1,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
	defer cancel()

	appearance, err := client.Appearance(ctx)
	require.NoError(t, err)

	require.Len(t, appearance.SupportLinks, 2)
	require.Equal(t, supportLinks, appearance.SupportLinks)
}

func TestDefaultSupportLinks(t *testing.T) {
	t.Parallel()

	client := coderdenttest.New(t, nil)
	coderdtest.CreateFirstUser(t, client)
	coderdenttest.AddLicense(t, client, coderdenttest.LicenseOptions{
		Features: license.Features{
			codersdk.FeatureAppearance: 1,
		},
	})

	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitMedium)
	defer cancel()

	appearance, err := client.Appearance(ctx)
	require.NoError(t, err)

	require.Len(t, appearance.SupportLinks, 3) // Documentation, Report a bug, Join the Coder Discord
	require.Equal(t, appearance.SupportLinks[0].Name, "Documentation")
	require.Equal(t, appearance.SupportLinks[1].Name, "Report a bug")
	require.Equal(t, appearance.SupportLinks[2].Name, "Join the Coder Discord")
}

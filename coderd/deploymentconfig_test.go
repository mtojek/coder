package coderd_test

import (
	"context"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/coder/coder/cli/config"
	"github.com/coder/coder/cli/deployment"
	"github.com/coder/coder/coderd/coderdtest"
	"github.com/coder/coder/testutil"
)

func TestDeploymentConfig(t *testing.T) {
	t.Parallel()
	hi := "hi"
	ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
	defer cancel()
	vip := deployment.NewViper()
	fs := pflag.NewFlagSet("test", pflag.ContinueOnError)
	fs.String(config.FlagName, hi, "usage")
	cfg, err := deployment.Config(fs, vip)
	require.NoError(t, err)
	// values should be returned
	cfg.AccessURL.Value = hi
	// values should not be returned
	cfg.OAuth2.Github.ClientSecret.Value = hi
	cfg.OIDC.ClientSecret.Value = hi
	cfg.PostgresURL.Value = hi
	cfg.SCIMAPIKey.Value = hi

	client := coderdtest.New(t, &coderdtest.Options{
		DeploymentConfig: cfg,
	})
	_ = coderdtest.CreateFirstUser(t, client)
	scrubbed, err := client.DeploymentConfig(ctx)
	require.NoError(t, err)
	// ensure normal values pass through
	require.EqualValues(t, hi, scrubbed.AccessURL.Value)
	// ensure secrets are removed
	require.Empty(t, scrubbed.OAuth2.Github.ClientSecret.Value)
	require.Empty(t, scrubbed.OIDC.ClientSecret.Value)
	require.Empty(t, scrubbed.PostgresURL.Value)
	require.Empty(t, scrubbed.SCIMAPIKey.Value)
}

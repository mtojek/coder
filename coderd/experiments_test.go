package coderd_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/coderd/coderdtest"
	"github.com/coder/coder/codersdk"
	"github.com/coder/coder/testutil"
)

func Test_Experiments(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		client := coderdtest.New(t, nil)

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		require.Empty(t, experiments)
	})

	t.Run("wildcard", func(t *testing.T) {
		t.Parallel()
		dc := coderdtest.DeploymentConfig(t)
		dc.Experimental = &codersdk.DeploymentConfigField[codersdk.ExperimentalConfig]{
			Value: []string{"*"},
		}
		client := coderdtest.New(t, &coderdtest.Options{
			DeploymentConfig: dc,
		})

		ctx, cancel := context.WithTimeout(context.Background(), testutil.WaitLong)
		defer cancel()

		experiments, err := client.Experiments(ctx)
		require.NoError(t, err)
		require.NotNil(t, experiments)
		require.ElementsMatch(t, codersdk.ExperimentsAll, experiments)
	})
}

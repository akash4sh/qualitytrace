package testsuite

import (
	"fmt"
	"testing"

	"github.com/intelops/qualitytrace/testing/cli-e2etest/environment"
	"github.com/intelops/qualitytrace/testing/cli-e2etest/helpers"
	"github.com/intelops/qualitytrace/testing/cli-e2etest/qualitytracecli"
	"github.com/stretchr/testify/require"
)

func TestDeleteTestSuite(t *testing.T) {
	// instantiate require with testing helper
	require := require.New(t)

	// setup isolated e2e environment
	env := environment.CreateAndStart(t)
	defer env.Close(t)

	cliConfig := env.GetCLIConfigPath(t)

	// Given I am a Tracetest CLI user
	// And I have my server recently created

	// When I try to delete a TestSuite that don't exist
	// Then it should return an error and say that this resource does not exist
	result := qualitytracecli.Exec(t, "delete testsuite --id dont-exist", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 1)
	require.Contains(result.StdErr, "Resource testsuite with ID dont-exist not found")

	// When I try to set up a new testsuite
	// Then it should be applied with success
	newTestSuitePath := env.GetTestResourcePath(t, "new-testsuite")

	result = qualitytracecli.Exec(t, fmt.Sprintf("apply testsuite --file %s", newTestSuitePath), qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)

	// When I try to delete the testsuite
	// Then it should delete with success
	result = qualitytracecli.Exec(t, "delete testsuite --id Qti5R3_VR", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)
	require.Contains(result.StdOut, "✔ Testsuite successfully deleted")

	// When I try to get a TestSuite again
	// Then it should return a message saying that the testsuite was not found
	result = qualitytracecli.Exec(t, "delete testsuite --id Qti5R3_VR", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 1)
	require.Contains(result.StdErr, "Resource testsuite with ID Qti5R3_VR not found")
}

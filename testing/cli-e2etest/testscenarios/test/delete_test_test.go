package test

import (
	"fmt"
	"testing"

	"github.com/intelops/qualitytrace/testing/cli-e2etest/environment"
	"github.com/intelops/qualitytrace/testing/cli-e2etest/helpers"
	"github.com/intelops/qualitytrace/testing/cli-e2etest/qualitytracecli"
	"github.com/intelops/qualitytrace/testing/cli-e2etest/testscenarios/types"
	"github.com/stretchr/testify/require"
)

func TestDeleteTest(t *testing.T) {
	// instantiate require with testing helper
	require := require.New(t)

	// setup isolated e2e environment
	env := environment.CreateAndStart(t)
	defer env.Close(t)

	cliConfig := env.GetCLIConfigPath(t)

	// Given I am a Tracetest CLI user
	// And I have my server recently created

	// When I try to delete an test that don't exist
	// Then it should return an error and say that this resource does not exist
	result := qualitytracecli.Exec(t, "delete test --id .env", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 1)
	require.Contains(result.StdErr, "Resource test with ID .env not found")

	// When I try to set up a new test
	// Then it should be applied with success
	newTestPath := env.GetTestResourcePath(t, "list")

	result = qualitytracecli.Exec(t, fmt.Sprintf("apply test --file %s", newTestPath), qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)

	testVars := helpers.UnmarshalYAML[types.TestResource](t, result.StdOut)
	require.Equal("Test", testVars.Type)
	require.Equal("fH_8AulVR", testVars.Spec.ID)

	// When I try to delete the test
	// Then it should delete with success
	result = qualitytracecli.Exec(t, "delete test --id fH_8AulVR", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)
	require.Contains(result.StdOut, "✔ Test successfully deleted")

	// When I try to get an test again
	// Then it should return a message saying that the test was not found
	result = qualitytracecli.Exec(t, "get test --id fH_8AulVR", qualitytracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)
	require.Contains(result.StdOut, "Resource test with ID fH_8AulVR not found")
}

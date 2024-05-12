package testrunner

import (
	"fmt"
	"testing"

	"github.com/intelops/qualityTrace/testing/cli-e2etest/environment"
	"github.com/intelops/qualityTrace/testing/cli-e2etest/helpers"
	"github.com/intelops/qualityTrace/testing/cli-e2etest/qualityTracecli"
	"github.com/intelops/qualityTrace/testing/cli-e2etest/testscenarios/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyTestRunner(t *testing.T) {
	// instantiate require with testing helper
	require := require.New(t)
	assert := assert.New(t)

	// setup isolated e2e environment
	env := environment.CreateAndStart(t)
	defer env.Close(t)

	cliConfig := env.GetCLIConfigPath(t)

	// Given I am a Tracetest CLI user
	// And I have my server recently created

	// When I try to set up a new testRunner
	// Then it should be applied with success
	testRunnerPath := env.GetTestResourcePath(t, "new-testrunner")

	result := qualityTracecli.Exec(t, fmt.Sprintf("apply testrunner --file %s", testRunnerPath), qualityTracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)

	// When I try to get a testRunner again
	// Then it should return the testRunner applied on the last step, with analytics disabled
	result = qualityTracecli.Exec(t, "get testrunner --id current", qualityTracecli.WithCLIConfig(cliConfig))
	helpers.RequireExitCodeEqual(t, result, 0)

	testRunner := helpers.UnmarshalYAML[types.TestRunnerResource](t, result.StdOut)
	assert.Equal("TestRunner", testRunner.Type)
	assert.Equal("current", testRunner.Spec.ID)
	assert.Equal("default", testRunner.Spec.Name)
	require.Len(testRunner.Spec.RequiredGates, 2)
	assert.Equal("analyzer-score", testRunner.Spec.RequiredGates[0])
	assert.Equal("test-specs", testRunner.Spec.RequiredGates[1])
}
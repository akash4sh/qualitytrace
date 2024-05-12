package testscenarios

import (
	"testing"

	"github.com/intelops/qualityTrace/testing/cli-e2etest/helpers"
	"github.com/intelops/qualityTrace/testing/cli-e2etest/qualityTracecli"
	"github.com/stretchr/testify/require"
)

func TestHelpCommand(t *testing.T) {
	// instantiate require with testing helper
	require := require.New(t)

	// Given I am a Tracetest CLI user
	// When I try to get help with the commands "qualityTrace help", "qualityTrace --help" or "qualityTrace -h"
	// Then I should receive a message with sucess

	possibleCommands := []string{"help", "--help", "-h"}

	for _, helpCommand := range possibleCommands {
		result := qualityTracecli.Exec(t, helpCommand)
		helpers.RequireExitCodeEqual(t, result, 0)
		require.Greater(len(result.StdOut), 0)
	}
}

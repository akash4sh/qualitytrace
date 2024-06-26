package filters_test

import (
	"testing"

	"github.com/intelops/qualitytrace/server/expression/filters"
	"github.com/intelops/qualitytrace/server/expression/value"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexGroup(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          string
		Regex          string
		ExpectedOutput string
	}{
		{
			Name:           "should_be_able_to_extract_one_group",
			Input:          `{ "id": 38, "name": "Qualitytrace" }`,
			Regex:          `"id": (\d+)`,
			ExpectedOutput: `38`,
		},
		{
			Name:           "should_be_able_to_extract_one_group_multiple_times",
			Input:          `[{ "id": 38, "name": "Qualitytrace" }, { "id": 39, "name": "Kusk" }]`,
			Regex:          `"id": (\d+)`,
			ExpectedOutput: `[38, 39]`,
		},
		{
			Name:           "should_be_able_to_extract_multiple_groups_multiple_times",
			Input:          `[{ "id": 38, "name": "Qualitytrace" }, { "id": 39, "name": "Kusk" }]`,
			Regex:          `"id": (\d+), "name": "(\w+)"`,
			ExpectedOutput: `[38, "Qualitytrace", 39, "Kusk"]`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			input := value.NewFromString(testCase.Input)
			output, err := filters.RegexGroup(input, testCase.Regex)
			require.NoError(t, err)

			assert.Equal(t, testCase.ExpectedOutput, output.String())
		})
	}
}

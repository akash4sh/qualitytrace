package filters

import (
	"fmt"
	"regexp"

	"github.com/intelops/qualitytrace/server/expression/types"
	"github.com/intelops/qualitytrace/server/expression/value"
)

func RegexGroup(input value.Value, args ...string) (value.Value, error) {
	if len(args) != 1 {
		return value.Value{}, fmt.Errorf("wrong number of args. Expected 1, got %d", len(args))
	}

	if input.IsArray() {
		return value.Value{}, fmt.Errorf("cannot process array of json objects")
	}

	regex, err := regexp.Compile(args[0])
	if err != nil {
		return value.Value{}, fmt.Errorf("invalid regex: %w", err)
	}

	groups := regex.FindAllStringSubmatch(input.Value().Value, -1)
	if groups == nil {
		return value.NewArray([]types.TypedValue{}), nil
	}

	output := make([]string, 0)
	for _, group := range groups {
		output = append(output, group[1:]...)
	}

	if len(output) == 1 {
		typedValue := types.GetTypedValue(output[0])
		return value.New(typedValue), nil
	}

	return value.NewArrayFromStrings(output), nil
}

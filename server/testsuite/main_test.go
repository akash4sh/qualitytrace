package testsuite_test

import (
	"os"
	"testing"

	"github.com/intelops/qualitytrace/server/testmock"
)

func TestMain(m *testing.M) {
	testmock.StartTestEnvironment()

	exitVal := m.Run()

	testmock.StopTestEnvironment()

	os.Exit(exitVal)
}

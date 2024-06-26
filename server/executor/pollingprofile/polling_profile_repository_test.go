package pollingprofile_test

import (
	"database/sql"
	"testing"

	"github.com/gorilla/mux"
	"github.com/intelops/qualitytrace/server/executor/pollingprofile"
	"github.com/intelops/qualitytrace/server/pkg/id"
	"github.com/intelops/qualitytrace/server/resourcemanager"
	rmtests "github.com/intelops/qualitytrace/server/resourcemanager/testutil"
)

func TestPollingProfileResource(t *testing.T) {
	rmtests.TestResourceType(t, rmtests.ResourceTypeTest{
		ResourceTypeSingular: pollingprofile.ResourceName,
		ResourceTypePlural:   pollingprofile.ResourceNamePlural,
		RegisterManagerFn: func(router *mux.Router, db *sql.DB) resourcemanager.Manager {
			pollingProfileRepo := pollingprofile.NewRepository(db)

			manager := resourcemanager.New[pollingprofile.PollingProfile](
				pollingprofile.ResourceName,
				pollingprofile.ResourceNamePlural,
				pollingProfileRepo,
				resourcemanager.DisableDelete(),
				resourcemanager.WithIDGen(id.GenerateID),
			)
			manager.RegisterRoutes(router)

			return manager
		},
		SampleJSON: `{
			"type": "PollingProfile",
			"spec": {
				"id": "current",
				"name": "default",
				"default": true,
				"strategy": "periodic",
				"periodic": {
					"timeout": "1m",
					"retryDelay": "5s",
					"selectorMatchRetries": 3
				}
			}
		}`,
		SampleJSONUpdated: `{
			"type": "PollingProfile",
			"spec": {
				"id": "current",
				"name": "long test",
				"default": true,
				"strategy": "periodic",
				"periodic": {
					"timeout": "1h",
					"retryDelay": "25s",
					"selectorMatchRetries": 3
				}
			}
		}`,
	},
		rmtests.ExcludeOperations(
			rmtests.OperationGetNotFound,
			rmtests.OperationUpdateNotFound,
			rmtests.OperationListSortSuccess,
			rmtests.OperationListNoResults,
		))
}

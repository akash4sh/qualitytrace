package test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/intelops/qualitytrace/server/executor/testrunner"
	"github.com/intelops/qualitytrace/server/pkg/id"
	"github.com/intelops/qualitytrace/server/pkg/sqlutil"
	"github.com/intelops/qualitytrace/server/variableset"
	"go.opentelemetry.io/otel/trace"
)

type RunRepository interface {
	CreateRun(context.Context, Test, Run) (Run, error)
	UpdateRun(context.Context, Run) error
	DeleteRun(context.Context, Run) error
	GetRun(_ context.Context, testID id.ID, runID int) (Run, error)
	GetTestRuns(_ context.Context, _ Test, take, skip int32) ([]Run, error)
	Count(context.Context, Test) (int, error)
	GetRunByTraceID(context.Context, trace.TraceID) (Run, error)
	GetLatestRunByTestVersion(context.Context, id.ID, int) (Run, error)

	GetTestSuiteRunSteps(_ context.Context, _ id.ID, runID int) ([]Run, error)
}

type runRepository struct {
	db *sql.DB
}

func NewRunRepository(db *sql.DB) RunRepository {
	return &runRepository{
		db: db,
	}
}

const createRunQuery = `
INSERT INTO test_runs (
	"id",
	"test_id",
	"test_version",

	-- timestamps
	"created_at",
	"service_triggered_at",
	"service_trigger_completed_at",
	"obtained_trace_at",
	"completed_at",

	-- trigger params
	"state",
	"trace_id",
	"span_id",

	-- result info
	"resolved_trigger",
	"trigger_results",
	"test_results",
	"trace",
	"outputs",
	"last_error",
	"pass",
	"fail",

	"metadata",

	-- variable set
	"variable_set",

	-- linter
	"linter",

	-- required gates
	"required_gates_result",

	"skip_trace_collection",

	"tenant_id"
) VALUES (
	nextval('` + runSequenceName + `'), -- id
	$1,   -- test_id
	$2,   -- test_version

	-- timestamps
	$3,              -- created_at
	$4,              -- service_triggered_at
	$5,              -- service_trigger_completed_at
	$6,              -- obtained_trace_at
	to_timestamp(0), -- completed_at

	-- trigger params
	$7, -- state
	$8, -- trace_id
	$9, -- span_id

	-- result info
	$10, -- resolved_trigger
	$11,  -- trigger_results
	'{}', -- test_results
	$12,  -- trace
	'[]', -- outputs
	NULL, -- last_error
	0,    -- pass
	0,    -- fail

	$13, -- metadata
	$14, -- variable_set
	$15, -- linter
	$16,  -- required_gates_result
	$17, -- skip_trace_collection
	$18  -- tenant_id
)
RETURNING "id"`

func (r *runRepository) CreateRun(ctx context.Context, test Test, run Run) (Run, error) {
	run.TestID = test.ID
	run.State = RunStateCreated
	run.TestVersion = test.SafeVersion()
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now()
	}

	encodedRun, err := EncodeRun(run)
	if err != nil {
		return Run{}, fmt.Errorf("cannot encode run: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, fmt.Errorf("sql beginTx: %w", err)
	}

	tenantID := sqlutil.TenantIDString(ctx)

	_, err = tx.ExecContext(ctx, replaceRunSequenceName(createSequenceQuery, test.ID, tenantID))
	if err != nil {
		tx.Rollback()
		return Run{}, fmt.Errorf("sql exec: %w", err)
	}

	params := sqlutil.TenantInsert(ctx,
		test.ID,
		test.SafeVersion(),
		run.CreatedAt,
		run.ServiceTriggeredAt,
		run.ServiceTriggerCompletedAt,
		run.ObtainedTraceAt,
		run.State,
		run.TraceID.String(),
		run.SpanID.String(),
		encodedRun.JsonResolvedTrigger,
		encodedRun.JsonTriggerResults,
		encodedRun.JsonTrace,
		encodedRun.JsonMetadata,
		encodedRun.JsonVariableSet,
		encodedRun.JsonLinter,
		encodedRun.JsonGatesResult,
		run.SkipTraceCollection,
	)

	var runID int
	err = tx.QueryRowContext(ctx, replaceRunSequenceName(createRunQuery, test.ID, tenantID), params...).Scan(&runID)
	if err != nil {
		tx.Rollback()
		return Run{}, fmt.Errorf("sql exec: %w", err)
	}

	tx.Commit()

	return r.GetRun(ctx, test.ID, runID)
}

const updateRunQuery = `
UPDATE test_runs SET

	-- timestamps
	"service_triggered_at" = $1,
	"service_trigger_completed_at" = $2,
	"obtained_trace_at" = $3,
	"completed_at" = $4,

	-- trigger params
	"state" = $5,
	"trace_id" = $6,
	"span_id" = $7,

	-- result info
	"resolved_trigger" = $8,
	"trigger_results" = $9,
	"test_results" = $10,
	"trace" = $11,
	"outputs" = $12,
	"last_error" = $13,
	"pass" = $14,
	"fail" = $15,

	"metadata" = $16,
	"variable_set" = $19,

	--- linter
	"linter" = $20,

	--- required gates
	"required_gates_result" = $21

WHERE id = $17 AND test_id = $18
`

func (r *runRepository) UpdateRun(ctx context.Context, run Run) error {
	encodedRun, err := EncodeRun(run)
	if err != nil {
		return fmt.Errorf("cannot encode run: %w", err)
	}

	pass, fail := run.ResultsCount()

	query, params := sqlutil.Tenant(
		ctx,
		updateRunQuery,
		run.ServiceTriggeredAt,
		run.ServiceTriggerCompletedAt,
		run.ObtainedTraceAt,
		run.CompletedAt,
		run.State,
		encodedRun.TraceID,
		encodedRun.SpanID,
		encodedRun.JsonResolvedTrigger,
		encodedRun.JsonTriggerResults,
		encodedRun.JsonTestResults,
		encodedRun.JsonTrace,
		encodedRun.JsonOutputs,
		encodedRun.EncodedLastError,
		pass,
		fail,
		encodedRun.JsonMetadata,
		run.ID,
		run.TestID,
		encodedRun.JsonVariableSet,
		encodedRun.JsonLinter,
		encodedRun.JsonGatesResult,
	)

	_, err = r.db.ExecContext(
		ctx,
		query,
		params...,
	)
	if err != nil {
		return fmt.Errorf("sql exec: %w", err)
	}

	return nil
}

func (r *runRepository) DeleteRun(ctx context.Context, run Run) error {
	queries := []string{
		"DELETE FROM test_suite_run_steps WHERE test_run_id = $1 AND test_run_test_id = $2",
		"DELETE FROM test_runs WHERE id = $1 AND test_id = $2",
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("sql BeginTx: %w", err)
	}

	for _, sql := range queries {
		query, params := sqlutil.Tenant(ctx, sql, run.ID, run.TestID)
		_, err := tx.ExecContext(ctx, query, params...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("sql error: %w", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("sql Commit: %w", err)
	}

	return nil
}

const (
	fields = `
	"id",
	"test_id",
	"test_version",

	-- timestamps
	"created_at",
	"service_triggered_at",
	"service_trigger_completed_at",
	"obtained_trace_at",
	"completed_at",

	-- trigger params
	"state",
	"trace_id",
	"span_id",

	-- result info
	"resolved_trigger",
	"trigger_results",
	"test_results",
	"trace",
	"outputs",
	"last_error",
	"metadata",
	"variable_set",

	-- test_suite run
	test_suite_run_steps.test_suite_run_id,
	test_suite_run_steps.test_suite_run_test_suite_id,
	"linter",
	"required_gates_result",
	"skip_trace_collection"
`

	baseSql = `
SELECT
	%s
FROM
	test_runs
LEFT OUTER JOIN test_suite_run_steps
ON test_suite_run_steps.test_run_id = test_runs.id AND test_suite_run_steps.test_run_test_id = test_runs.test_id
`
)

var (
	selectRunQuery = fmt.Sprintf(baseSql, fields)
	countRunQuery  = fmt.Sprintf(baseSql, "COUNT(*)")
)

func (r *runRepository) GetRun(ctx context.Context, testID id.ID, runID int) (Run, error) {
	query, params := sqlutil.TenantWithPrefix(ctx, selectRunQuery+" WHERE id = $1 AND test_id = $2", "test_runs.", runID, testID)

	run, err := readRunRow(r.db.QueryRowContext(ctx, query, params...))
	if err != nil {
		return Run{}, fmt.Errorf("cannot read row: %w", err)
	}

	return run, nil
}

func (r *runRepository) GetTestRuns(ctx context.Context, test Test, take, skip int32) ([]Run, error) {
	query, params := sqlutil.TenantWithPrefix(ctx, selectRunQuery+" WHERE test_id = $1", "test_runs.", test.ID, take, skip)
	stmt, err := r.db.Prepare(query + " ORDER BY created_at DESC LIMIT $2 OFFSET $3")
	if err != nil {
		return []Run{}, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, params...)
	if err != nil {
		return []Run{}, err
	}

	runs, err := r.readRunRows(ctx, rows)
	if err != nil {
		return []Run{}, err
	}

	return runs, nil
}

func (r *runRepository) Count(ctx context.Context, test Test) (int, error) {
	query, params := sqlutil.TenantWithPrefix(ctx, countRunQuery+" WHERE test_id = $1", "test_runs.", test.ID)
	count := 0

	err := r.db.
		QueryRowContext(ctx, query, params...).
		Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("sql query: %w", err)
	}

	return count, nil
}

func (r *runRepository) GetRunByTraceID(ctx context.Context, traceID trace.TraceID) (Run, error) {
	query, params := sqlutil.Tenant(ctx, selectRunQuery+" WHERE trace_id = $1", traceID.String())
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return Run{}, err
	}
	defer stmt.Close()

	run, err := readRunRow(stmt.QueryRowContext(ctx, params...))
	if err != nil {
		return Run{}, fmt.Errorf("cannot read row: %w", err)
	}
	return run, nil
}

func (r *runRepository) GetLatestRunByTestVersion(ctx context.Context, testID id.ID, version int) (Run, error) {
	query, params := sqlutil.TenantWithPrefix(ctx, selectRunQuery+" WHERE test_id = $1 AND test_version = $2", "test_runs.", testID.String(), version)
	stmt, err := r.db.Prepare(query + " ORDER BY created_at DESC LIMIT 1")

	if err != nil {
		return Run{}, err
	}
	defer stmt.Close()

	run, err := readRunRow(stmt.QueryRowContext(ctx, params...))
	if err != nil {
		return Run{}, err
	}
	return run, nil
}

func (r *runRepository) readRunRows(ctx context.Context, rows *sql.Rows) ([]Run, error) {
	var runs []Run

	for rows.Next() {
		run, err := readRunRow(rows)
		if err != nil {
			return []Run{}, fmt.Errorf("cannot read row: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func readRunRow(row scanner) (Run, error) {
	encodedRun := EncodedRun{}

	var lastErr *string

	err := row.Scan(
		&encodedRun.ID,
		&encodedRun.TestID,
		&encodedRun.TestVersion,
		&encodedRun.CreatedAt,
		&encodedRun.ServiceTriggeredAt,
		&encodedRun.ServiceTriggerCompletedAt,
		&encodedRun.ObtainedTraceAt,
		&encodedRun.CompletedAt,
		&encodedRun.State,
		&encodedRun.TraceID,
		&encodedRun.SpanID,
		&encodedRun.JsonResolvedTrigger,
		&encodedRun.JsonTriggerResults,
		&encodedRun.JsonTestResults,
		&encodedRun.JsonTrace,
		&encodedRun.JsonOutputs,
		&lastErr,
		&encodedRun.JsonMetadata,
		&encodedRun.JsonVariableSet,
		&encodedRun.TestSuiteRunID,
		&encodedRun.TestSuiteID,
		&encodedRun.JsonLinter,
		&encodedRun.JsonGatesResult,
		&encodedRun.SkipTraceCollection,
	)

	if err != nil {
		return Run{}, err
	}

	if lastErr != nil {
		encodedRun.EncodedLastError = *lastErr
	}

	return encodedRun.ToRun()
}

func (r *runRepository) GetTestSuiteRunSteps(ctx context.Context, id id.ID, runID int) ([]Run, error) {
	query := selectRunQuery + `
WHERE test_suite_run_steps.test_suite_run_id = $1 AND test_suite_run_steps.test_suite_run_test_suite_id = $2
`
	query, params := sqlutil.TenantWithPrefix(ctx, query, "test_runs.", strconv.Itoa(runID), id)
	query += ` ORDER BY test_runs.created_at ASC`

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return []Run{}, fmt.Errorf("prepare: %w", err)
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, params...)
	if err != nil {
		return []Run{}, fmt.Errorf("query context: %w", err)
	}

	steps, err := r.readRunRows(ctx, rows)
	if err != nil {
		return []Run{}, fmt.Errorf("read row: %w", err)
	}

	return steps, nil
}

type EncodedRun struct {
	ID                  int
	TestID              string
	TestVersion         int
	State               string
	Pass                int
	Fail                int
	SkipTraceCollection bool

	// Timestamps
	CreatedAt                 time.Time
	ServiceTriggeredAt        time.Time
	ServiceTriggerCompletedAt time.Time
	ObtainedTraceAt           time.Time
	CompletedAt               time.Time

	JsonResolvedTrigger,
	JsonTriggerResults,
	JsonTestResults,
	JsonTrace,
	JsonOutputs,
	JsonVariableSet,
	JsonLinter,
	JsonGatesResult,
	JsonMetadata []byte

	EncodedLastError string
	TraceID,
	SpanID string

	TestSuiteID,
	TestSuiteRunID sql.NullString
}

func (er EncodedRun) ToRun() (Run, error) {
	r := Run{}
	err := json.Unmarshal(er.JsonTriggerResults, &r.TriggerResult)
	if err != nil {
		return Run{}, fmt.Errorf("cannot parse TriggerResult: %w", err)
	}

	err = json.Unmarshal(er.JsonResolvedTrigger, &r.ResolvedTrigger)
	if err != nil {
		return Run{}, fmt.Errorf("cannot parse ResolvedTrigger: %w", err)
	}

	err = json.Unmarshal(er.JsonTestResults, &r.Results)
	if err != nil {
		return Run{}, fmt.Errorf("cannot parse Results: %w", err)
	}

	if er.JsonTrace != nil && string(er.JsonTrace) != "null" {
		err = json.Unmarshal(er.JsonTrace, &r.Trace)
		if err != nil {
			return Run{}, fmt.Errorf("cannot parse Trace: %w", err)
		}
	}

	if er.JsonLinter != nil {
		err = json.Unmarshal(er.JsonLinter, &r.Linter)
		if err != nil {
			return Run{}, fmt.Errorf("cannot parse linter: %w", err)
		}
	}

	err = json.Unmarshal(er.JsonOutputs, &r.Outputs)
	if err != nil {
		// try with raw outputs
		var rawOutputs []variableset.VariableSetValue
		err = json.Unmarshal(er.JsonOutputs, &rawOutputs)

		for _, value := range rawOutputs {
			r.Outputs.Add(value.Key, RunOutput{
				Name:   value.Key,
				Value:  value.Value,
				SpanID: "",
			})
		}

		if err != nil {
			return Run{}, fmt.Errorf("cannot parse Outputs: %w", err)
		}
	}

	err = json.Unmarshal(er.JsonMetadata, &r.Metadata)
	if err != nil {
		return Run{}, fmt.Errorf("cannot parse Metadata: %w", err)
	}

	err = json.Unmarshal(er.JsonVariableSet, &r.VariableSet)
	if err != nil {
		return Run{}, fmt.Errorf("cannot parse VariableSet: %w", err)
	}

	if er.JsonGatesResult != nil {
		err = json.Unmarshal(er.JsonGatesResult, &r.RequiredGatesResult)
		if err != nil {
			return Run{}, fmt.Errorf("cannot parse required gates result: %w", err)
		}
	} else {
		// fallback for retro-compatibility
		r.RequiredGatesResult = r.GenerateRequiredGateResult(testrunner.DefaultTestRunner.RequiredGates)
	}

	if er.TraceID != "" {
		r.TraceID, err = trace.TraceIDFromHex(er.TraceID)
		if err != nil {
			return Run{}, fmt.Errorf("cannot parse TraceID: %w", err)
		}
	}

	if er.SpanID != "" {
		r.SpanID, err = trace.SpanIDFromHex(er.SpanID)
		if err != nil {
			return Run{}, fmt.Errorf("cannot parse SpanID: %w", err)
		}
	}

	if er.EncodedLastError != "" {
		r.LastError = fmt.Errorf(er.EncodedLastError)
	}

	if er.TestSuiteID.Valid && er.TestSuiteRunID.Valid {
		r.TestSuiteID = er.TestSuiteID.String
		r.TestSuiteRunID = er.TestSuiteRunID.String
	}

	r.ID = er.ID
	r.TestID = id.ID(er.TestID)
	r.TestVersion = er.TestVersion
	r.CreatedAt = er.CreatedAt
	r.ServiceTriggeredAt = er.ServiceTriggeredAt
	r.ServiceTriggerCompletedAt = er.ServiceTriggerCompletedAt
	r.ObtainedTraceAt = er.ObtainedTraceAt
	r.CompletedAt = er.CompletedAt
	r.State = RunState(er.State)
	r.SkipTraceCollection = er.SkipTraceCollection

	return r, nil
}

func EncodeRun(run Run) (EncodedRun, error) {
	er := EncodedRun{
		TraceID:             run.TraceID.String(),
		SpanID:              run.SpanID.String(),
		ID:                  run.ID,
		TestID:              string(run.TestID),
		TestVersion:         run.TestVersion,
		State:               string(run.State),
		Pass:                run.Pass,
		Fail:                run.Fail,
		SkipTraceCollection: run.SkipTraceCollection,
	}

	var err error
	er.JsonResolvedTrigger, err = json.Marshal(run.ResolvedTrigger)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("resolved trigger encoding error: %w", err)
	}

	er.JsonTriggerResults, err = json.Marshal(run.TriggerResult)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("trigger results encoding error: %w", err)
	}

	er.JsonTestResults, err = json.Marshal(run.Results)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("test results encoding error: %w", err)
	}

	er.JsonTrace, err = json.Marshal(run.Trace)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("trace encoding error: %w", err)
	}

	er.JsonOutputs, err = json.Marshal(run.Outputs)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("outputs encoding error: %w", err)
	}

	er.JsonMetadata, err = json.Marshal(run.Metadata)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("encoding error: %w", err)
	}

	er.JsonVariableSet, err = json.Marshal(run.VariableSet)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("encoding error: %w", err)
	}

	er.JsonLinter, err = json.Marshal(run.Linter)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("encoding error: %w", err)
	}

	er.JsonGatesResult, err = json.Marshal(run.RequiredGatesResult)
	if err != nil {
		return EncodedRun{}, fmt.Errorf("encoding error: %w", err)
	}

	if run.LastError != nil {
		er.EncodedLastError = run.LastError.Error()
	}

	return er, nil
}

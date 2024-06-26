package types

import (
	"time"

	"github.com/intelops/qualitytrace/server/pkg/maps"
)

type TestResource struct {
	Type string `json:"type"`
	Spec Test   `json:"spec"`
}

type Test struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Version     int                          `json:"version"`
	Trigger     Trigger                      `json:"trigger"`
	Specs       []TestSpec                   `json:"specs"`
	Outputs     maps.Ordered[string, Output] `json:"outputs"`
	Summary     Summary                      `json:"summary"`
}

type Trigger struct {
	Type        string       `json:"type"`
	HTTPRequest *HTTPRequest `json:"httpRequest"`
	GRPCRequest *GRPCRequest `json:"grpc"`
}

type GRPCRequest struct {
	ProtobufFile string `json:"protobufFile,omitempty"`
}

type HTTPRequest struct {
	Method  string       `json:"method,omitempty"`
	URL     string       `json:"url"`
	Body    string       `json:"body,omitempty"`
	Headers []HTTPHeader `json:"headers,omitempty"`
}

type HTTPHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Output struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

type TestSpec struct {
	Selector   string   `json:"selector"`
	Name       string   `json:"name,omitempty"`
	Assertions []string `json:"assertions"`
}

type SpanSelector struct {
	Filters       []SelectorFilter     `json:"filters"`
	PseudoClass   *SelectorPseudoClass `json:"pseudoClass,omitempty"`
	ChildSelector *SpanSelector        `json:"childSelector,omitempty"`
}

type SelectorFilter struct {
	Property string `json:"property"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type SelectorPseudoClass struct {
	Name     string `json:"name"`
	Argument *int32 `json:"argument,omitempty"`
}

type Summary struct {
	Runs    int     `json:"runs"`
	LastRun LastRun `json:"lastRun"`
}

type LastRun struct {
	Time   time.Time `json:"time"`
	Passes int       `json:"passes"`
	Fails  int       `json:"fails"`
}

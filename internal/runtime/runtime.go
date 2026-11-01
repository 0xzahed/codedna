// Package runtime aggregates and analyzes function execution events captured
// at runtime, comparing observed execution against static code relationships.
// Runtime instrumentation is opt-in and NEVER automatically executes untrusted code.
package runtime

import (
	"math"
	"sort"
	"strings"

	"github.com/codedna/codedna/internal/model"
)

// edgeKey tracks dynamic caller -> callee pairs.
type edgeKey struct {
	from string
	to   string
}

// IngestTraces processes a stream of trace events into aggregated analytics.
func IngestTraces(events []model.RuntimeTraceEvent, staticRels []model.Relationship) *model.RuntimeSummary {
	summary := &model.RuntimeSummary{
		TotalEvents: len(events),
	}

	if len(events) == 0 {
		return summary
	}

	// 1. Function frequency and duration analysis
	type fnStats struct {
		name      string
		file      string
		count     int
		totalTime int64
	}
	fnMap := make(map[string]*fnStats)
	var totalDuration int64

	// 2. Track execution paths per run_id
	pathsByRun := make(map[string][]model.RuntimeTraceEvent)

	// 3. Dynamic edges discovered
	runtimeEdgeMap := make(map[edgeKey]int)

	for _, e := range events {
		totalDuration += e.DurationUs
		key := e.File + ":" + e.FunctionName
		if _, exists := fnMap[key]; !exists {
			fnMap[key] = &fnStats{
				name: e.FunctionName,
				file: e.File,
			}
		}
		fnMap[key].count++
		fnMap[key].totalTime += e.DurationUs

		// Group by run id
		runID := e.RunID
		if runID == "" {
			runID = "default"
		}
		pathsByRun[runID] = append(pathsByRun[runID], e)

		// Discover dynamic caller -> callee edge
		if e.CallerFunction != "" {
			callerKey := e.CallerFile + ":" + e.CallerFunction
			calleeKey := e.File + ":" + e.FunctionName
			runtimeEdgeMap[edgeKey{from: callerKey, to: calleeKey}]++
		}
	}

	summary.UniqueFunctions = len(fnMap)
	if len(events) > 0 {
		summary.AvgDurationUs = math.Round(float64(totalDuration)/float64(len(events))*100) / 100
	}

	// Build FrequentFunctions sorted by count descending
	for _, st := range fnMap {
		avg := 0.0
		if st.count > 0 {
			avg = math.Round(float64(st.totalTime)/float64(st.count)*100) / 100
		}
		summary.FrequentFunctions = append(summary.FrequentFunctions, model.FunctionFrequency{
			Function:    st.name,
			File:        st.file,
			CallCount:   st.count,
			TotalTimeUs: st.totalTime,
			AvgTimeUs:   avg,
		})
	}
	sort.Slice(summary.FrequentFunctions, func(i, j int) bool {
		return summary.FrequentFunctions[i].CallCount > summary.FrequentFunctions[j].CallCount
	})

	// Build ExecutionPaths
	for runID, traceEvents := range pathsByRun {
		// Sort trace events by timestamp
		sort.Slice(traceEvents, func(i, j int) bool {
			return traceEvents[i].TimestampNs < traceEvents[j].TimestampNs
		})

		var steps []string
		var runTotalTime int64
		root := ""
		if len(traceEvents) > 0 {
			root = traceEvents[0].FunctionName
		}

		for _, step := range traceEvents {
			steps = append(steps, step.FunctionName+" ("+step.File+")")
			runTotalTime += step.DurationUs
		}

		summary.ExecutionPaths = append(summary.ExecutionPaths, model.ExecutionPath{
			TraceID:   runID,
			Root:      root,
			Steps:     steps,
			TotalTime: runTotalTime,
		})
	}

	// Build runtime edges
	for k := range runtimeEdgeMap {
		summary.RuntimeEdges = append(summary.RuntimeEdges, model.GraphEdge{
			ID:                "runtime:" + k.from + "->" + k.to,
			Source:            k.from,
			Target:            k.to,
			Kind:              "runtime_called",
			Confidence:        1.0,
			Method:            "observed-execution",
			IsRuntimeObserved: true,
		})
	}

	// 4. Compare Static vs Runtime Graph
	summary.Comparison = compareStaticVsRuntime(staticRels, runtimeEdgeMap)

	return summary
}

func compareStaticVsRuntime(staticRels []model.Relationship, runtimeEdges map[edgeKey]int) model.RuntimeComparison {
	comp := model.RuntimeComparison{
		StaticEdgeCount:  len(staticRels),
		RuntimeEdgeCount: len(runtimeEdges),
	}

	observedStatic := 0
	unexecutedStatic := 0

	// Check which static edges were observed at runtime
	for _, sRel := range staticRels {
		if sRel.Kind != model.RelCalls && sRel.Kind != model.RelInstantiates {
			continue
		}

		matched := false
		for rEdge := range runtimeEdges {
			// Compare symbol IDs or function names
			if strings.Contains(rEdge.from, sRel.FromID) || strings.Contains(sRel.FromID, rEdge.from) {
				if strings.Contains(rEdge.to, sRel.ToID) || strings.Contains(sRel.ToID, rEdge.to) ||
					(sRel.ToName != "" && strings.Contains(rEdge.to, sRel.ToName)) {
					matched = true
					break
				}
			}
		}

		if matched {
			observedStatic++
		} else {
			unexecutedStatic++
		}
	}

	comp.ObservedStaticEdges = observedStatic
	comp.UnexecutedStaticEdges = unexecutedStatic

	dynamicOnly := 0
	for rEdge := range runtimeEdges {
		foundInStatic := false
		for _, sRel := range staticRels {
			if strings.Contains(rEdge.from, sRel.FromID) && (strings.Contains(rEdge.to, sRel.ToID) || strings.Contains(rEdge.to, sRel.ToName)) {
				foundInStatic = true
				break
			}
		}
		if !foundInStatic {
			dynamicOnly++
		}
	}
	comp.DynamicOnlyEdges = dynamicOnly

	if len(staticRels) > 0 {
		pct := (float64(observedStatic) / float64(len(staticRels))) * 100.0
		comp.CoveragePercentage = math.Round(pct*10) / 10
	}

	return comp
}

// TracerSnippet provides drop-in tracer snippets for languages.
func TracerSnippet(lang string) string {
	switch strings.ToLower(lang) {
	case "python":
		return `# CodeDNA Runtime Tracer Adapter (Python)
# Opt-in instrumentation: drop this into your test suite or startup code.
import time, sys, requests

API_URL = "http://localhost:8080/api/runtime/events"
events = []

def tracer(frame, event, arg):
    if event == "call":
        func_name = frame.f_code.co_name
        filename = frame.f_code.co_filename
        line = frame.f_lineno
        caller = frame.f_back.f_code.co_name if frame.f_back else ""
        caller_file = frame.f_back.f_code.co_filename if frame.f_back else ""
        events.append({
            "runId": "py-trace",
            "functionName": func_name,
            "file": filename,
            "line": line,
            "durationUs": 0,
            "timestampNs": time.time_ns(),
            "callerFunction": caller,
            "callerFile": caller_file
        })
        if len(events) >= 100:
            try:
                requests.post(API_URL, json=events, timeout=1)
                events.clear()
            except Exception:
                pass
    return tracer

# sys.setprofile(tracer)
`
	case "javascript", "typescript":
		return `// CodeDNA Runtime Tracer Adapter (Node.js)
// Opt-in instrumentation: wraps functions to record execution paths
const http = require('http');

let events = [];
function recordEvent(functionName, file, line, callerFunction, durationUs) {
  events.push({
    runId: 'node-trace',
    functionName,
    file,
    line,
    durationUs,
    timestampNs: Date.now() * 1000000,
    callerFunction
  });
  if (events.length >= 50) {
    flush();
  }
}

function flush() {
  if (events.length === 0) return;
  const payload = JSON.stringify(events);
  events = [];
  const req = http.request({
    hostname: 'localhost',
    port: 8080,
    path: '/api/runtime/events',
    method: 'POST',
    headers: { 'Content-Type': 'application/json' }
  });
  req.on('error', () => {});
  req.write(payload);
  req.end();
}

module.exports = { recordEvent, flush };
`
	case "go":
		return "// CodeDNA Runtime Tracer Adapter (Go)\n" +
			"package tracer\n\n" +
			"import (\n\t\"bytes\"\n\t\"encoding/json\"\n\t\"net/http\"\n\t\"time\"\n)\n\n" +
			"type Event struct {\n" +
			"\tRunID string `json:\"runId\"`\n" +
			"\tFunctionName string `json:\"functionName\"`\n" +
			"\tFile string `json:\"file\"`\n" +
			"\tLine int `json:\"line\"`\n" +
			"\tDurationUs int64 `json:\"durationUs\"`\n" +
			"\tTimestampNs int64 `json:\"timestampNs\"`\n" +
			"\tCallerFunction string `json:\"callerFunction,omitempty\"`\n" +
			"}\n"

	default:
		return "// CodeDNA Runtime Tracer generic JSON endpoint: POST /api/runtime/events"
	}
}

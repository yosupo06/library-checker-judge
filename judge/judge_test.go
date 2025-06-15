package main

import (
	"embed"
	"flag"
	"os"
	"path"
	"testing"

	"github.com/yosupo06/library-checker-judge/storage"
)

var (
	TESTLIB_PATH    = path.Join("sources", "testlib.h")
	DUMMY_CASE_NAME = "case_00"
)

// NOTE: A+B test constants have been moved to ../langs/integration_test.go
// The following constants are no longer used here:
// - APLUSB_DIR, CHECKER_PATH, PARAMS_H_PATH
// - SAMPLE_IN_PATH, SAMPLE_OUT_PATH, SAMPLE_WA_OUT_PATH

//go:embed sources/*
var sources embed.FS

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

// NOTE: prepareProblemFiles function has been moved to ../langs/integration_test.go
// This function was primarily used for A+B language verification tests which have been migrated.
// If needed for other judge tests, this function can be re-implemented with specific test requirements.

// NOTE: testAplusB and testAplusBAC functions have been moved to ../langs/integration_test.go
// These functions are no longer used in judge module tests since A+B language verification
// has been moved to the langs module where it belongs conceptually.

// NOTE: A+B language verification tests have been moved to ../langs/integration_test.go
// These tests verify that Docker environments for each programming language work correctly.
// The tests were moved to the langs module since they're primarily testing language 
// configuration and Docker image functionality rather than judge-specific logic.
//
// For language verification testing, run: cd ../langs && go test -v ./...
//
// The following functions have been moved:
// - All TestXxxAplusBAC functions (AC tests for each language)
// - Error case tests (TestCppAplusBWA, TestCppAplusBPE, TestCppAplusBTLE, TestCppAplusBRE)
// - Compilation error test (TestAplusBCE)
// - Failure test (TestCppAplusBFail)

package main

import (
	"testing"
)

func TestChecker_RunOnCurrentCodebase(t *testing.T) {
	checker := NewChecker("../../../", false, false)
	checker.checkPackageAndFileNaming()
	checker.checkModuleStructures()
	checker.checkLayerIsolationAndImports()
	checker.checkServiceLayerRules()
	checker.checkHandlerLayerRules()
	checker.checkDTOSecurity()
	checker.checkErrorHandlingAndPanics()
	checker.checkRegisterAndDIContracts()

	errorCount := 0
	for _, v := range checker.violations {
		if v.Severity == SevError {
			errorCount++
			t.Errorf("Unexpected error violation: [%s] %s:%d: %s", v.Rule, v.File, v.Line, v.Message)
		}
	}

	if errorCount > 0 {
		t.Fatalf("Checker reported %d error violations on the codebase", errorCount)
	}

	if checker.passCount == 0 {
		t.Fatalf("Checker did not record any passed checks")
	}
}

func TestIsGeneratedFile(t *testing.T) {
	if !isGeneratedFile("cmd/api/modules_gen.go") {
		t.Errorf("Expected cmd/api/modules_gen.go to be recognized as generated")
	}
	if !isGeneratedFile("internal/modules/graphql/exec.go") {
		t.Errorf("Expected internal/modules/graphql/exec.go to be recognized as generated")
	}
	if isGeneratedFile("internal/modules/user/service.go") {
		t.Errorf("Did not expect user/service.go to be recognized as generated")
	}
}

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LogScriptTestSuite encapsulates shared environment for log script tests
type LogScriptTestSuite struct {
	env ScriptTestEnv // Shared environment
}

// SetupSuite prepares the environment before tests run
func (suite *LogScriptTestSuite) SetupSuite(t *testing.T) {
	// Setup script test environment
	suite.env = SetupScriptTestEnv(t)
}

// Test 1: Valid log levels and filtering behavior
func (suite *LogScriptTestSuite) TestLogLevels(t *testing.T) {
	tests := []struct {
		level       string
		message     string
		expectedMsg string
		envLevel    string
		shouldPrint bool
	}{
		{"DEBUG", "Debug test message", "[DEBUG] Debug test message", "DEBUG", true},
		{"INFO", "Info test message", "[INFO] Info test message", "INFO", true},
		{"WARN", "Warning test message", "[WARN] Warning test message", "INFO", true},
		{"ERROR", "Error test message", "[ERROR] Error test message", "INFO", true},

		// Filtering tests
		{"DEBUG", "Filtered out debug", "[DEBUG] Filtered out debug", "INFO", false},
		{"INFO", "Filtered out info", "[INFO] Filtered out info", "WARN", false},
		{"WARN", "Filtered out warning", "[WARN] Filtered out warning", "ERROR", false},
		{"ERROR", "Should print error", "[ERROR] Should print error", "ERROR", true},

		// Edge cases
		{"INVALID", "Invalid log level", "", "DEBUG", false}, // Invalid log level should not output
		{"INFO", "", "[INFO] ", "DEBUG", true},               // Empty message should still log
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.level, test.envLevel), func(t *testing.T) {
			// Execute the log script inside the container
			cmd := []string{
				"sh", "-c",
				fmt.Sprintf("LOG_LEVEL=%s log %s \"%s\"", test.envLevel, test.level, test.message),
			}
			output, _ := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

			if test.shouldPrint {
				assert.Contains(t, output, test.expectedMsg, "Unexpected log output")
			} else {
				assert.Empty(t, output, "Expected no output but got some")
			}
		})
	}
}

// Test 2: Missing arguments
func (suite *LogScriptTestSuite) TestMissingArguments(t *testing.T) {
	// Run the script without arguments
	cmd := []string{"sh", "-c", "log"}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Expect an error because the script exits with status 1
	require.Error(t, err, "Expected the script to exit with an error due to missing arguments")

	// Verify the output contains usage instructions
	assert.Contains(t, output, "Usage: log", "Expected usage instructions for missing arguments")
}

// Test 3: Invalid log level
func (suite *LogScriptTestSuite) TestInvalidLogLevel(t *testing.T) {
	cmd := []string{"sh", "-c", "log INVALID 'Invalid level test'"}
	output, _ := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Invalid levels should produce no output
	assert.Empty(t, output, "Invalid log level should not produce output")
}

// Main test runner for the suite
func TestLogScriptTestSuite(t *testing.T) {
	suite := &LogScriptTestSuite{}
	suite.SetupSuite(t)

	t.Run("TestLogLevels", suite.TestLogLevels)
	t.Run("TestMissingArguments", suite.TestMissingArguments)
	t.Run("TestInvalidLogLevel", suite.TestInvalidLogLevel)
}
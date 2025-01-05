package tests

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ParseYamlTestSuite encapsulates shared environment for parse_yaml tests
type ParseYamlTestSuite struct {
	env ScriptTestEnv // Shared environment
}

// SetupSuite prepares the environment before tests run
func (suite *ParseYamlTestSuite) SetupSuite(t *testing.T) {
	// Setup script test environment
	suite.env = SetupScriptTestEnv(t)
}

// Helper function to create a config file inside the container
func (suite *ParseYamlTestSuite) createConfigFile(t *testing.T, content string) string {
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	filePath := fmt.Sprintf("/tmp/config-%s.yaml", safeName)

	cmd := []string{
		"sh", "-c", fmt.Sprintf("echo '%s' | tee %s", content, filePath),
	}
	_, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to create config file")

	// Ensure file permissions are correct
	cmd = []string{"sh", "-c", fmt.Sprintf("chmod 644 %s", filePath)}
	_, err = ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to set permissions")

	return filePath
}
// Test 1: Valid YAML config
func (suite *ParseYamlTestSuite) TestValidConfig(t *testing.T) {
	config := `
server:
  address: "127.0.0.1"
  min_port: 21000
  max_port: 21010
  tls_cert: "/etc/ftp/cert.pem"
  tls_key: "/etc/ftp/key.pem"
users:
  - username: "user1"
    password_env: "USER1_PASS"
  - username: "user2"
    password_env: "USER2_PASS"
`

	configPath := suite.createConfigFile(t, config)

	// Execute the parse_yaml script
	cmd := []string{"sh", "-c", fmt.Sprintf("parse_yaml %s", configPath)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to execute parse_yaml")

	// Validate the output
	assert.Contains(t, output, "CONFIG_FILE_DETECTED=1")
	assert.Contains(t, output, "YAML_ADDRESS='127.0.0.1'")
	assert.Contains(t, output, "YAML_MIN_PORT='21000'")
	assert.Contains(t, output, "YAML_MAX_PORT='21010'")
	assert.Contains(t, output, "YAML_TLS_CERT='/etc/ftp/cert.pem'")
	assert.Contains(t, output, "YAML_TLS_KEY='/etc/ftp/key.pem'")
	assert.Contains(t, output, "YAML_USER_COUNT=2")
	assert.Contains(t, output, "YAML_USER_0_NAME='user1'")
	assert.Contains(t, output, "YAML_USER_0_PASS_ENV='USER1_PASS'")
	assert.Contains(t, output, "YAML_USER_1_NAME='user2'")
	assert.Contains(t, output, "YAML_USER_1_PASS_ENV='USER2_PASS'")
}

// Test 2: Empty YAML config
func (suite *ParseYamlTestSuite) TestEmptyConfig(t *testing.T) {
	config := ""
	configPath := suite.createConfigFile(t, config)

	cmd := []string{"sh", "-c", fmt.Sprintf("parse_yaml %s", configPath)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to execute parse_yaml")

	assert.Contains(t, output, "CONFIG_FILE_DETECTED=1")
	assert.Contains(t, output, "YAML_ADDRESS=''")
	assert.Contains(t, output, "YAML_MIN_PORT=''")
	assert.Contains(t, output, "YAML_MAX_PORT=''")
	assert.Contains(t, output, "YAML_TLS_CERT=''")
	assert.Contains(t, output, "YAML_TLS_KEY=''")
	assert.Contains(t, output, "YAML_USER_COUNT=0")
}

// Test 3: Missing Config File
func (suite *ParseYamlTestSuite) TestMissingConfig(t *testing.T) {
	cmd := []string{"sh", "-c", "parse_yaml /tmp/missing.yaml"}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to execute parse_yaml")

	assert.Contains(t, output, "CONFIG_FILE_DETECTED=0")
	assert.Contains(t, output, "YAML_ADDRESS=''")
	assert.Contains(t, output, "YAML_MIN_PORT=''")
	assert.Contains(t, output, "YAML_MAX_PORT=''")
	assert.Contains(t, output, "YAML_TLS_CERT=''")
	assert.Contains(t, output, "YAML_TLS_KEY=''")
	assert.Contains(t, output, "YAML_USER_COUNT=0")
}

// Test 4: Invalid YAML Format
func (suite *ParseYamlTestSuite) TestInvalidYaml(t *testing.T) {
	config := `
server:
  address: "127.0.0.1
  min_port: 21000
  max_port: 21010
users:
  - username: "user1"
    password_env: "USER1_PASS
`
	configPath := suite.createConfigFile(t, config)

	cmd := []string{"sh", "-c", fmt.Sprintf("parse_yaml %s", configPath)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to execute parse_yaml")

	// Invalid YAML should be treated as empty
	assert.Contains(t, output, "CONFIG_FILE_DETECTED=1")
	assert.Contains(t, output, "YAML_ADDRESS=''")
	assert.Contains(t, output, "YAML_MIN_PORT=''")
	assert.Contains(t, output, "YAML_MAX_PORT=''")
	assert.Contains(t, output, "YAML_TLS_CERT=''")
	assert.Contains(t, output, "YAML_TLS_KEY=''")
	assert.Contains(t, output, "YAML_USER_COUNT=0")
}

// Main test runner
func TestParseYamlTestSuite(t *testing.T) {
	suite := &ParseYamlTestSuite{}
	suite.SetupSuite(t)

	t.Run("TestValidConfig", suite.TestValidConfig)
	t.Run("TestEmptyConfig", suite.TestEmptyConfig)
	t.Run("TestMissingConfig", suite.TestMissingConfig)
	t.Run("TestInvalidYaml", suite.TestInvalidYaml)
}
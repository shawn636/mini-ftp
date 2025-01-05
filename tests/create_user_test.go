package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// CreateUserTestSuite encapsulates shared environment for create_user tests
type CreateUserTestSuite struct {
	env ScriptTestEnv // Shared environment
}

// SetupSuite prepares the environment before tests run
func (suite *CreateUserTestSuite) SetupSuite(t *testing.T) {
	// Setup script test environment
	suite.env = SetupScriptTestEnv(t)
}

// Test 1: Create valid user
func (suite *CreateUserTestSuite) TestCreateValidUser(t *testing.T) {
	username := "testuser"
	password := "securepass"

	// Execute the create_user script
	cmd := []string{"sh", "-c", fmt.Sprintf("create_user %s %s", username, password)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to execute create_user")

	// Clean output for assertions
	output = stripAnsiCodes(output)

	// Validate the output logs
	assert.Contains(t, output, "Adding user:", "Expected log indicating user creation")
	assert.Contains(t, output, "User testuser created successfully.", "Expected success confirmation")

	// Verify the user exists in the system
	cmd = []string{"sh", "-c", fmt.Sprintf("id -u %s", username)}
	output, err = ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "User should exist in the system")
	assert.NotEmpty(t, output, "Expected valid UID for the user")

	// Verify the user's home directory
	cmd = []string{"sh", "-c", fmt.Sprintf("stat -c '%%U:%%G' /ftp/%s", username)}
	output, err = ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to check user directory ownership")
	assert.Contains(t, output, fmt.Sprintf("%s:%s", username, username), "Directory ownership should match the user")
}

// Test 2: Missing username
func (suite *CreateUserTestSuite) TestMissingUsername(t *testing.T) {
	cmd := []string{"sh", "-c", "create_user '' 'password'"}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Expect an error
	require.Error(t, err, "Expected error due to missing username")
	assert.Contains(t, stripAnsiCodes(output), "Usage: create_user <username> <password>", "Expected usage error")
}

// Test 3: Missing password
func (suite *CreateUserTestSuite) TestMissingPassword(t *testing.T) {
	cmd := []string{"sh", "-c", "create_user 'user' ''"}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Expect an error
	require.Error(t, err, "Expected error due to missing password")
	assert.Contains(t, stripAnsiCodes(output), "Usage: create_user <username> <password>", "Expected usage error")
}

// Test 4: User already exists
func (suite *CreateUserTestSuite) TestUserAlreadyExists(t *testing.T) {
	username := "existinguser"
	password := "password"

	// Create the user first
	cmd := []string{"sh", "-c", fmt.Sprintf("create_user %s %s", username, password)}
	_, create_user_err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, create_user_err, "Failed to create initial user")

	// Attempt to create the user again
	cmd = []string{"sh", "-c", fmt.Sprintf("create_user %s %s", username, password)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Expect an error since the user already exists
	require.Error(t, err, "Expected error due to duplicate user")
	assert.Contains(t, stripAnsiCodes(output), fmt.Sprintf("User '%s' already exists", username), "Expected duplicate user error")
}

// Test 5: Invalid username
func (suite *CreateUserTestSuite) TestInvalidUsername(t *testing.T) {
	invalidUsername := "invalid:user"
	password := "password"

	cmd := []string{"sh", "-c", fmt.Sprintf("create_user '%s' '%s'", invalidUsername, password)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)

	// Expect an error due to invalid username
	require.Error(t, err, "Expected error due to invalid username")
	assert.Contains(t, stripAnsiCodes(output), "Invalid username:", "Expected invalid username error")
}

// Test 6: Directory Permissions
func (suite *CreateUserTestSuite) TestDirectoryPermissions(t *testing.T) {
	username := "diruser"
	password := "password"

	// Create the user
	cmd := []string{"sh", "-c", fmt.Sprintf("create_user %s %s", username, password)}
	_, create_user_err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, create_user_err, "Failed to create user")

	// Check directory permissions
	cmd = []string{"sh", "-c", fmt.Sprintf("stat -c '%%a' /ftp/%s", username)}
	output, err := ExecCommandInContainer(t, suite.env.ContainerName, cmd)
	require.NoError(t, err, "Failed to check directory permissions")
	assert.Contains(t, output, "755", "Expected directory permissions to be 755")
}

// Main test runner
func TestCreateUserTestSuite(t *testing.T) {
	suite := &CreateUserTestSuite{}
	suite.SetupSuite(t)

	t.Run("TestCreateValidUser", suite.TestCreateValidUser)
	t.Run("TestMissingUsername", suite.TestMissingUsername)
	t.Run("TestMissingPassword", suite.TestMissingPassword)
	t.Run("TestUserAlreadyExists", suite.TestUserAlreadyExists)
	t.Run("TestInvalidUsername", suite.TestInvalidUsername)
	t.Run("TestDirectoryPermissions", suite.TestDirectoryPermissions)
}
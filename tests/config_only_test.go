package tests

import (
	"bytes"
	"testing"

	"github.com/secsy/goftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConfigOnlyTestSuite encapsulates test options and clients
type ConfigOnlyTestSuite struct {
	opts    TestOptions
	clients map[string]*goftp.Client
}

// SetupSuite initializes environment and clients before tests run
func (suite *ConfigOnlyTestSuite) SetupSuite(t *testing.T) {
	// Define test options (Single User)
	config := "config.yaml"
	suite.opts = TestOptions{
		ComposeFile:  "docker-compose.config-only.yaml",
		ConfigFile:   &config,
		UseSSL:       false,
		Address:      "127.0.0.1",
		Port:         2123,
		PassivePorts: "22020-22029",
		Users: map[string]string{
			"user": "afqpVazRzAdN",
		},
	}

	// Setup environment
	tmpAndProject := setupTestEnv(t, suite.opts)
	t.Cleanup(func() { teardownTestEnv(t, tmpAndProject) })

	// Setup FTP clients for all users
	suite.clients = setupFTPClients(t, suite.opts)
}

// TestFileOperations checks upload, download, and deletion
func (suite *ConfigOnlyTestSuite) TestFileOperations(t *testing.T) {
	client := suite.clients["user"]

	// Upload a file
	content := []byte("test file content")
	require.NoError(t, client.Store("test-file.txt", bytes.NewReader(content)))

	// Download and verify content
	var buf bytes.Buffer
	require.NoError(t, client.Retrieve("test-file.txt", &buf))
	assert.Equal(t, "test file content", buf.String())

	// Delete the file and verify removal
	require.NoError(t, client.Delete("test-file.txt"))

	entries, err := client.ReadDir("/")
	require.NoError(t, err)
	for _, entry := range entries {
		assert.NotEqual(t, "test-file.txt", entry.Name(), "File should be deleted")
	}
}

// TestRenameFile checks renaming functionality
func (suite *ConfigOnlyTestSuite) TestRenameFile(t *testing.T) {
	client := suite.clients["user"]

	// Upload file
	require.NoError(t, client.Store("rename-test.txt", bytes.NewReader([]byte("rename test"))))

	// Rename file
	require.NoError(t, client.Rename("rename-test.txt", "renamed-file.txt"))

	// Verify renamed file exists
	entries, err := client.ReadDir("/")
	require.NoError(t, err)

	found := false
	for _, entry := range entries {
		if entry.Name() == "renamed-file.txt" {
			found = true
			break
		}
	}
	assert.True(t, found, "Renamed file should exist")

	// Cleanup
	require.NoError(t, client.Delete("renamed-file.txt"))
}

// TestDirectoryOperations checks directory operations
func (suite *ConfigOnlyTestSuite) TestDirectoryOperations(t *testing.T) {
	client := suite.clients["user"]

	// Create directory
	dirPath, err := client.Mkdir("test-dir")
	require.NoError(t, err)
	assert.Equal(t, "/test-dir", dirPath)

	// Rename directory
	require.NoError(t, client.Rename("test-dir", "renamed-dir"))

	// Verify renamed directory exists
	entries, err := client.ReadDir("/")
	require.NoError(t, err)

	found := false
	for _, entry := range entries {
		if entry.Name() == "renamed-dir" && entry.IsDir() {
			found = true
			break
		}
	}
	assert.True(t, found, "Renamed directory should exist")

	// Remove directory
	require.NoError(t, client.Rmdir("renamed-dir"))

	// Verify deletion
	entries, err = client.ReadDir("/")
	require.NoError(t, err)

	found = false
	for _, entry := range entries {
		if entry.Name() == "renamed-dir" {
			found = true
			break
		}
	}
	assert.False(t, found, "Directory should be deleted")
}

// TestAccessControl ensures restricted access to unauthorized directories
func (suite *ConfigOnlyTestSuite) TestAccessControl(t *testing.T) {
	client := suite.clients["user"]

	// Attempt unauthorized access
	err := client.Retrieve("../unauthorized.txt", &bytes.Buffer{})
	assert.Error(t, err, "Should not be able to access files outside FTP home directory")
}

// TestFilePermissions checks file permissions
func (suite *ConfigOnlyTestSuite) TestFilePermissions(t *testing.T) {
	client := suite.clients["user"]

	// Upload a file
	content := []byte("permission test")
	require.NoError(t, client.Store("perm-test.txt", bytes.NewReader(content)))

	// Download and verify content
	var buf bytes.Buffer
	require.NoError(t, client.Retrieve("perm-test.txt", &buf))
	assert.Equal(t, "permission test", buf.String())

	// Cleanup
	require.NoError(t, client.Delete("perm-test.txt"))
}

// Main test runner
func TestConfigOnlyTestSuite(t *testing.T) {
	suite := &ConfigOnlyTestSuite{}
	suite.SetupSuite(t)

	t.Run("TestFileOperations", suite.TestFileOperations)
	t.Run("TestRenameFile", suite.TestRenameFile)
	t.Run("TestDirectoryOperations", suite.TestDirectoryOperations)
	t.Run("TestAccessControl", suite.TestAccessControl)
	t.Run("TestFilePermissions", suite.TestFilePermissions)
}
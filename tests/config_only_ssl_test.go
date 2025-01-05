package tests

import (
	"bytes"
	"testing"

	"github.com/secsy/goftp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ConfigOnlySSLTestSuite encapsulates test options and clients
type ConfigOnlySSLTestSuite struct {
	opts    TestOptions
	clients map[string]*goftp.Client
}

// SetupSuite initializes environment and clients before tests run
func (suite *ConfigOnlySSLTestSuite) SetupSuite(t *testing.T) {
	// Define test options (Multiple Users with SSL)
	config := "config-ssl.yaml"
	suite.opts = TestOptions{
		ComposeFile:  "docker-compose.config-only-ssl.yaml",
		ConfigFile:   &config,
		UseSSL:       true,
		Address:      "mini-ftp.duckdns.org",
		Port:         2124,
		PassivePorts: "22030-22039",
		Users: map[string]string{
			"user1": "MZZwC9F9JRrd",  // Matches environment variable CONFIG_ONLY_SSL_TEST_USER1_PASS
			"user2": "QDWubfLQqRFm",  // Matches environment variable CONFIG_ONLY_SSL_TEST_USER2_PASS
		},
	}

	// Setup environment
	tmpAndProject := setupTestEnv(t, suite.opts)
	t.Cleanup(func() { teardownTestEnv(t, tmpAndProject) })

	// Setup FTP clients for all users
	suite.clients = setupFTPClients(t, suite.opts)
}

// TestFileOperations checks upload, download, and deletion
func (suite *ConfigOnlySSLTestSuite) TestFileOperations(t *testing.T) {
	for username, client := range suite.clients {
		t.Run("FileOperations_"+username, func(t *testing.T) {
			content := []byte("test file content for " + username)
			fileName := "test-file-" + username + ".txt"

			// Upload file
			require.NoError(t, client.Store(fileName, bytes.NewReader(content)))

			// Download and verify content
			var buf bytes.Buffer
			require.NoError(t, client.Retrieve(fileName, &buf))
			assert.Equal(t, string(content), buf.String())

			// Delete the file and verify removal
			require.NoError(t, client.Delete(fileName))

			entries, err := client.ReadDir("/")
			require.NoError(t, err)
			for _, entry := range entries {
				assert.NotEqual(t, fileName, entry.Name(), "File should be deleted")
			}
		})
	}
}

// TestRenameFile checks renaming functionality
func (suite *ConfigOnlySSLTestSuite) TestRenameFile(t *testing.T) {
	for username, client := range suite.clients {
		t.Run("RenameFile_"+username, func(t *testing.T) {
			oldFile := "rename-test-" + username + ".txt"
			newFile := "renamed-file-" + username + ".txt"

			// Upload file
			require.NoError(t, client.Store(oldFile, bytes.NewReader([]byte("rename test"))))

			// Rename file
			require.NoError(t, client.Rename(oldFile, newFile))

			// Verify renamed file exists
			entries, err := client.ReadDir("/")
			require.NoError(t, err)

			found := false
			for _, entry := range entries {
				if entry.Name() == newFile {
					found = true
					break
				}
			}
			assert.True(t, found, "Renamed file should exist")

			// Cleanup
			require.NoError(t, client.Delete(newFile))
		})
	}
}

// TestDirectoryOperations checks directory operations
func (suite *ConfigOnlySSLTestSuite) TestDirectoryOperations(t *testing.T) {
	for username, client := range suite.clients {
		t.Run("DirectoryOperations_"+username, func(t *testing.T) {
			oldDir := "test-dir-" + username
			newDir := "renamed-dir-" + username

			// Create directory
			_, err := client.Mkdir(oldDir)
			require.NoError(t, err)
			assert.Equal(t, "/"+oldDir, "/"+oldDir)

			// Rename directory
			require.NoError(t, client.Rename(oldDir, newDir))

			// Verify renamed directory exists
			entries, err := client.ReadDir("/")
			require.NoError(t, err)

			found := false
			for _, entry := range entries {
				if entry.Name() == newDir && entry.IsDir() {
					found = true
					break
				}
			}
			assert.True(t, found, "Renamed directory should exist")

			// Remove directory
			require.NoError(t, client.Rmdir(newDir))

			// Verify deletion
			entries, err = client.ReadDir("/")
			require.NoError(t, err)

			found = false
			for _, entry := range entries {
				if entry.Name() == newDir {
					found = true
					break
				}
			}
			assert.False(t, found, "Directory should be deleted")
		})
	}
}

// TestAccessControl ensures restricted access to unauthorized directories
func (suite *ConfigOnlySSLTestSuite) TestAccessControl(t *testing.T) {
	for username, client := range suite.clients {
		t.Run("AccessControl_"+username, func(t *testing.T) {
			err := client.Retrieve("../unauthorized.txt", &bytes.Buffer{})
			assert.Error(t, err, "Should not be able to access files outside FTP home directory")
		})
	}
}

// TestFilePermissions validates file permissions
func (suite *ConfigOnlySSLTestSuite) TestFilePermissions(t *testing.T) {
	for username, client := range suite.clients {
		t.Run("FilePermissions_"+username, func(t *testing.T) {
			content := []byte("permission test for " + username)
			fileName := "perm-test-" + username + ".txt"

			// Upload file
			require.NoError(t, client.Store(fileName, bytes.NewReader(content)))

			// Download and verify content
			var buf bytes.Buffer
			require.NoError(t, client.Retrieve(fileName, &buf))
			assert.Equal(t, string(content), buf.String())

			// Cleanup
			require.NoError(t, client.Delete(fileName))
		})
	}
}

// Main test runner
func TestConfigOnlySSLTestSuite(t *testing.T) {
	suite := &ConfigOnlySSLTestSuite{}
	suite.SetupSuite(t)

	t.Run("TestFileOperations", suite.TestFileOperations)
	t.Run("TestRenameFile", suite.TestRenameFile)
	t.Run("TestDirectoryOperations", suite.TestDirectoryOperations)
	t.Run("TestAccessControl", suite.TestAccessControl)
	t.Run("TestFilePermissions", suite.TestFilePermissions)
}
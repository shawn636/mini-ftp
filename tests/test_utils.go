package tests

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/secsy/goftp"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestOptions defines requirements for environment setup
type TestOptions struct {
	ComposeFile  string            // Source compose file (e.g., docker-compose.env-only.yaml)
	ConfigFile   *string           // Optional config file
	UseSSL       bool
	Address      string            // Domain or IP address to connect to
	Port         int               // Custom FTP port for this test
	PassivePorts string            // Passive port range
	Users        map[string]string // Multiple username-password pairs
}

// ScriptTestEnv defines the environment for script tests
type ScriptTestEnv struct {
	TempDir       string
	ContainerName string
	ImageName     string
}

// Setup test environment based on test options
func setupTestEnv(t *testing.T, opts TestOptions) string {
	projectRoot, err := filepath.Abs(filepath.Join(".."))
	require.NoError(t, err, "Failed to determine project root")

	alpineVersion := os.Getenv("ALPINE_VERSION")
	if alpineVersion == "" {
		alpineVersion = "latest"
	}

	tmpDir, err := os.MkdirTemp("", "test-env-*")
	require.NoError(t, err, "Failed to create temp directory")

	projectName := "test-" + strings.Replace(filepath.Base(tmpDir), "test-env-", "", 1)

	// Copy essential files
	copyFiles(t, projectRoot, tmpDir, []string{"Dockerfile"})
	err = copyDir(filepath.Join(projectRoot, "scripts"), filepath.Join(tmpDir, "scripts"))
	require.NoError(t, err, "Failed to copy scripts directory")

	configDir := filepath.Join(tmpDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755), "Failed to create config directory")
	copyFiles(t, filepath.Join(projectRoot, "tests/fixtures"), configDir, []string{"vsftpd.conf"})

	// SSL-specific configuration
	if opts.UseSSL || opts.ConfigFile != nil {
		copyFiles(t, projectRoot, tmpDir, []string{".env"})
		copyFiles(t, filepath.Join(projectRoot, "tests/fixtures"), tmpDir, []string{"Dockerfile.ssl"})
		copyFiles(t, filepath.Join(projectRoot, "tests/fixtures"), tmpDir, []string{"generate-cert.sh"})
	}

	srcCompose := filepath.Join(projectRoot, "tests/fixtures", opts.ComposeFile)
	destCompose := filepath.Join(tmpDir, "docker-compose.yaml")
	data, err := os.ReadFile(srcCompose)
	require.NoError(t, err, "Failed to read compose file")
	require.NoError(t, os.WriteFile(destCompose, data, 0644), "Failed to write compose file")

	if opts.ConfigFile != nil {
		copyFiles(t, filepath.Join(projectRoot, "tests/fixtures"), tmpDir, []string{*opts.ConfigFile})
	}

	// Dynamically set environment variables for all users
	envVars := []string{
		fmt.Sprintf("ALPINE_VERSION=%s", alpineVersion),
	}
	for username, password := range opts.Users {
		envKey := fmt.Sprintf("%s_PASS", strings.ToUpper(username))
		envVars = append(envVars, fmt.Sprintf("%s=%s", envKey, password))
	}

	cmd := exec.Command(
		"docker",
		"compose",
		"--project-name", projectName,
		"up", "--build", "-d", "--wait",
	)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), envVars...)

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Error during 'docker compose up':\n%s", string(output))
		t.Fatalf("Failed to start Docker Compose: %v", err)
	}

	// require.NoError(t, cmd.Run(), "Failed to start Docker Compose")

	containerName := projectName + "-ftp-1"
	verifyAlpineVersion(t, containerName, alpineVersion)

	return fmt.Sprintf("%s:%s", tmpDir, projectName)
}

// Teardown test environment
func teardownTestEnv(t *testing.T, tmpDirAndProject string) {
	parts := strings.Split(tmpDirAndProject, ":")
	tmpDir := parts[0]
	projectName := parts[1]

	// Check if logs should be printed
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		// Fetch logs for all containers in the project
		cmd := exec.Command(
			"docker",
			"compose",
			"--project-name", projectName,
			"logs",
		)
		cmd.Dir = tmpDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // Print logs but continue even if this command fails
	}

	// Tear down Docker Compose stack completely
	cmd := exec.Command(
		"docker",
		"compose",
		"--project-name", projectName,
		"down", "--volumes", "--rmi", "all",
	)
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run(), "Failed to tear down Docker Compose")

	// Remove temp directory
	require.NoError(t, os.RemoveAll(tmpDir), "Failed to remove temp directory")
}

// SetupScriptTestEnv creates a temporary test environment and starts a container for script tests
func SetupScriptTestEnv(t *testing.T) ScriptTestEnv {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "script-test-env-*")
	require.NoError(t, err, "Failed to create temporary directory")

	// Get project root dynamically
	projectRoot, err := filepath.Abs(filepath.Join(".."))
	require.NoError(t, err, "Failed to determine project root")

	// Copy the Dockerfile into the temp directory
	copyFiles(t, filepath.Join(projectRoot, "tests/fixtures"), tmpDir, []string{"Dockerfile"})

	// Copy the scripts directory into the temp directory
	require.NoError(t, copyDir(filepath.Join(projectRoot, "scripts"), filepath.Join(tmpDir, "scripts")),
		"Failed to copy scripts directory")

	// Generate unique names for the container and image
	imageName := "script-test-image-" + filepath.Base(tmpDir)
	containerName := "script-test-container-" + filepath.Base(tmpDir)

	// Build the Docker image with ALPINE_VERSION as a build argument (suppress output)
	alpineVersion := os.Getenv("ALPINE_VERSION")
	cmd := exec.Command("docker", "build", "--build-arg", fmt.Sprintf("ALPINE_VERSION=%s", alpineVersion), "-t", imageName, ".")
	cmd.Dir = tmpDir
	cmd.Stdout = nil // Suppress standard output
	cmd.Stderr = nil // Suppress standard error
	require.NoError(t, cmd.Run(), "Failed to build Docker image")

	// Start the container (suppress output)
	cmd = exec.Command("docker", "run", "--rm", "-d", "--name", containerName, imageName, "sleep", "infinity")
	cmd.Stdout = nil // Suppress standard output
	cmd.Stderr = nil // Suppress standard error
	require.NoError(t, cmd.Run(), "Failed to start Docker container")

	if alpineVersion == "" {
		alpineVersion = "latest"
	}

	verifyAlpineVersion(t, containerName, alpineVersion)

	// Cleanup after tests
	t.Cleanup(func() {
		TeardownScriptTestEnv(t, ScriptTestEnv{
			TempDir:       tmpDir,
			ContainerName: containerName,
			ImageName:     imageName,
		})
	})

	// Return environment info for tests
	return ScriptTestEnv{
		TempDir:       tmpDir,
		ContainerName: containerName,
		ImageName:     imageName,
	}
}

// TeardownScriptTestEnv shuts down the container and removes related resources
func TeardownScriptTestEnv(t *testing.T, env ScriptTestEnv) {
	// Check if logs should be printed
	if os.Getenv("LOG_LEVEL") == "DEBUG" {
		// Fetch logs for the container
		cmd := exec.Command("docker", "logs", env.ContainerName)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		_ = cmd.Run() // Print logs but continue even if this command fails
	}

	// Stop and remove the container
	cmd := exec.Command("docker", "rm", "-f", env.ContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), "Failed to remove container")

	// Remove the Docker image
	cmd = exec.Command("docker", "rmi", "-f", env.ImageName)
	cmd.Stdout = nil // Suppress standard output
	cmd.Stderr = nil // Suppress standard error
	require.NoError(t, cmd.Run(), "Failed to remove Docker image")

	// Remove the temporary directory
	require.NoError(t, os.RemoveAll(env.TempDir), "Failed to remove temporary directory")
}

// ExecCommandInContainer runs a shell command inside the specified container
func ExecCommandInContainer(t *testing.T, containerName string, command []string) (string, error) {
	cmd := exec.Command("docker", append([]string{"exec", containerName}, command...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// Copy files
func copyFiles(t *testing.T, srcDir, destDir string, files []string) {
	for _, file := range files {
		srcPath := filepath.Join(srcDir, file)
		destPath := filepath.Join(destDir, file)

		data, err := os.ReadFile(srcPath)
		require.NoError(t, err, "Failed to read source file: "+srcPath)
		require.NoError(t, os.WriteFile(destPath, data, 0644), "Failed to write to destination: "+destPath)
	}
}

// Copy entire directory
func copyDir(src string, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}

// setupFTPClients initializes FTP clients for all configured users
func setupFTPClients(t *testing.T, opts TestOptions) map[string]*goftp.Client {
	clients := make(map[string]*goftp.Client)

	for username, password := range opts.Users {
		config := goftp.Config{
			User:     username,
			Password: password,
			Timeout:  10 * time.Second,
		}

		if opts.UseSSL {
			config.TLSConfig = &tls.Config{
				ServerName:         opts.Address,
				InsecureSkipVerify: true,
			}
			config.TLSMode = goftp.TLSExplicit
		}

		address := fmt.Sprintf("%s:%d", opts.Address, opts.Port)
		client, err := goftp.DialConfig(config, address)
		require.NoError(t, err, fmt.Sprintf("Failed to connect to FTP server as user: %s", username))

		clients[username] = client
	}

	return clients
}

// stripAnsiCodes removes ANSI escape codes from a string
func stripAnsiCodes(input string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return re.ReplaceAllString(input, "")
}

func verifyAlpineVersion(t *testing.T, containerName, expectedVersion string) {
    // Execute command to retrieve Alpine version inside the container
    output, err := ExecCommandInContainer(t, containerName, []string{"cat", "/etc/alpine-release"})
    require.NoError(t, err, "Failed to retrieve Alpine version from container")

    // Trim any surrounding whitespace or newline characters
    actualVersion := strings.TrimSpace(output)

    // If expectedVersion is "latest", fetch the latest version online
    if expectedVersion == "latest" {
        expectedVersion = fetchLatestAlpineVersion(t)
    }

    // Compare the actual version with the expected version
	require.True(t, strings.HasPrefix(actualVersion, expectedVersion), "Alpine version mismatch: expected prefix %s, got %s", expectedVersion, actualVersion)
}

type AlpineRelease struct {
	Version string `yaml:"version"`
}
func fetchLatestAlpineVersion(t *testing.T) string {
	// URL for the Alpine release metadata
	url := "https://dl-cdn.alpinelinux.org/alpine/latest-stable/releases/x86_64/latest-releases.yaml"

	// Send an HTTP GET request
	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to fetch Alpine releases metadata")
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read Alpine releases metadata")

	// Parse the YAML response
	var releases []AlpineRelease
	err = yaml.Unmarshal(body, &releases)
	require.NoError(t, err, "Failed to parse Alpine releases metadata")

	// Check that at least one release was found
	require.NotEmpty(t, releases, "No releases found in metadata")

	// Extract and return only the major and minor version (e.g., "3.14" from "3.14.2")
	versionParts := strings.Split(releases[0].Version, ".")
	require.True(t, len(versionParts) >= 2, "Invalid version format in metadata")
	return fmt.Sprintf("%s.%s", versionParts[0], versionParts[1])
}
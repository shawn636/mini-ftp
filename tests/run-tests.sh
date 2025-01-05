#!/usr/bin/env bash

# Configurable Alpine versions
VERSIONS=("3.19" "3.20" "3.21" "latest")

# Exit immediately on error
set -e

# Color codes (safe for printf)
RED=$(printf '\033[0;31m')
GREEN=$(printf '\033[0;32m')
YELLOW=$(printf '\033[1;33m')
BLUE=$(printf '\033[1;34m')
CYAN=$(printf '\033[1;36m')
WHITE=$(printf '\033[1;37m')
NC=$(printf '\033[0m') # No color

# Emojis
CHECKMARK="✅"
CROSSMARK="❌"
SPARKLES="🌟"
ROCKET="🚀"
COMPUTER="💻"
TEST_TUBE="🧪"
REFRESH="🔄"
SIREN="🚨"
PARTY="🎉"

# Usage info
usage() {
  printf "%sUsage:%s %s [OPTIONS]\n\n" "$CYAN" "$NC" "$0"
  printf "%sOptions:%s\n" "$CYAN" "$NC"
  printf "  %s--help, -h%s                Show this help message\n" "$GREEN" "$NC"
  printf "  %s--alpine-latest%s          Test only the latest Alpine version\n" "$GREEN" "$NC"
  printf "  %s--alpine-version, -v VER%s Test specific Alpine versions (can specify multiple)\n\n" "$GREEN" "$NC"
  printf "%sExamples:%s\n" "$CYAN" "$NC"
  printf "  %s --alpine-latest\n" "$0"
  printf "  %s --alpine-version 3.19 -v 3.21\n" "$0"
  printf "  %s\n\n" "$0"
  exit 0
}

# Parse arguments
CUSTOM_VERSIONS=()
LATEST_ONLY=false

while [[ "$#" -gt 0 ]]; do
  case $1 in
    --help|-h) usage ;;                                # Support for --help and -h
    --alpine-latest) LATEST_ONLY=true ;;               # Test only latest
    --alpine-version|-v) CUSTOM_VERSIONS+=("$2"); shift ;; # Support multiple -v flags
    *) printf "%s %sUnknown option: %s%s\n" "$SIREN" "$RED" "$1" "$NC"; usage ;; # Invalid options
  esac
  shift
done

# Determine versions to test
if [ "$LATEST_ONLY" = true ]; then
  VERSIONS=("latest")
elif [ ${#CUSTOM_VERSIONS[@]} -gt 0 ]; then
  VERSIONS=("${CUSTOM_VERSIONS[@]}")
fi

# Save the original working directory
ORIGINAL_DIR=$(pwd)

# Define a trap to always restore the working directory
cleanup() {
  cd "$ORIGINAL_DIR"
  printf "%s %sCleaning up and restoring working directory.%s\n" "$REFRESH" "$CYAN" "$NC"
}
trap cleanup EXIT

# Move to the 'tests' directory where the go.mod file is located
cd tests

# Start the test suite
printf "%s %sStarting test suite...%s\n" "$SPARKLES" "$CYAN" "$NC"

SUCCESS=true

for VERSION in "${VERSIONS[@]}"; do
  printf "%s======================================%s\n" "$YELLOW" "$NC"
  printf "%s %sTesting with Alpine version: %s%s%s\n" "$COMPUTER" "$WHITE" "$BLUE" "$VERSION" "$NC"
  printf "%s======================================%s\n" "$YELLOW" "$NC"

  # Set the Alpine version for the test
  export ALPINE_VERSION=$VERSION

  # Run Go tests and catch any failure
  printf "%s %sRunning tests...%s\n" "$TEST_TUBE" "$CYAN" "$NC"
  if ! go test -p 4 -timeout 5m -v ./...; then
    SUCCESS=false
    printf "%s %sTests failed for Alpine version: %s%s\n" "$CROSSMARK" "$RED" "$VERSION" "$NC"
    break
  fi

  printf "%s %sTests completed successfully for Alpine version: %s%s\n" "$CHECKMARK" "$GREEN" "$VERSION" "$NC"
done

# Print final result
if [ "$SUCCESS" = true ]; then
  printf "%s======================================%s\n" "$GREEN" "$NC"
  printf "%s %sAll tests passed successfully!%s\n" "$PARTY" "$GREEN" "$NC"
else
  printf "%s======================================%s\n" "$RED" "$NC"
  printf "%s %sSome tests failed!%s\n" "$SIREN" "$RED" "$NC"
  exit 1
fi
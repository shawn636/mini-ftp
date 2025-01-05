#!/usr/bin/env bash
# log - Simple logging utility

# --- Colors for Logs ---
COLOR_DEBUG="\033[36m"  # Cyan
COLOR_INFO="\033[0m"    # Default
COLOR_WARN="\033[33m"   # Yellow
COLOR_ERROR="\033[31m"  # Red
COLOR_RESET="\033[0m"   # Reset

# --- Validate Arguments ---
if [ $# -lt 2 ]; then
  echo "Usage: log <LEVEL> <MESSAGE>" >&2
  exit 1
fi

# --- Logging Function ---
LEVEL="$1"
MESSAGE="$2"
LOG_LEVEL="${LOG_LEVEL:-INFO}" # Default log level

# --- Validate Log Level ---
VALID_LEVELS=("DEBUG" "INFO" "WARN" "ERROR")
if [[ ! " ${VALID_LEVELS[*]} " =~ " $LEVEL " ]]; then
  exit 0 # Silently exit if log level is invalid
fi

# Determine color based on log level
COLOR="$COLOR_RESET"
case "$LEVEL" in
  DEBUG) COLOR="$COLOR_DEBUG" ;;
  INFO)  COLOR="$COLOR_INFO"  ;;
  WARN)  COLOR="$COLOR_WARN"  ;;
  ERROR) COLOR="$COLOR_ERROR" ;;
esac

# Filter logs based on the current log level
case "$LOG_LEVEL" in
  DEBUG) ;; # Show everything
  INFO)  [ "$LEVEL" = "DEBUG" ] && exit 0 ;;
  WARN)  [ "$LEVEL" = "DEBUG" ] || [ "$LEVEL" = "INFO" ] && exit 0 ;;
  ERROR) [ "$LEVEL" != "ERROR" ] && exit 0 ;;
esac

# Print the log message
TIMESTAMP=$(date '+%Y-%m-%d %H:%M:%S')
echo -e "${COLOR}[$TIMESTAMP] [$LEVEL] $MESSAGE${COLOR_RESET}"
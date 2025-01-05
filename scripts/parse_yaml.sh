#!/usr/bin/env bash
# parse_yaml - Parses the YAML config file and outputs shell variables.

CONFIG_FILE="${1:-/etc/ftp/config.yaml}"  # Default config path if none provided

# --- Handle Missing Config File ---
# --- Check if file exists ---
if [ ! -f "$CONFIG_FILE" ]; then
  echo "CONFIG_FILE_DETECTED=0"
  echo "YAML_ADDRESS=''"
  echo "YAML_MIN_PORT=''"
  echo "YAML_MAX_PORT=''"
  echo "YAML_TLS_CERT=''"
  echo "YAML_TLS_KEY=''"
  echo "YAML_USER_COUNT=0"
  exit 0
fi

# --- Validate YAML syntax ---
if ! yq '.' "$CONFIG_FILE" >/dev/null 2>&1; then
  echo "CONFIG_FILE_DETECTED=1"
  echo "YAML_ADDRESS=''"
  echo "YAML_MIN_PORT=''"
  echo "YAML_MAX_PORT=''"
  echo "YAML_TLS_CERT=''"
  echo "YAML_TLS_KEY=''"
  echo "YAML_USER_COUNT=0"
  exit 0
fi

# --- Indicate Config Found ---
echo "CONFIG_FILE_DETECTED=1"

# --- Parse Server Settings ---
server_address=$(yq '.server.address // ""' "$CONFIG_FILE")
server_min_port=$(yq '.server.min_port // ""' "$CONFIG_FILE")
server_max_port=$(yq '.server.max_port // ""' "$CONFIG_FILE")

# --- Parse TLS Settings ---
server_tls_cert=$(yq '.server.tls_cert // ""' "$CONFIG_FILE")
server_tls_key=$(yq '.server.tls_key // ""' "$CONFIG_FILE")

# --- Parse User Count ---
user_count=$(yq '.users | length' "$CONFIG_FILE" 2>/dev/null || echo "0")

# --- Output Variables ---
echo "YAML_ADDRESS='$server_address'"
echo "YAML_MIN_PORT='$server_min_port'"
echo "YAML_MAX_PORT='$server_max_port'"
echo "YAML_TLS_CERT='$server_tls_cert'"
echo "YAML_TLS_KEY='$server_tls_key'"
echo "YAML_USER_COUNT=$user_count"

# --- Parse User Details ---
i=0
while [ "$i" -lt "$user_count" ]; do
  username=$(yq ".users[$i].username // \"\"" "$CONFIG_FILE")
  pass_env=$(yq ".users[$i].password_env // \"\"" "$CONFIG_FILE")

  # Fetch password override from environment variable
  env_val="$(printenv "$pass_env")"

  # --- Output Resolved Variables ---
  echo "YAML_USER_${i}_NAME='$username'"
  echo "YAML_USER_${i}_PASS_ENV='$pass_env'"
  echo "YAML_USER_${i}_PASS_OVERRIDE='${env_val}'"
  i=$((i + 1))
done
#!/usr/bin/env bash

echo "Reached entrypoint"

# --- Logging Function ---
log() {
  local level="$1"
  shift
  local message="$*"
  local timestamp
  timestamp=$(date +"%Y-%m-%d %H:%M:%S")
  echo "[$timestamp] [$level] $message"
}

# --- Verify Required Commands ---
for cmd in parse_yaml log create_user; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    echo "Error: '$cmd' command not found in PATH. Aborting."
    exit 1
  fi
done

# --- Setup Defaults ---
CONFIG_FILE="${CONFIG_FILE:-}"  # Empty by default; explicitly set to use a config
MIN_PORT="${MIN_PORT:-21000}"
MAX_PORT="${MAX_PORT:-21010}"
LOG_LEVEL="${LOG_LEVEL:-INFO}"

# --- TLS Defaults ---
START_TIME="$(date +%s)"
TIMEOUT="${TLS_TIMEOUT:-120}"
TLS_OPT=""

# --- Config Handling ---
if [ -n "$CONFIG_FILE" ] && [ -f "$CONFIG_FILE" ]; then
  log INFO "📄 Config file detected: $CONFIG_FILE"

  # Revert to eval parsing
  eval "$(parse_yaml "$CONFIG_FILE")"

  # Debug: Log parsed YAML variables
  if [ "$LOG_LEVEL" = "DEBUG" ]; then
    log DEBUG "==========================="
    log DEBUG "YAML Configuration Details:"
    log DEBUG "Address: ${YAML_ADDRESS:-None}"
    log DEBUG "Min Port: ${YAML_MIN_PORT:-None}"
    log DEBUG "Max Port: ${YAML_MAX_PORT:-None}"
    log DEBUG "TLS Cert: ${YAML_TLS_CERT:-None}"
    log DEBUG "TLS Key: ${YAML_TLS_KEY:-None}"
    log DEBUG "User Count: ${YAML_USER_COUNT:-0}"
    log DEBUG "==========================="
  fi
else
  log INFO "📄 No config file specified. Using environment variables."
fi

# --- Apply Config Overrides ---
MIN_PORT="${YAML_MIN_PORT:-$MIN_PORT}"
MAX_PORT="${YAML_MAX_PORT:-$MAX_PORT}"
TLS_CERT="${TLS_CERT:-$YAML_TLS_CERT}"
TLS_KEY="${TLS_KEY:-$YAML_TLS_KEY}"

# --- Passive Mode Address ---
ADDRESS="${ADDRESS:-${YAML_ADDRESS:-}}"
if [ -n "$ADDRESS" ]; then
  ADDR_OPT="-opasv_address=$ADDRESS"
fi
log INFO "🔧 Passive Mode Address: ${ADDRESS:-None}"

# --- TLS Check ---
if [ -n "$TLS_CERT" ] || [ -n "$TLS_KEY" ]; then
  log INFO "🔒 TLS is enabled. Checking for cert/key files..."
  while true; do
    if [ -f "$TLS_CERT" ] && [ -f "$TLS_KEY" ]; then
      log INFO "✅ TLS cert and key found. Proceeding with TLS enabled."
      TLS_OPT="-orsa_cert_file=$TLS_CERT -orsa_private_key_file=$TLS_KEY -ossl_enable=YES \
      -oallow_anon_ssl=NO -oforce_local_data_ssl=YES -oforce_local_logins_ssl=YES \
      -ossl_tlsv1=YES -ossl_sslv2=NO -ossl_sslv3=NO -ossl_ciphers=HIGH"
      break
    fi

    ELAPSED=$(( "$(date +%s)" - START_TIME ))
    if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
      log ERROR "❌ TLS cert/key not found after $TIMEOUT seconds. Exiting."
      exit 1
    fi

    if [ $((ELAPSED % 5)) -eq 0 ]; then
      log WARN "⏳ Waiting for TLS cert/key files... ($ELAPSED seconds elapsed)"
    fi
    sleep 1
  done
else
  log WARN "🚧 TLS is not enabled. Proceeding without TLS."
fi

# --- User Setup ---

# Create FTP_USER from Environment Variables if defined
if [ -n "$FTP_USER" ] && [ -n "$FTP_PASS" ]; then
  # Check if FTP_USER is already defined in the YAML
  for i in $(seq 0 $((YAML_USER_COUNT - 1))); do
    eval "USERNAME=\$YAML_USER_${i}_NAME"

    if [ "$USERNAME" = "$FTP_USER" ]; then
      log WARN "🚧 User '$FTP_USER' is defined in both environment variables and config file."
      log WARN "🚧 Ignoring config file values and using FTP_USER and FTP_PASS from environment variables."
      break
    fi
  done

  # Create the user from environment variables
  PASSWORD="$FTP_PASS"
  PASSWORD_PREVIEW="${PASSWORD: -5}" # Debug preview of password

  if [ "$LOG_LEVEL" = "DEBUG" ]; then
    log DEBUG "----------------------------"
    log DEBUG "User (Env):"
    log DEBUG "  Username: $FTP_USER"
    log DEBUG "  Final Password (last 5): ******${PASSWORD_PREVIEW}"
  fi

  create_user "$FTP_USER" "$PASSWORD"
else
  log INFO "🛑 No FTP_USER or FTP_PASS provided. Skipping environment-based user creation."
fi

# Process YAML Users
for i in $(seq 0 $((YAML_USER_COUNT - 1))); do
  eval "USERNAME=\$YAML_USER_${i}_NAME"
  eval "PASS_ENV=\$YAML_USER_${i}_PASS_ENV"

  # Skip if the YAML username matches FTP_USER
  if [ "$USERNAME" = "$FTP_USER" ]; then
    continue
  fi

  # Retrieve password from environment variable
  PASSWORD="$(printenv "$PASS_ENV")"

  if [ -z "$PASSWORD" ]; then
    log ERROR "❌ Password for user '$USERNAME' is missing or empty!"
    exit 1
  fi

  # Debug Logs
  if [ "$LOG_LEVEL" = "DEBUG" ]; then
    PASSWORD_PREVIEW="${PASSWORD: -5}"
    log DEBUG "----------------------------"
    log DEBUG "User [$i]:"
    log DEBUG "  Username: $USERNAME"
    log DEBUG "  Env Variable: $PASS_ENV"
    log DEBUG "  Final Password (last 5): ******${PASSWORD_PREVIEW}"
  fi

  # Create user from YAML
  log INFO "👤 Creating user: $USERNAME"
  create_user "$USERNAME" "$PASSWORD"
done

log DEBUG "🔧 Passive Mode Port Range: $MIN_PORT - $MAX_PORT"
PASV_PORT_OPTS="-opasv_min_port=$MIN_PORT -opasv_max_port=$MAX_PORT"

log INFO "🚀 Starting vsftpd..."
vsftpd $PASV_PORT_OPTS $ADDR_OPT $TLS_OPT /etc/vsftpd/vsftpd.conf

[ -d /var/run/vsftpd ] || mkdir /var/run/vsftpd
pgrep vsftpd | tail -n 1 > /var/run/vsftpd/vsftpd.pid

# --- Signal readiness by creating marker file ---
touch /var/run/ftp-ready
log INFO "✅ FTP server is ready."

exec pidproxy /var/run/vsftpd/vsftpd.pid true
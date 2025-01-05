#!/usr/bin/env bash
# create_user - Adds an FTP user to the system

# Input arguments
NAME="$1"
PASS="$2"

# --- Error Handling ---
if [ -z "$NAME" ] || [ -z "$PASS" ]; then
  log ERROR "Usage: create_user <username> <password>"
  exit 1
fi

if [[ "$NAME" =~ [^a-zA-Z0-9_.-] ]]; then
  log ERROR "Invalid username: '$NAME'. Allowed characters: a-z, A-Z, 0-9, ., -, _"
  exit 1
fi

FTP_DIR="/ftp/$NAME"

# --- Check if user already exists ---
if id "$NAME" &>/dev/null; then
  log ERROR "User '$NAME' already exists"
  exit 1
fi

# --- Check if group already exists ---
if getent group "$NAME" >/dev/null; then
  log ERROR "Group '$NAME' already exists"
  exit 1
fi

# --- Generate UID/GID ---
NEXT_UID=$(($(getent passwd | awk -F: '{print $3}' | sort -n | tail -n 1) + 1))
NEXT_GID=$(($(getent group | awk -F: '{print $3}' | sort -n | tail -n 1) + 1))

# --- Create Group ---
log DEBUG "🔧 Creating group $NAME (GID: $NEXT_GID)"
if ! addgroup -g "$NEXT_GID" "$NAME"; then
  log ERROR "Failed to create group '$NAME'"
  exit 1
fi

# --- Create User ---
log INFO "👤 Adding user: $NAME (UID: $NEXT_UID, GID: $NEXT_GID)"
if ! printf "%s\n%s\n" "$PASS" "$PASS" | adduser -h "$FTP_DIR" -s /sbin/nologin -u "$NEXT_UID" -G "$NAME" "$NAME"; then
  log ERROR "Failed to create user '$NAME'"
  exit 1
fi

# --- Create FTP Directory ---
log DEBUG "📂 Creating FTP directory at $FTP_DIR"
mkdir -p "$FTP_DIR"
if ! chown "$NAME:$NAME" "$FTP_DIR"; then
  log ERROR "Failed to set ownership for '$FTP_DIR'"
  exit 1
fi

if ! chmod 755 "$FTP_DIR"; then
  log ERROR "Failed to set permissions for '$FTP_DIR'"
  exit 1
fi

log INFO "✅ User $NAME created successfully."
exit 0
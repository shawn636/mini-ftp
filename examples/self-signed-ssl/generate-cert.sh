#!/usr/bin/env sh

# This script generates a self-signed certificate for testing purposes.
# It is not intended for production use. You should use a certificate
# signed by a trusted certificate authority (CA) for production use.
# See the /examples directory for an example of how to use a certificate
# signed by a trusted CA, using SWAG with Let's Encrypt or ZeroSSL.

CERT_PATH="/ssl"
CERT_FILE="${CERT_PATH}/fullchain.pem"
KEY_FILE="${CERT_PATH}/privkey.pem"

# Create directory structure if it doesn't exist
mkdir -p "${CERT_PATH}"

# Check if both cert and key already exist
if [ -f "${CERT_FILE}" ] && [ -f "${KEY_FILE}" ]; then
  echo "Certificate and key already exist. Skipping generation."
else
  echo "Generating new self-signed certificate..."

  # Generate self-signed certificate
  openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout "${KEY_FILE}" \
    -out "${CERT_FILE}" \
    -subj "/C=US/ST=CA/L=LA/O=TestOrg/OU=Dev/CN=mini-ftp.duckdns.org"

  echo "Certificates generated at ${CERT_PATH}"
fi

exec tail -f /dev/null
FROM alpine:latest

# Install OpenSSL
RUN apk add --no-cache openssl

# Create a script for generating certs
COPY generate-cert.sh /usr/local/bin/generate-cert.sh
RUN chmod +x /usr/local/bin/generate-cert.sh

# Default entrypoint
ENTRYPOINT ["/usr/local/bin/generate-cert.sh"]
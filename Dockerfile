# --- Stage 1: Build pidproxy ---

# Use Alpine Linux as the base image for building pidproxy
ARG ALPINE_VERSION
ARG BASE_IMG=alpine:${ALPINE_VERSION:-latest}
FROM $BASE_IMG AS pidproxy

# Install dependencies required for building pidproxy
RUN apk --no-cache add alpine-sdk \
    # Clone the latest version of pidproxy
    && git clone https://github.com/ZentriaMC/pidproxy.git \
    # Enter the repository directory
    && cd pidproxy \
    # Optimize build for native CPU architecture
    && sed -i 's/-mtune=generic/-mtune=native/g' Makefile \
    # Build the pidproxy binary 
    && make \      
    # Move binary to a system path
    && mv pidproxy /usr/bin/pidproxy \
    # Go back to the root directory
    && cd .. \
    # Clean up build files
    && rm -rf pidproxy \
    # Remove build dependencies
    && apk del alpine-sdk                                        

# --- Stage 2: Final Image ---

# Use a clean Alpine image as the runtime environment
FROM $BASE_IMG

# Copy the compiled pidproxy binary from the build stage
COPY --from=pidproxy /usr/bin/pidproxy /usr/bin/pidproxy

# Install runtime dependencies
RUN apk --no-cache add vsftpd tini bash shadow jq curl \
    && curl -sL $(curl -s https://api.github.com/repos/mikefarah/yq/releases/latest | jq -r '.assets[] | select(.name | contains("linux_amd64")) | .browser_download_url') -o /usr/bin/yq \
    && chmod +x /usr/bin/yq

COPY config/vsftpd.conf /etc/vsftpd/vsftpd.conf


COPY scripts/ /bin/
RUN for f in /bin/*.sh; do \
    chmod +x "$f" && \
    mv "$f" "bin/$(basename "$f" .sh)"; \
    done

# Set permissions for the startup script and FTP root directory
RUN mkdir -p /ftp \
    && chmod 755 /ftp

# Create the chroot_list file and add root user
RUN touch /etc/vsftpd/chroot_list \
    && chmod 644 /etc/vsftpd/chroot_list \
    && echo "root" >> /etc/vsftpd/chroot_list

# Prepares the FTP directory and script for execution
# Ensures correct permissions to avoid permission errors


# Expose ports:
EXPOSE 21 21000-21010

# - 21: FTP control port
# - 21000-21010: Passive mode data transfer ports


# This healthceck is admittedly a bit hacky
# the vsftpd process automatically runs in the background
# even before we initialize TLS, so we can't rely on the process
# alone to indicate readiness. Instead, we create a file
# /var/run/ftp-ready to indicate that the FTP server is ready
# Ideally, we would rework the entrypoint script so that
# the vsftpd process only starts after TLS is initialized.
HEALTHCHECK --interval=15s --timeout=10s --start-period=60s --retries=3 \
    CMD sh -c '[ -f /var/run/ftp-ready ] && [ -f /var/run/vsftpd/vsftpd.pid ] && \
    cat /var/run/vsftpd/vsftpd.pid | xargs -I{} sh -c "ps | grep \" {} \" > /dev/null" || exit 1'

# Healthcheck:
# - Verifies the PID file exists and the vsftpd process is running
# - Ensures Docker can detect failures and restart the container if needed


# Entrypoint using tini
ENTRYPOINT ["/sbin/tini", "--", "/bin/docker-entrypoint"]

# Entrypoint:
# - Uses tini as PID 1 for signal handling and zombie reaping
# - Launches the vsftpd startup script
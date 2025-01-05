# This is a copy of our main Dockerfile but without the entrypoint.
# We instead make the cmd /bin/sh so that we can test scripts
# from inside the env where they will actually be run.


ARG ALPINE_VERSION
ARG BASE_IMG=alpine:${ALPINE_VERSION:-latest}
FROM $BASE_IMG AS pidproxy

RUN apk --no-cache add alpine-sdk \
    && git clone https://github.com/ZentriaMC/pidproxy.git \
    && cd pidproxy \
    && sed -i 's/-mtune=generic/-mtune=native/g' Makefile \
    && make \      
    && mv pidproxy /usr/bin/pidproxy \
    && cd .. \
    && rm -rf pidproxy \
    && apk del alpine-sdk                                        

FROM $BASE_IMG
COPY --from=pidproxy /usr/bin/pidproxy /usr/bin/pidproxy
RUN apk --no-cache add vsftpd tini bash shadow jq curl \
    && curl -sL $(curl -s https://api.github.com/repos/mikefarah/yq/releases/latest | jq -r '.assets[] | select(.name | contains("linux_amd64")) | .browser_download_url') -o /usr/bin/yq \
    && chmod +x /usr/bin/yq
COPY scripts/ /bin/
RUN for f in /bin/*.sh; do \
    chmod +x "$f" && \
    mv "$f" "bin/$(basename "$f" .sh)"; \
    done
RUN mkdir -p /ftp \
    && chmod 755 /ftp
RUN touch /etc/vsftpd/chroot_list \
    && chmod 644 /etc/vsftpd/chroot_list \
    && echo "root" >> /etc/vsftpd/chroot_list
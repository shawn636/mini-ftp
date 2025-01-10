![logo](https://github.com/shawn636/mini-ftp/blob/main/assets/logo.png?raw=true)

[![Docker Stars](https://img.shields.io/docker/stars/shawn636/mini-ftp.svg)](https://hub.docker.com/r/delfer/alpine-ftp-server/) [![Docker Pulls](https://img.shields.io/docker/pulls/shawn636/mini-ftp.svg)](https://hub.docker.com/r/shawn636/mini-ftp.svg)  [![Tests](https://github.com/shawn636/mini-ftp/actions/workflows/test.yaml/badge.svg)](https://github.com/shawn636/mini-ftp/actions/workflows/test.yaml) [![Automated Builds](https://github.com/shawn636/mini-ftp/actions/workflows/build-and-release.yaml/badge.svg)](https://github.com/shawn636/mini-ftp/actions/workflows/build-and-release.yaml)

Small and flexible docker image with vsftpd server

# mini-ftp

A lightweight FTP server with support for YAML-based configuration and secure password handling.



## **Why Choose mini-ftp?**

- **Multi-Platform Ready**
  - Supports a wide range of architectures: amd64, arm64, arm/v7, arm/v6, ppc64le, and s390x. Perfect for modern cloud environments, edge devices, and legacy systems.

- **Scalable User Management**
  - Manage single or multiple FTP users effortlessly with environment variables or YAML-based configurations.

- **Security First**
  - TLS support and environment-variable-backed passwords keep your configurations secure and version-control safe.

- **Lightweight and Fast**
  - Built on Alpine Linux, mini-ftp has a small footprint while delivering blazing performance.



## Key Features

- **Multi-Architecture Support**
  - Tested and validated on multiple architectures to ensure compatibility across a variety of systems.

- **Quick Start**
  - Spin up an FTP server with a single user in seconds using environment variables.

- **Advanced Configurations**
  - Leverage YAML for scalable setups, multiple users, and advanced server settings.

- **TLS Encryption**
  - Secure connections are easy to enable with TLS certificates and keys.

- **Portable and Predictable**
  - Files are always served from /ftp, simplifying directory mappings.



## Multi-Platform and Multi-Architecture Support

**mini-ftp** is tested on the following architectures:

- `amd64`: Ideal for standard servers and x86 cloud instances.

- `arm64`: Optimized for modern ARM-based devices and cloud platforms like AWS Graviton.

- `arm/v7`, `arm/v6`: Perfect for older Raspberry Pi models and IoT/embedded devices.

- `ppc64le`, `s390x`: Ensures compatibility with enterprise-grade IBM Power Systems and mainframes.

Our CI/CD pipeline rigorously tests the image on all supported architectures, ensuring reliability across platforms.



## Getting Started

### Quick Start

Launch an FTP server with minimal configuration:

```bash
docker run -d \
    -p "21:21" \
    -p 21000-21010:21000-21010 \
    -e FTP_USER=one \
    -e FTP_PASS=1234 \
    -v "$(pwd)/ftp-data:/ftp" \
    shawn636/mini-ftp
```

### Advanced Configuration

Harness the power of YAML for multi-user setups and additional features. See the [Configuration](#configuration) section below for details.

### Examples

Looking for inspiration or ready-to-use configurations? Check out the [**examples/** directory](https://github.com/shawn636/mini-ftp/tree/main/examples) for:

- **Docker Compose Files** for single-user and multi-user setups.
- **YAML Configurations** for advanced server and user management.
- **TLS-Enabled Deployments** using secure connections.

These examples cover a variety of use cases, helping you get up and running quickly with configurations tailored to your needs.



## Configuration

### Environment Variables

#### Server Settings

- `ADDRESS` – External address for passive ports (optional, should resolve to the server’s IP).

- `MIN_PORT` – Minimum passive port (optional, default 21000).

- `MAX_PORT` – Maximum passive port (optional, default 21010).

- `CONFIG_FILE` – Path to a YAML config file for additional settings and users (optional).



#### Single User Settings

- `FTP_USER` – Default username for quick setups (required if no config file).

- `FTP_PASS` – Password for the default user (required if no config file).

- `FTP_HOME` – Fixed home directory for the default user (always /ftp).

- `FTP_UID` – User ID for the default user (optional, default 1000).

- `FTP_GID` – Group ID for the default user (optional, default 1000).



#### TLS Settings

- `TLS_CERT` - Path to the TLS certificate file. Enables FTPS if set.
- `TLS_KEY` - Path to the TLS private key file. Required if `TLS_CERT` is set.
- `TLS_TIMEOUT` - Timeout (in seconds) to wait for TLS cert/key to appear (default: )



## YAML Config File

For advanced setups with multiple users and server settings, you can use a YAML config file referenced via the `CONFIG_FILE` environment variable.



### Acceptable Keys and Values in config.yaml

#### Server Settings

| Key           | Description                                                  | Required | Default                     |
| ------------- | ------------------------------------------------------------ | -------- | --------------------------- |
| `address`     | External address for passive ports (should resolve to server's IP). | No       | Derived from container's IP |
| `min_port`    | Minimum port number for passive connections                  | No       | 21000                       |
| `max_port`    | Maximum port number for passive connections                  | No       | 21010                       |
| `tls_cert`    | The **path** to the TLS certificate file for enabling encrypted connections. | No       | None                        |
| `tls_key`     | The **path** to the TLS private key file for enabling encrypted connections. | No       | None                        |
| `tls_timeout` | Timeout (in seconds) to wait for TLS cert and key to appear  | No       | 120                          |

**Note**: If `tls_cert` and `tls_key` are both provided, SFTP is automatically enabled.



#### User Settings
| Key        | Description                                                  | Required | Default                     |
| ---------- | ------------------------------------------------------------ | -------- | --------------------------- |
| `username` | Username for FTP access. | Yes    | None |
| `password_env` | **Name of env variable** containing the user's password. | Yes    | None                    |
| `uid` | User ID for the account. | No       | Increments from 1000 |
| `gid` | Group ID for the account.                                | No       | Increments from 1000   |

**Note:** Passwords must always be stored in environment variables and referenced here using `password_env` for security.




**Example** config.yaml

```yaml
server:
  address: ftp.site.domain
  min_port: 21000
  max_port: 21010
  tls_cert: /etc/ssl/certs/server-cert.pem
  tls_key: /etc/ssl/private/server-key.pem

users:
  - username: user1
    password_env: USER1_PASS  # Uses environment variable for security
    uid: 1001
    gid: 1001
  - username: user2
    password_env: USER2_PASS
    uid: 1002
    gid: 1002
  - username: guest
    password_env: GUEST_PASS
    uid: 2000
    gid: 2000
```



#### Config Validation and Error Handling

- **Missing Passwords:** If a password_env variable is missing or undefined, the server logs a warning and skips the user during initialization.

- **Invalid Ports:** If min_port or max_port are invalid or out of range, defaults are used instead.

- **TLS Errors:** If either tls_cert or tls_key is missing when the other is provided, the server logs an error and disables FTPS.

- **Unknown Keys:** Any unknown keys in the config file are ignored, and a warning is logged.



#### Key Notes

1. **Environment-Backed Security** – Passwords are never stored in the config file directly.

2. **Simplified TLS Setup** – Just drop in a cert and key, and FTPS is enabled automatically.

3. **Validation at Startup** – Logs warnings for invalid configurations instead of failing outright.

4. **Scalable Design** – Easily supports adding more users by updating the YAML config.



## Example Password Storage with .env

Create a .env file to securely manage passwords:

```bash
USER1_PASS=supersecurepassword
USER2_PASS=anothersecurepassword
GUEST_PASS=guestpassword
```



Then use it with Docker Compose

```bash
docker compose --env-file .env up -d
```



## SFTP with SSL via SWAG (Secure Web Application Gateway)

For TLS-enabled setups, you can use **SWAG** to handle SSL certificates. Refer to the [SWAG GitHub Repository](https://github.com/linuxserver/docker-swag) for more details.



**Example docker-compose.yaml with SWAG:**

```yaml
services:
  swag:
    image: lscr.io/linuxserver/swag
    container_name: swag
    cap_add:
      - NET_ADMIN
    environment:
      - PUID=1000
      - PGID=1000
      - URL=ftp.site.domain
      - VALIDATION=http
      - EMAIL=your@email.com
    volumes:
      - ./config:/config
    ports:
      - 443:443
      - 80:80
    restart: unless-stopped

  ftp:
    image: shawnlong636/mini-ftp
    container_name: ftp
    ports:
      - "21:21"
      - "21000-21010:21000-21010"
    environment:
      - CONFIG_FILE=/etc/ftp/config.yaml
    volumes:
      - ./ftp-data:/ftp
      - ./config.yaml:/etc/ftp/config.yaml
      - ./config:/config
    restart: unless-stopped
```



## **Example Scenarios**

#### IoT Devices (ARM/v6 and ARM/v7)

Deploy mini-ftp on low-powered IoT devices for fast, reliable file sharing.

```bash
docker run --platform linux/arm/v6 \
    -e FTP_USER=pi \
    -e FTP_PASS=raspberry \
    shawn636/mini-ftp
```



#### Enterprise Servers (PPC64LE and S390X)

Run mini-ftp on high-end servers for secure, scalable FTP services.

```bash
docker run --platform linux/ppc64le \
    -e FTP_USER=admin \
    -e FTP_PASS=supersecret \
    shawn636/mini-ftp
```



## Badges of Honor

We take pride in:

- **Automated Multi-Arch Builds**
  - Every release is built and validated for all supported platforms.

- **Comprehensive Testing**
  - Rigorous CI ensures stability across configurations and architectures.

- **Community-Driven Development**
  - Contributions and feedback are welcome!

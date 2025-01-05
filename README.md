![logo](/Users/shawnlong/Repos/Personal/mini-ftp/assets/logo.png)

# mini-ftp

A lightweight FTP server with support for YAML-based configuration and secure password handling.



## Key Features

- **Single User Quick Start** – Use environment variables for a simple setup.

- **Multiple Users with YAML Config** – Add multiple users and server settings using a YAML file.

- **Secure Passwords** – Reference passwords via environment variables to keep secrets out of config files.

- **Standardized FTP Root** – Files are always served from /ftp, making configuration predictable.

- **TLS Support** – Easily enable secure connections using a TLS certificate and key.



## Usage

### Quick Start with Default User

```bash
docker run -d \
    -p "21:21" \
    -p 21000-21010:21000-21010 \
    -e FTP_USER=one \
    -e FTP_PASS=1234 \
    -v "$(pwd)/ftp-data:/ftp" \
    shawnlong636/mini-ftp
```



### Using Docker Compose

```yaml
services:
  ftp:
    image: shawnlong636/mini-ftp
    ports:
      - "21:21"
      - "21000-21010:21000-21010"
    environment:
      - FTP_USER=one
      - FTP_PASS=1234
      - CONFIG_FILE=/etc/ftp/config.yaml
    volumes:
      - ./ftp-data:/ftp
      - ./config.yaml:/etc/ftp/config.yaml
```



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



### Key Notes

- **Default User Setup** – Environment variables are perfect for quick tests and single-user setups.

- **Advanced Config with YAML** – Use a config file for larger setups and multi-user environments.

- **Security First** – Passwords are stored in environment variables, keeping configs safe for version control.

- **TLS Ready** – Supports encrypted connections via certificates.

- **SWAG Integration** – Simplifies SSL certificate management with automatic renewal.

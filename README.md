# goappmon

`goappmon` is a lightweight application control center for mobile and web applications.

It ships as a single Go binary with SQLite storage, Gin HTTP handlers, bcrypt-backed admin auth, server-rendered HTML, and JSON APIs for app status, version policy, and feature flags.

## Start Here

- [Project overview](docs/project-overview.md)
- [Tech stack](docs/tech-stack.md)
- [Architecture](docs/architecture.md)
- [Development guide](docs/development.md)
- [Deployment guide](docs/deployment.md)
- [Testing guide](docs/testing.md)
- [Release guide](docs/release.md)
- [Security policy](docs/security.md)
- [Contributing guide](docs/contributing.md)
- [Roadmap](docs/roadmap.md)
- [Docs index](docs/README.md)
- [Contributor guide](CONTRIBUTING.md)
- [Code of conduct](CODE_OF_CONDUCT.md)
- [Security reporting](SECURITY.md)
- [Changelog](CHANGELOG.md)

## Current Status

- Module path: `github.com/phyowaiyan-dev/goappmon`
- Go version: `1.26.1`
- License: MIT
- Storage: `storage/goappmon.sqlite`
- Server-rendered admin UI: yes
- Public JSON APIs: yes
- Source code: implemented MVP

## Repository Contents

- `go.mod` - module definition
- `LICENSE` - MIT license
- `README.md` - landing page and docs entry point
- `docs/` - production-oriented project documentation
- `CONTRIBUTING.md` - contributor workflow and standards
- `CODE_OF_CONDUCT.md` - community behavior policy
- `SECURITY.md` - vulnerability reporting and security contact guidance
- `CHANGELOG.md` - release history placeholder
- `internal/` - private implementation notes and work items

## Build

```bash
go build ./...
```

## Run

```bash
go run ./cmd/goappmon
```

## GitHub Releases

Every tagged release triggers `.github/workflows/release.yml` and publishes prebuilt archives for:

- Linux `amd64`
- Linux `arm64`
- macOS `amd64`
- macOS `arm64`
- Windows `amd64`

The archives contain a single runnable binary for the target platform, so production servers do not need to build from source.

## Ubuntu Deployment

`goappmon` is designed to run as a single Go binary on Ubuntu Server behind either Apache2 or Nginx.

### 1. Install prerequisites

```bash
sudo apt update
sudo apt install -y git build-essential ca-certificates
```

If you are deploying from a GitHub Release, download the Linux binary archive from the Releases page instead of building from source.

```bash
sudo mkdir -p /opt/goappmon
curl -L -o /tmp/goappmon-linux-amd64.tar.gz "https://github.com/phyowaiyan-dev/goappmon/releases/latest/download/goappmon-linux-amd64.tar.gz"
tar -xzf /tmp/goappmon-linux-amd64.tar.gz -C /tmp
sudo install -m 755 /tmp/goappmon /opt/goappmon/goappmon
```

### 2. Create runtime directories

The app stores SQLite data and its session secret under `storage/`.

```bash
sudo mkdir -p /opt/goappmon/storage
sudo chown -R www-data:www-data /opt/goappmon
```

### 3. Run the app

The default address is `:18180`.

```bash
GOAPPMON_ADDR=:18180 /opt/goappmon/goappmon
```

You can also override the defaults with environment variables:

```bash
export GOAPPMON_ADDR=:18180
export GOAPPMON_DB_PATH=/opt/goappmon/storage/goappmon.sqlite
export GOAPPMON_SESSION_KEY_PATH=/opt/goappmon/storage/session.key
export GOAPPMON_LOG_LEVEL=info
/opt/goappmon/goappmon
```

### 4. Point your domain to the server

For your root domain and any subdomain, create `A` records that point to your Ubuntu server public IPv4 address.

Example:

- `example.com` -> `203.0.113.10`
- `www.example.com` -> `203.0.113.10`
- `admin.example.com` -> `203.0.113.10`

If you use IPv6, add `AAAA` records too.

Make sure your DNS provider has fully propagated before requesting SSL certificates.

### 5. Open firewall ports

Allow HTTP and HTTPS traffic to the server:

```bash
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw allow OpenSSH
sudo ufw enable
```

### 6. Run as a systemd service

Create `/etc/systemd/system/goappmon.service`:

```ini
[Unit]
Description=GoAppMon
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/opt/goappmon
Environment=GOAPPMON_ADDR=127.0.0.1:18180
Environment=GOAPPMON_DB_PATH=/opt/goappmon/storage/goappmon.sqlite
Environment=GOAPPMON_SESSION_KEY_PATH=/opt/goappmon/storage/session.key
Environment=GOAPPMON_LOG_LEVEL=info
ExecStart=/opt/goappmon/goappmon
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Then enable it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now goappmon
sudo systemctl status goappmon
```

### 7. Reverse proxy with Nginx

Install Nginx:

```bash
sudo apt install -y nginx
```

Create `/etc/nginx/sites-available/goappmon`:

```nginx
server {
    listen 80;
    server_name example.com www.example.com admin.example.com;

    location / {
        proxy_pass http://127.0.0.1:18180;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Enable and reload:

```bash
sudo ln -s /etc/nginx/sites-available/goappmon /etc/nginx/sites-enabled/goappmon
sudo nginx -t
sudo systemctl reload nginx
```

### 8. Reverse proxy with Apache2

Install Apache2 and enable required modules:

```bash
sudo apt install -y apache2
sudo a2enmod proxy proxy_http headers rewrite ssl
```

Create a virtual host, for example `/etc/apache2/sites-available/goappmon.conf`:

```apache
<VirtualHost *:80>
    ServerName example.com
    ServerAlias www.example.com admin.example.com

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:18180/
    ProxyPassReverse / http://127.0.0.1:18180/

    RequestHeader set X-Forwarded-Proto "http"
</VirtualHost>
```

Enable and reload:

```bash
sudo a2ensite goappmon
sudo apache2ctl configtest
sudo systemctl reload apache2
```

### 9. Add SSL with Certbot

For Nginx:

```bash
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d example.com -d www.example.com -d admin.example.com
```

For Apache2:

```bash
sudo apt install -y certbot python3-certbot-apache
sudo certbot --apache -d example.com -d www.example.com -d admin.example.com
```

Certbot will renew automatically on Ubuntu through systemd timers. You can verify with:

```bash
sudo certbot renew --dry-run
```

### 10. Final checks

After deployment, confirm:

- `https://example.com` loads the app
- `https://example.com/admin/login` opens the login page
- `storage/goappmon.sqlite` exists and is writable by the service user
- the reverse proxy forwards requests to `127.0.0.1:18180`

## Deployment Notes

- Keep the binary and `storage/` directory together on the server.
- Do not expose the Go app directly to the internet when using Apache2 or Nginx.
- Use HTTPS for all production traffic.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for the full text.

# Deployment Guide

## Overview

`goappmon` runs as a single Go binary and stores its data in SQLite.
The recommended production setup on Ubuntu is:

- Go binary on the server
- SQLite database and session key under `storage/`
- systemd to manage the process
- Apache2 or Nginx as a reverse proxy
- Certbot for SSL certificates

## Default Runtime Paths

- database: `storage/goappmon.sqlite`
- session key: `storage/session.key`
- default address: `:18180`

## Prepare the Server

Install packages:

```bash
sudo apt update
sudo apt install -y git build-essential ca-certificates
```

Install Go `1.26.1` or newer, then clone and build:

```bash
git clone https://github.com/phyowaiyan-dev/goappmon.git
cd goappmon
go build -o goappmon ./cmd/goappmon
```

If you prefer the production binary from GitHub Releases, download the Linux archive for your server architecture and install it into `/opt/goappmon/` instead of building from source.

Create the storage directory:

```bash
mkdir -p storage
chmod 755 storage
```

## DNS Setup

Point your domain and subdomains to the Ubuntu server with `A` records.

Example:

- `example.com` -> server public IPv4
- `www.example.com` -> server public IPv4
- `admin.example.com` -> server public IPv4

If your server has IPv6, add `AAAA` records too.

Wait for DNS propagation before issuing SSL certificates.

## Run with systemd

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

Enable it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now goappmon
sudo systemctl status goappmon
```

## Reverse Proxy with Nginx

Install Nginx:

```bash
sudo apt install -y nginx
```

Example site file:

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

## Reverse Proxy with Apache2

Install Apache2:

```bash
sudo apt install -y apache2
sudo a2enmod proxy proxy_http headers rewrite ssl
```

Example virtual host:

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

## SSL with Certbot

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

Test renewal:

```bash
sudo certbot renew --dry-run
```

## Firewall

Allow required ports:

```bash
sudo ufw allow OpenSSH
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable
```

## Post-Deployment Checklist

- app starts without errors
- `storage/goappmon.sqlite` is created
- `storage/session.key` is created
- reverse proxy points to `127.0.0.1:18180`
- DNS `A` records resolve correctly
- HTTPS certificate is active

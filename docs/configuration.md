# Configuration

Domicile is configured through environment variables.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `1323` | Port the web server listens on |
| `DOMICILE_NO_BROWSER` | _(unset)_ | Set to any value to prevent auto-opening a browser on startup (always set this in Docker) |

## Volumes

| Path | Description |
|------|-------------|
| `/app/data` | SQLite database storage. Mount this to persist your data across container restarts and updates. |

## Ports

| Port | Description |
|------|-------------|
| `1323` | Web UI (HTTP) |

## Reverse Proxy

If running behind a reverse proxy (Nginx, Caddy, Traefik), point it at port 1323. Example with Caddy:

```
domicile.yourdomain.com {
    reverse_proxy localhost:1323
}
```

Example with Nginx:

```nginx
server {
    listen 80;
    server_name domicile.yourdomain.com;

    location / {
        proxy_pass http://localhost:1323;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Data & Backups

The SQLite database is stored at `/app/data/domicile.db`. To back up:

```bash
# Copy the database file
cp ./data/domicile.db ./domicile-backup-$(date +%Y%m%d).db
```

Or use the built-in backup/restore feature under **Settings → Backup**.

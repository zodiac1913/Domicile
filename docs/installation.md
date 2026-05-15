# Installation

## Docker (Recommended)

### Docker Compose

Create a `compose.yml`:

```yaml
services:
  domicile:
    image: ghcr.io/zodiac1913/domicile:latest
    container_name: domicile
    restart: unless-stopped
    ports:
      - "1323:1323"
    volumes:
      - ./data:/app/data
    environment:
      - PORT=1323
      - DOMICILE_NO_BROWSER=1
```

Start the container:

```bash
docker compose up -d
```

### Docker Run

```bash
docker run -d \
  --name domicile \
  --restart unless-stopped \
  -p 1323:1323 \
  -v ./data:/app/data \
  -e DOMICILE_NO_BROWSER=1 \
  ghcr.io/zodiac1913/domicile:latest
```

### First Run

1. Navigate to `http://your-server:1323`
2. You'll be redirected to the setup page to create your first admin account
3. Start adding your products, locations, and recipes

### Existing Database

If you have an existing `domicile.db`, place it in your data directory before starting:

```bash
mkdir -p ./data
cp /path/to/your/domicile.db ./data/
docker compose up -d
```

---

## Build from Source

### Prerequisites

- Go 1.25+
- [Templ CLI](https://templ.guide/quick-start/installation) v0.3.1001+

### Steps

```bash
git clone https://github.com/zodiac1913/Domicile.git
cd Domicile

go mod download
templ generate
go run server.go
```

The app finds an available port starting at 1323 and opens your browser automatically.

---

## Portainer

1. Go to **Stacks** → **Add Stack**
2. Paste the Docker Compose YAML from above
3. Deploy the stack
4. Access Domicile on port 1323

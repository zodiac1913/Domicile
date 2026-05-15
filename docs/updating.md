# Updating

## Docker Compose

```bash
# Pull the latest image
docker compose pull

# Recreate the container with the new image
docker compose up -d
```

Your data is safe — it lives in the mounted `./data` volume, not inside the container.

## Docker Run

```bash
docker pull ghcr.io/zodiac1913/domicile:latest
docker stop domicile
docker rm domicile
docker run -d \
  --name domicile \
  --restart unless-stopped \
  -p 1323:1323 \
  -v ./data:/app/data \
  -e DOMICILE_NO_BROWSER=1 \
  ghcr.io/zodiac1913/domicile:latest
```

## Before Updating

It's good practice to back up your database before major updates:

```bash
cp ./data/domicile.db ./data/domicile-backup-$(date +%Y%m%d).db
```

## Version Pinning

Instead of `latest`, you can pin to a specific version tag:

```yaml
image: ghcr.io/zodiac1913/domicile:v1.0.0
```

Check the [releases page](https://github.com/zodiac1913/Domicile/releases) for available versions.

# Docker Quick Start Guide

This guide will help you quickly get Timeship running with Docker.

## Prerequisites

- Docker installed on your system
- (Optional) Docker Compose for easier management

## Quick Start

### Option 1: Using Task (Recommended)

If you have [Task](https://taskfile.dev/) installed:

```bash
# Build the Docker image
task docker:build

# Run the container
task docker:run

# Or use docker-compose
task docker:compose:up
```

### Option 2: Using Docker directly

```bash
# Build the image
docker build -t timeship:latest .

# Run the container
docker run -d \
  --name timeship \
  -p 8080:8080 \
  -v /your/zfs/pool:/data:ro \
  timeship:latest
```

### Option 3: Using Docker Compose

1. Edit `docker-compose.yml` to configure your volume mounts:
   ```yaml
   volumes:
     - /your/zfs/pool:/data:ro
   ```

2. Start the service:
   ```bash
   docker compose up -d
   ```

3. View logs:
   ```bash
   docker compose logs -f
   ```

4. Stop the service:
   ```bash
   docker compose down
   ```

## Configuration

### Mounting Volumes

To access your ZFS snapshots, you need to mount them into the container:

```bash
docker run -d \
  -p 8080:8080 \
  -v /tank/dataset:/data:ro,z \
  -v /tank/dataset/.zfs/snapshot:/data/.zfs/snapshot:ro,z \
  timeship:latest
```

**Important:** 
- Use read-only (`:ro`) mounts for safety.
- On SELinux systems (Fedora, RHEL, CentOS), add `:z` flag to allow container access.

### Environment Variables

Configure Timeship using environment variables:

- `TIMESHIP_ROOT` - Root directory to serve (default: `/data`)
- `TIMESHIP_API_PREFIX` - API path prefix (default: `/api`)

Example:
```bash
docker run -d \
  -p 8080:8080 \
  -e TIMESHIP_ROOT=/data \
  -e TIMESHIP_API_PREFIX=/api \
  -v /tank/dataset:/data:ro,z \
  timeship:latest
```

### Using .env file with Docker Compose

Create a `.env` file in the project root:

```env
TIMESHIP_ROOT=/data
TIMESHIP_API_PREFIX=/api
```

Docker Compose will automatically load these variables.

## Accessing Timeship

Once running, open your browser to:
- http://localhost:8080

## Building for Multiple Architectures

To build for different platforms:

```bash
# Build for linux/amd64
docker buildx build --platform linux/amd64 -t timeship:amd64 .

# Build for linux/arm64
docker buildx build --platform linux/arm64 -t timeship:arm64 .

# Build multi-arch image
docker buildx build --platform linux/amd64,linux/arm64 -t timeship:latest .
```

## Troubleshooting

### Container won't start

Check logs:
```bash
docker logs timeship
```

Or with docker-compose:
```bash
docker compose logs
```

### Can't access snapshots

Make sure you've mounted the snapshot directory with read permissions:
```bash
ls -la /tank/dataset/.zfs/snapshot
```

### Permission denied when accessing mounted volumes (SELinux)

If you're running on a system with SELinux (Fedora, RHEL, CentOS, etc.) and get "Permission denied" errors when the container tries to access mounted volumes:

**Solution 1: Add SELinux labels to volume mounts (Recommended)**

Add the `:z` flag to your volume mounts in `docker-compose.yml`:
```yaml
volumes:
  - /your/data:/data:ro,z
```

Or with docker run:
```bash
docker run -d -p 8080:8080 -v /your/data:/data:ro,z timeship:latest
```

**What the flags mean:**
- `:z` - Shared content label (multiple containers can access)
- `:Z` - Private unshared label (only this container can access)

**Solution 2: Check SELinux status**
```bash
getenforce  # Check if SELinux is enforcing
```

**Solution 3: Verify the container can access the mount**
```bash
docker compose run --rm timeship ls -la /data
```

### Port already in use

Change the host port mapping:
```bash
docker run -d -p 3000:8080 -v /data:/data:ro timeship:latest
```

## Advanced Usage

### Running with custom configuration

```bash
docker run -d \
  --name timeship \
  -p 8080:8080 \
  -v /tank/dataset:/data:ro \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  timeship:latest
```

### Health checks

Add a health check to your docker-compose.yml:

```yaml
services:
  timeship:
    # ... other config ...
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## Production Deployment

For production use, consider:

1. **Use a reverse proxy** (nginx, Traefik, Caddy) for HTTPS
2. **Set resource limits**:
   ```yaml
   services:
     timeship:
       deploy:
         resources:
           limits:
             cpus: '2'
             memory: 1G
           reservations:
             cpus: '0.5'
             memory: 256M
   ```
3. **Enable logging**:
   ```yaml
   services:
     timeship:
       logging:
         driver: "json-file"
         options:
           max-size: "10m"
           max-file: "3"
   ```

## Updating

To update to the latest version:

```bash
# Pull latest code
git pull

# Rebuild image
task docker:build

# Restart container
docker compose down
docker compose up -d
```

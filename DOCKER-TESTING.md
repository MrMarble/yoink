# Docker Testing Guide

This guide helps you set up a complete testing environment for Yoink using Docker Compose with both Prowlarr and Jackett indexers.

## Quick Start

1. **Start the services:**
   ```bash
   docker-compose up -d
   ```

2. **Access the web interfaces:**
   - qBittorrent: http://localhost:8080 (admin/adminadmin)
   - Prowlarr: http://localhost:8081
   - Jackett: http://localhost:9117

3. **Configure the services** (see detailed setup below)

4. **Test yoink:**
   ```bash
   # Build yoink
   go build ./cmd/yoink
   
   # Test with Prowlarr
   ./yoink --config ./config.docker.prowlarr.yaml --dry-run
   
   # Test with Jackett
   ./yoink --config ./config.docker.jackett.yaml --dry-run
   ```

## Detailed Setup

### 1. qBittorrent Setup

After starting the services, access qBittorrent at http://localhost:8080:

- **Default credentials:** admin/adminadmin
- The service is pre-configured to work with yoink
- Downloads will be saved to `./downloads/` directory

### 2. Prowlarr Setup

Access Prowlarr at http://localhost:8081:

1. Complete the initial setup wizard
2. Add indexers in **Settings > Indexers**
3. Copy the **API Key** from **Settings > General**
4. Update `config.docker.prowlarr.yaml` with your API key:
   ```yaml
   prowlarr:
     host: "http://prowlarr:9696"
     api_key: "YOUR_PROWLARR_API_KEY_HERE"  # Replace with actual key
   ```
5. Update the indexer IDs in the config to match your Prowlarr setup

### 3. Jackett Setup

Access Jackett at http://localhost:9117:

1. Copy the **API Key** from the top-right corner
2. Add indexers by clicking **+ Add indexer**
3. Configure each indexer with your tracker credentials
4. Update `config.docker.jackett.yaml` with your API key:
   ```yaml
   jackett:
     host: "http://jackett:9117"
     api_key: "YOUR_JACKETT_API_KEY_HERE"  # Replace with actual key
   ```
5. Update the indexer IDs to match your Jackett indexer names

## Testing Commands

```bash
# Build yoink
go build ./cmd/yoink

# Test configuration parsing
./yoink print-config --config ./config.docker.prowlarr.yaml
./yoink print-config --config ./config.docker.jackett.yaml

# List available indexers
./yoink indexers --config ./config.docker.prowlarr.yaml
./yoink indexers --config ./config.docker.jackett.yaml

# Dry run (safe testing)
./yoink --config ./config.docker.prowlarr.yaml --dry-run
./yoink --config ./config.docker.jackett.yaml --dry-run

# Real run (downloads torrents)
./yoink --config ./config.docker.prowlarr.yaml
./yoink --config ./config.docker.jackett.yaml
```

## Running Yoink in Docker

Uncomment the `yoink` service in `docker-compose.yml` to run yoink inside Docker:

```bash
# Edit docker-compose.yml and uncomment the yoink service
# Then rebuild and run
docker-compose down
docker-compose up --build -d
```

## Troubleshooting

### Connection Issues
- Ensure all services are running: `docker-compose ps`
- Check service logs: `docker-compose logs [service-name]`
- Verify network connectivity: `docker-compose exec yoink ping prowlarr`

### API Key Issues
- Double-check API keys are correctly copied from web interfaces
- Ensure no extra spaces or characters in configuration files
- Test API access directly with curl:
  ```bash
  # Test Prowlarr
  curl "http://localhost:8081/api/v1/indexer" -H "X-Api-Key: YOUR_API_KEY"
  
  # Test Jackett  
  curl "http://localhost:9117/api/v2.0/indexers/all/results" -H "X-Api-Key: YOUR_API_KEY"
  ```

### Permissions
- Downloads directory: `mkdir -p downloads && chmod 755 downloads`
- Docker volumes: All volumes use PUID/PGID 1000

## Cleanup

```bash
# Stop services
docker-compose down

# Remove volumes (deletes all configuration)
docker-compose down -v

# Remove downloaded files
rm -rf downloads/
```

## Directory Structure After Setup

```
.
├── docker-compose.yml
├── config.docker.prowlarr.yaml
├── config.docker.jackett.yaml
├── downloads/              # Downloaded torrents
└── yoink                  # Built executable
```
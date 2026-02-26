# Deployment Guide

This document provides instructions for deploying the B2BC Bathhouse Marketplace API to production environments.

## Prerequisites

- Docker and Docker Compose installed
- PostgreSQL 14+ with PostGIS extension
- Redis 6+
- Go 1.25+ (for local development)
- Kubernetes cluster (optional, for K8s deployment)

## Environment Configuration

Before deployment, configure the following environment variables:

### Required Variables

- `BANI_ENVIRONMENT` - Deployment environment (dev/staging/production)
- `BANI_PORT` - Server port (default: 8080)
- `BANI_DATABASE_DSN` - PostgreSQL connection string
- `BANI_JWT_SECRET` - JWT signing secret (minimum 32 characters in production)
- `BANI_REDIS_ADDR` - Redis connection address (host:port)

### Optional Variables

- `BANI_LOGGER_LEVEL` - Logging level (debug/info/warn/error, default: info)
- `BANI_CORS_ALLOWED_ORIGINS` - CORS allowed origins (comma-separated)
- `BANI_CORS_ALLOWED_METHODS` - CORS allowed methods (comma-separated)
- `BANI_CORS_ALLOWED_HEADERS` - CORS allowed headers (comma-separated)

## Docker Deployment

### Building the Docker Image

```bash
docker build -t bani-api:latest .
```

The Dockerfile uses multi-stage builds to minimize the final image size:
1. **Builder stage**: Compiles the Go application
2. **Final stage**: Runs the application with minimal Alpine Linux

### Running with Docker Compose

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  db:
    image: postgis/postgis:14-3.3-alpine
    environment:
      POSTGRES_DB: bani_db
      POSTGRES_USER: bani_user
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bani_user -d bani_db"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  api:
    build: .
    environment:
      BANI_ENVIRONMENT: production
      BANI_PORT: 8080
      BANI_DATABASE_DSN: postgres://bani_user:${DB_PASSWORD}@db:5432/bani_db?sslmode=require
      BANI_REDIS_ADDR: redis:6379
      BANI_JWT_SECRET: ${JWT_SECRET}
      BANI_LOGGER_LEVEL: info
    ports:
      - "8080:8080"
    depends_on:
      db:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 10s

volumes:
  postgres_data:
```

Deploy with Docker Compose:

```bash
docker-compose up -d
```

## Kubernetes Deployment

### ConfigMap and Secrets

Create configuration files for Kubernetes:

```bash
kubectl create configmap bani-config \
  --from-literal=environment=production \
  --from-literal=logger-level=info

kubectl create secret generic bani-secrets \
  --from-literal=db-password=<password> \
  --from-literal=jwt-secret=<secret> \
  --from-literal=redis-password=<password>
```

### Deployment Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bani-api
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: bani-api
  template:
    metadata:
      labels:
        app: bani-api
    spec:
      containers:
      - name: api
        image: bani-api:latest
        imagePullPolicy: IfNotPresent
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: BANI_ENVIRONMENT
          value: "production"
        - name: BANI_PORT
          value: "8080"
        - name: BANI_DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: bani-secrets
              key: db-dsn
        - name: BANI_REDIS_ADDR
          value: "redis-cluster:6379"
        - name: BANI_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: bani-secrets
              key: jwt-secret
        - name: BANI_LOGGER_LEVEL
          value: "info"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: bani-api-service
spec:
  selector:
    app: bani-api
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: LoadBalancer
```

Apply the deployment:

```bash
kubectl apply -f deployment.yaml
```

## Health Checks

The application provides two health check endpoints for container orchestration:

### Liveness Probe (`/health`)

Returns `200 OK` if the application process is running. Used by container orchestrators to determine if the container should be restarted.

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "success": true,
  "status": "ok"
}
```

### Readiness Probe (`/ready`)

Returns `200 OK` if all dependencies (database, Redis) are accessible. Used by load balancers to determine if the container should receive traffic.

```bash
curl http://localhost:8080/ready
```

Response (when ready):
```json
{
  "success": true,
  "status": "ready",
  "database": "connected",
  "redis": "connected"
}
```

Response (when not ready):
```json
{
  "success": false,
  "status": "not_ready",
  "database": "disconnected",
  "redis": "connected"
}
```

## Database Migrations

Migrations are run automatically during application startup. The migrations directory is included in the Docker image.

To manually run migrations before deployment:

```bash
docker run --rm \
  -e BANI_DATABASE_DSN="postgres://user:pass@db:5432/bani_db" \
  bani-api:latest \
  migrate -path /app/migrations -database $BANI_DATABASE_DSN up
```

## Security Considerations

1. **Environment Variables**: Store sensitive data in secure vaults (e.g., Kubernetes Secrets, AWS Secrets Manager)
2. **JWT Secret**: Minimum 32 characters in production, generated with cryptographically secure random source
3. **Database Connection**: Use `sslmode=require` for production PostgreSQL connections
4. **Non-root User**: The Docker container runs as non-root user (app:app) for security
5. **CORS**: Explicitly configure allowed origins for your frontend domains
6. **Logging Level**: Use `info` or `warn` in production, avoid `debug` for performance

## Monitoring and Logging

The application uses structured logging with the following levels:
- `debug`: Detailed diagnostic information
- `info`: General informational messages
- `warn`: Warning messages for potentially problematic conditions
- `error`: Error messages for serious conditions

Configure log aggregation tools (e.g., ELK Stack, Splunk) to collect logs from containers.

## Scaling

For horizontal scaling:
1. Use a load balancer to distribute traffic
2. Ensure all instances share the same database and Redis
3. Set up proper connection pooling in database configuration
4. Monitor CPU and memory usage to determine scaling thresholds

## Rollback Procedure

If deployment fails:

```bash
# With Docker Compose
docker-compose down
docker-compose up -d  # with previous version image

# With Kubernetes
kubectl rollout undo deployment/bani-api
```

## Troubleshooting

### Container fails to start
1. Check logs: `docker logs <container_id>`
2. Verify all environment variables are set
3. Ensure database is accessible

### Health checks failing
1. Check readiness probe: `curl http://localhost:8080/ready`
2. Verify database connection string
3. Verify Redis connectivity

### High memory usage
1. Check logger level (debug level may produce excessive logs)
2. Monitor database query performance
3. Check Redis memory usage

### Database connection timeouts
1. Verify network connectivity
2. Check PostgreSQL is running and accessible
3. Review database query timeout settings (default: 30 seconds)

## Additional Resources

- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Docker Best Practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [PostgreSQL with PostGIS](https://postgis.net/documentation/)

# SHBucket - Self-Hosted S3 Clone

A powerful, self-hosted S3-compatible object storage solution with a beautiful web interface, built with Go and React. Features customizable authentication, distributed storage, and complete UI-based configuration management.

## 🚀 Quick Start - How to See the Web View

The fastest way to get SHBucket running and access the web interface:

### 1. Clone and Start with Docker Compose

```bash
git clone <your-repo-url>
cd SHBucket
docker-compose up -d
```

### 2. Access the Web Interface

Once the containers are running, open your browser and navigate to:

**🌐 Web Interface: http://localhost:3000**

- **Default Admin Login:**
  - Username: `admin@shbucket.local` 
  - Password: `admin123`

- **API Endpoint:** http://localhost:8080/api/v1
- **Health Check:** http://localhost:8080/health

## 📋 Table of Contents

- [Prerequisites](#prerequisites)
- [Installation Methods](#installation-methods)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Documentation](#api-documentation)
- [Development](#development)
- [Troubleshooting](#troubleshooting)

## 🔧 Prerequisites

- **Docker & Docker Compose** (recommended)
- **Go 1.21+** (for manual installation)
- **Node.js 18+** (for manual installation)
- **PostgreSQL 13+** (included in Docker setup)

## 🐳 Installation Methods

### Method 1: Docker Compose (Recommended)

This is the easiest way to run SHBucket with all dependencies included.

#### 1. Create Docker Compose File

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: shbucket
      POSTGRES_USER: shbucket
      POSTGRES_PASSWORD: shbucket_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U shbucket"]
      interval: 30s
      timeout: 10s
      retries: 3

  shbucket-api:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - DATABASE_URL=postgres://shbucket:shbucket_password@postgres:5432/shbucket?sslmode=disable
      - REDIS_URL=redis://redis:6379
      - STORAGE_PATH=/app/storage
      - CONFIG_PATH=/app/config
      - SSL_CERT_PATH=/app/certs
      - JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
      - ADMIN_EMAIL=admin@shbucket.local
      - ADMIN_PASSWORD=admin123
    volumes:
      - storage_data:/app/storage
      - config_data:/app/config
      - cert_data:/app/certs
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  shbucket-web:
    build:
      context: ./web
      dockerfile: Dockerfile
    environment:
      - REACT_APP_API_URL=http://localhost:8080/api/v1
    ports:
      - "3000:3000"
    depends_on:
      - shbucket-api

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - cert_data:/etc/nginx/certs:ro
      - config_data:/etc/nginx/conf.d:ro
    depends_on:
      - shbucket-api
      - shbucket-web

volumes:
  postgres_data:
  storage_data:
  config_data:
  cert_data:
  redis_data:
```

#### 2. Create Application Dockerfile

Create `Dockerfile` in the root directory:

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shbucket ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl nginx apache2-utils
WORKDIR /app

# Create directories
RUN mkdir -p /app/storage /app/config /app/certs /app/logs

# Copy binary
COPY --from=builder /app/shbucket .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Set permissions
RUN chmod +x shbucket

# Create non-root user
RUN addgroup -g 1001 -S shbucket && \
    adduser -S shbucket -u 1001 -G shbucket && \
    chown -R shbucket:shbucket /app

USER shbucket

EXPOSE 8080

CMD ["./shbucket"]
```

#### 3. Create Web Dockerfile

Create `web/Dockerfile`:

```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package files
COPY package*.json ./
RUN npm ci

# Copy source and build
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built app
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx config
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 3000

CMD ["nginx", "-g", "daemon off;"]
```

#### 4. Create Nginx Configuration

Create `web/nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    server {
        listen 3000;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html;
        
        # Handle React Router
        location / {
            try_files $uri $uri/ /index.html;
        }
        
        # API proxy
        location /api/ {
            proxy_pass http://shbucket-api:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

#### 5. Start the Application

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Check status
docker-compose ps
```

### Method 2: Development Setup

For development with database administration tools, create `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: shbucket
      POSTGRES_USER: shbucket
      POSTGRES_PASSWORD: shbucket_password
    volumes:
      - postgres_dev_data:/var/lib/postgresql/data
    ports:
      - "5433:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U shbucket"]
      interval: 30s
      timeout: 10s
      retries: 3

  redis:
    image: redis:7-alpine
    ports:
      - "6380:6379"
    volumes:
      - redis_dev_data:/data

  adminer:
    image: adminer
    ports:
      - "8081:8080"
    depends_on:
      - postgres

volumes:
  postgres_dev_data:
  redis_dev_data:
```

Start development environment:

```bash
# Start development services
docker-compose -f docker-compose.dev.yml up -d

# Run the application locally
export DATABASE_URL="postgres://shbucket:shbucket_password@localhost:5433/shbucket?sslmode=disable"
export REDIS_URL="redis://localhost:6380"
go run ./cmd/server

# In another terminal, start the web development server
cd web
npm start
```

### Method 3: Manual Installation

#### 1. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Node.js dependencies
cd web
npm install
```

#### 2. Setup PostgreSQL

```bash
# Install PostgreSQL (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib

# Create database and user
sudo -u postgres psql
CREATE DATABASE shbucket;
CREATE USER shbucket WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE shbucket TO shbucket;
\q
```

#### 3. Run Database Migrations

```bash
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path ./migrations -database "postgres://shbucket:your_password@localhost:5432/shbucket?sslmode=disable" up
```

#### 4. Build and Run

```bash
# Build the API
go build -o bin/shbucket ./cmd/server

# Build the web app
cd web
npm run build
cd ..

# Run the API
export DATABASE_URL="postgres://shbucket:your_password@localhost:5432/shbucket?sslmode=disable"
export STORAGE_PATH="./storage"
export CONFIG_PATH="./config"
./bin/shbucket

# In another terminal, serve the web app (development)
cd web
npm start
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file:

```env
# Database
DATABASE_URL=postgres://shbucket:shbucket_password@localhost:5432/shbucket?sslmode=disable

# Server
PORT=8080
HOST=0.0.0.0

# Storage
STORAGE_PATH=/app/storage
CONFIG_PATH=/app/config
SSL_CERT_PATH=/app/certs

# Security
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
BCRYPT_COST=12

# Admin User (created on first run)
ADMIN_EMAIL=admin@shbucket.local
ADMIN_PASSWORD=admin123

# Redis (optional, for caching)
REDIS_URL=redis://localhost:6379

# Features
ENABLE_METRICS=true
ENABLE_CORS=true
LOG_LEVEL=info

# Web UI
REACT_APP_API_URL=http://localhost:8080/api/v1
```

### PostgreSQL Configuration

The application is configured to use PostgreSQL by default. The database schema is automatically created using migrations.

**Connection String Format:**
```
postgres://username:password@host:port/database?sslmode=disable
```

**Required PostgreSQL version:** 13+

### Directory Structure

```
SHBucket/
├── storage/          # File storage directory
├── config/           # Generated web server configs
├── certs/            # SSL certificates
├── logs/             # Application logs
└── migrations/       # Database migrations
```

## 🎯 Usage

### Accessing the Web Interface

1. **Open your browser** and navigate to `http://localhost:3000`
2. **Login** with the default admin credentials:
   - Email: `admin@shbucket.local`
   - Password: `admin123`

### Main Features

#### 1. **Dashboard**
- Overview of buckets, files, and storage usage
- System health and metrics

#### 2. **Bucket Management**
- Create, update, and delete buckets
- Configure bucket-level authentication
- Set public/private access

#### 3. **File Management**
- Upload files (single and multipart)
- Download and preview files
- File-level authentication overrides

#### 4. **System Configuration**
- **System Settings**: App name, timezone, upload limits
- **Web Server Config**: Nginx/Apache configuration through UI
- **SSL Certificates**: Automatic Let's Encrypt and custom certificates
- **Security Settings**: Authentication, rate limiting, CORS

#### 5. **Node Management**
- Configure distributed storage nodes
- Monitor node health and capacity

### API Usage

#### Authentication

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@shbucket.local","password":"admin123"}'
```

#### Bucket Operations

```bash
# Create bucket
curl -X POST http://localhost:8080/api/v1/buckets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"my-bucket","description":"My test bucket"}'

# List buckets
curl -X GET http://localhost:8080/api/v1/buckets \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### File Operations

```bash
# Upload file
curl -X POST http://localhost:8080/api/v1/buckets/my-bucket/files \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "file=@/path/to/your/file.jpg"

# Download file
curl -X GET http://localhost:8080/api/v1/buckets/my-bucket/files/file.jpg \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -o downloaded-file.jpg
```

## 📚 API Documentation

Once running, API documentation is available at:
- **Swagger UI**: http://localhost:8080/swagger/
- **OpenAPI Spec**: http://localhost:8080/docs/swagger.yaml

## 🛠️ Development

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 13+
- Docker & Docker Compose

### Quick Development Setup

#### Option 1: All-in-One Development
```bash
# Clone repository
git clone <your-repo-url>
cd SHBucket

# Start everything with one command
./scripts/dev-full.sh
```

#### Option 2: Run Components Independently

**Database services only:**
```bash
./scripts/dev-db-only.sh
```

**Backend API only:**
```bash
./scripts/dev-backend.sh
```

**Frontend Web UI only:**
```bash
./scripts/dev-frontend.sh
```

See **[DEVELOPMENT.md](DEVELOPMENT.md)** for detailed independent development setup.

### Development URLs

- **API**: http://localhost:8080
- **Web App**: http://localhost:3000
- **PostgreSQL**: localhost:5433 (dev) / localhost:5432 (prod)
- **Redis**: localhost:6380 (dev) / localhost:6379 (prod)
- **Adminer** (dev only): http://localhost:8081

### Hot Reload Support

Install `air` for Go hot reload:
```bash
go install github.com/cosmtrek/air@latest
```

React hot reload is enabled by default with `npm start`.

### Running Tests

```bash
# Run Go tests
go test ./...

# Run web tests
cd web
npm test
```

### Development Guides

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Complete guide for running projects independently
- **[scripts/README.md](scripts/README.md)** - Development scripts documentation
- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Quick start guide

## 🔍 Troubleshooting

### Common Issues

#### 1. **Cannot connect to database**
```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Check connection
psql "postgres://shbucket:shbucket_password@localhost:5432/shbucket"
```

#### 2. **Web interface not loading**
```bash
# Check if services are running
docker-compose ps

# Check logs
docker-compose logs shbucket-web
docker-compose logs shbucket-api
```

#### 3. **Permission denied errors**
```bash
# Fix storage permissions
sudo chown -R $(whoami):$(whoami) ./storage ./config ./certs
```

#### 4. **SSL certificate issues**
```bash
# Check certificate status in the web UI
# Navigate to System Settings > SSL Certificates

# Or check via API
curl -X GET http://localhost:8080/api/v1/config/ssl \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Health Checks

```bash
# Check API health
curl http://localhost:8080/health

# Check database connection
curl http://localhost:8080/api/v1/health/db

# Check all services
docker-compose ps
```

### Logs

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f shbucket-api
docker-compose logs -f postgres
```

## 🔐 Security Notes

1. **Change default passwords** in production
2. **Use HTTPS** with proper SSL certificates
3. **Configure firewall** to restrict access
4. **Regular backups** of database and storage
5. **Monitor logs** for suspicious activity

## 🚀 Features

- 🔐 **Flexible Authentication**: Multiple auth types per bucket (signed URLs, API keys, JWT, sessions)
- 📁 **Bucket Management**: Create buckets with custom rules and settings
- 🔒 **File-Level Security**: Override bucket auth rules on individual files
- 🌐 **Public URLs**: Secure access to files through signed URLs
- 🔄 **Distributed Storage**: Scale across multiple nodes when storage is full
- ⚡ **High Performance**: Built with Go and Fiber framework
- 🏗️ **CQRS Architecture**: Clean separation of commands and queries
- 🎨 **Beautiful Web UI**: Dark-themed React interface for complete management
- ⚙️ **UI-Based Configuration**: Manage nginx/apache configs through web interface
- 🔒 **SSL Management**: Automatic Let's Encrypt certificate management

## 📝 License

MIT License - see LICENSE file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## 📞 Support

- **Issues**: [GitHub Issues](your-repo-url/issues)
- **Documentation**: [Wiki](your-repo-url/wiki)
- **Discussions**: [GitHub Discussions](your-repo-url/discussions)
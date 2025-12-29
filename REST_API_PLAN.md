# REST API Plan for ligneous-gedcom

**Date:** 2025-01-27  
**Target Domain:** ligneous.org  
**Purpose:** Expose gedcom-go functionality via REST API for web applications and integrations

---

## 1. Architecture Overview

### 1.1 High-Level Architecture

```
Internet
    ↓
nginx (ligneous.org) - SSL termination, reverse proxy
    ↓
Go REST API Server (port 8080)
    ↓
gedcom-go library (in-memory processing)
    ↓
Optional: Redis (caching) / PostgreSQL (metadata storage)
```

### 1.2 Technology Stack

- **API Server:** Go (using `net/http` or `gin`/`chi` router)
- **Reverse Proxy:** nginx (SSL, rate limiting, static file serving)
- **Caching:** Redis (optional, for parsed graph caching)
- **Database:** PostgreSQL (optional, for user data, file metadata, API keys)
- **File Storage:** Local filesystem or S3-compatible storage
- **Authentication:** JWT tokens or API keys

---

## 2. API Design Principles

### 2.1 RESTful Design

- **Resource-based URLs:** `/api/v1/individuals/{xref}`
- **HTTP methods:** GET (read), POST (create), PUT (update), DELETE (remove)
- **Status codes:** 200 (OK), 201 (Created), 400 (Bad Request), 404 (Not Found), 500 (Error)
- **JSON responses:** Consistent structure with `data`, `error`, `meta` fields

### 2.2 API Versioning

- **URL-based:** `/api/v1/...`, `/api/v2/...`
- **Header-based:** `Accept: application/vnd.ligneous.v1+json` (optional)

### 2.3 Response Format

```json
{
  "data": { ... },
  "meta": {
    "request_id": "uuid",
    "timestamp": "2025-01-27T10:00:00Z",
    "version": "v1"
  },
  "error": null
}
```

Error Response:
```json
{
  "data": null,
  "error": {
    "code": "INVALID_XREF",
    "message": "Individual @I999@ not found",
    "details": { ... }
  },
  "meta": { ... }
}
```

---

## 3. Core Endpoints

### 3.1 File Management

#### Upload & Parse GEDCOM File
```
POST /api/v1/files
Content-Type: multipart/form-data

Request:
- file: GEDCOM file
- name: (optional) friendly name
- description: (optional) description

Response:
{
  "data": {
    "file_id": "uuid",
    "name": "family.ged",
    "size": 1048576,
    "individuals_count": 1234,
    "families_count": 567,
    "parse_duration_ms": 234,
    "status": "parsed",
    "created_at": "2025-01-27T10:00:00Z"
  }
}
```

#### Get File Info
```
GET /api/v1/files/{file_id}

Response:
{
  "data": {
    "file_id": "uuid",
    "name": "family.ged",
    "size": 1048576,
    "individuals_count": 1234,
    "families_count": 567,
    "parse_errors": 0,
    "parse_warnings": 12,
    "created_at": "2025-01-27T10:00:00Z",
    "graph_built": true
  }
}
```

#### List Files
```
GET /api/v1/files?limit=50&offset=0

Response:
{
  "data": [
    { "file_id": "...", "name": "...", ... }
  ],
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0
  }
}
```

#### Delete File
```
DELETE /api/v1/files/{file_id}

Response:
{
  "data": {
    "file_id": "uuid",
    "deleted": true
  }
}
```

---

### 3.2 Individual Queries

#### Get Individual
```
GET /api/v1/files/{file_id}/individuals/{xref}

Response:
{
  "data": {
    "xref": "@I1@",
    "name": "John /Doe/",
    "given_name": "John",
    "surname": "Doe",
    "sex": "M",
    "birth_date": "1 JAN 1900",
    "birth_place": "New York, USA",
    "death_date": "15 MAR 1980",
    "death_place": "Los Angeles, USA",
    "notes": [...],
    "sources": [...]
  }
}
```

#### List Individuals
```
GET /api/v1/files/{file_id}/individuals?limit=100&offset=0&name=John

Query Parameters:
- limit: max results (default: 100, max: 1000)
- offset: pagination offset
- name: filter by name (substring match)
- sex: filter by sex (M/F/U)
- birth_year: filter by birth year
- birth_place: filter by birth place
- has_children: boolean
- has_spouse: boolean
- living: boolean

Response:
{
  "data": [
    { "xref": "@I1@", "name": "...", ... }
  ],
  "meta": {
    "total": 1234,
    "limit": 100,
    "offset": 0
  }
}
```

#### Search Individuals (Advanced Filtering)
```
POST /api/v1/files/{file_id}/individuals/search

Request:
{
  "filters": {
    "name": "John",
    "sex": "M",
    "birth_date_range": {
      "start": "1800-01-01",
      "end": "1900-12-31"
    },
    "birth_place": "New York",
    "has_children": true,
    "living": false
  },
  "limit": 100,
  "offset": 0
}

Response:
{
  "data": [...],
  "meta": {
    "total": 45,
    "limit": 100,
    "offset": 0
  }
}
```

---

### 3.3 Relationship Queries

#### Get Parents
```
GET /api/v1/files/{file_id}/individuals/{xref}/parents

Response:
{
  "data": [
    { "xref": "@I2@", "name": "Father", ... },
    { "xref": "@I3@", "name": "Mother", ... }
  ]
}
```

#### Get Children
```
GET /api/v1/files/{file_id}/individuals/{xref}/children

Response:
{
  "data": [
    { "xref": "@I4@", "name": "Child 1", ... },
    { "xref": "@I5@", "name": "Child 2", ... }
  ]
}
```

#### Get Siblings
```
GET /api/v1/files/{file_id}/individuals/{xref}/siblings

Response:
{
  "data": [...]
}
```

#### Get Spouses
```
GET /api/v1/files/{file_id}/individuals/{xref}/spouses

Response:
{
  "data": [...]
}
```

#### Get Extended Relationships
```
GET /api/v1/files/{file_id}/individuals/{xref}/grandparents
GET /api/v1/files/{file_id}/individuals/{xref}/grandchildren
GET /api/v1/files/{file_id}/individuals/{xref}/uncles
GET /api/v1/files/{file_id}/individuals/{xref}/aunts
GET /api/v1/files/{file_id}/individuals/{xref}/cousins?degree=1
GET /api/v1/files/{file_id}/individuals/{xref}/nephews
GET /api/v1/files/{file_id}/individuals/{xref}/nieces
```

#### Get Ancestors
```
GET /api/v1/files/{file_id}/individuals/{xref}/ancestors?max_generations=5&include_self=false

Query Parameters:
- max_generations: limit depth (default: unlimited)
- include_self: include starting individual (default: false)

Response:
{
  "data": [
    { "xref": "@I2@", "name": "...", "generation": 1, ... },
    { "xref": "@I3@", "name": "...", "generation": 1, ... },
    { "xref": "@I4@", "name": "...", "generation": 2, ... }
  ],
  "meta": {
    "total": 15,
    "max_generations": 5,
    "generations_found": 3
  }
}
```

#### Get Descendants
```
GET /api/v1/files/{file_id}/individuals/{xref}/descendants?max_generations=3&include_self=false

Response: (similar to ancestors)
```

#### Calculate Relationship
```
GET /api/v1/files/{file_id}/individuals/{xref1}/relationship/{xref2}

Response:
{
  "data": {
    "from": { "xref": "@I1@", "name": "..." },
    "to": { "xref": "@I2@", "name": "..." },
    "relationship_type": "cousin",
    "degree": 1,
    "removal": 0,
    "is_direct": false,
    "is_collateral": true,
    "path": {
      "length": 4,
      "nodes": ["@I1@", "@I3@", "@I4@", "@I2@"]
    }
  }
}
```

#### Find Paths
```
GET /api/v1/files/{file_id}/individuals/{xref1}/paths/{xref2}?max_length=10&include_blood=true&include_marital=false

Query Parameters:
- max_length: maximum path length (default: 10)
- include_blood: include blood relations (default: true)
- include_marital: include marital relations (default: true)
- all: return all paths or just shortest (default: false = shortest only)

Response:
{
  "data": {
    "shortest_path": {
      "length": 4,
      "type": "blood",
      "nodes": ["@I1@", "@I3@", "@I4@", "@I2@"]
    },
    "all_paths": [
      { "length": 4, "type": "blood", "nodes": [...] },
      { "length": 5, "type": "mixed", "nodes": [...] }
    ]
  }
}
```

#### Common Ancestors
```
GET /api/v1/files/{file_id}/individuals/{xref1}/common-ancestors/{xref2}

Response:
{
  "data": [
    { "xref": "@I5@", "name": "...", ... }
  ]
}
```

---

### 3.4 Family Queries

#### Get Family
```
GET /api/v1/files/{file_id}/families/{xref}

Response:
{
  "data": {
    "xref": "@F1@",
    "husband": { "xref": "@I1@", "name": "..." },
    "wife": { "xref": "@I2@", "name": "..." },
    "children": [
      { "xref": "@I3@", "name": "..." }
    ],
    "marriage_date": "1 JAN 1920",
    "marriage_place": "New York, USA",
    "events": [...]
  }
}
```

#### List Families
```
GET /api/v1/files/{file_id}/families?limit=100&offset=0

Response:
{
  "data": [...],
  "meta": { ... }
}
```

---

### 3.5 Graph Analytics

#### Graph Metrics
```
GET /api/v1/files/{file_id}/metrics

Response:
{
  "data": {
    "individuals_count": 1234,
    "families_count": 567,
    "graph_diameter": 12,
    "average_path_length": 4.5,
    "graph_density": 0.003,
    "average_degree": 3.2,
    "connected_components": 1,
    "longest_path_length": 15
  }
}
```

#### Centrality Measures
```
GET /api/v1/files/{file_id}/centrality?type=degree

Query Parameters:
- type: degree, betweenness, closeness (default: degree)

Response:
{
  "data": {
    "@I1@": 15.0,
    "@I2@": 12.0,
    ...
  }
}
```

#### Most Connected Individuals
```
GET /api/v1/files/{file_id}/most-connected?limit=10&type=degree

Response:
{
  "data": [
    { "xref": "@I1@", "name": "...", "centrality": 15.0 },
    { "xref": "@I2@", "name": "...", "centrality": 12.0 }
  ]
}
```

---

### 3.6 Duplicate Detection

#### Find Duplicates
```
POST /api/v1/files/{file_id}/duplicates

Request:
{
  "min_threshold": 0.60,
  "high_confidence_threshold": 0.85,
  "use_phonetic_matching": true,
  "use_relationship_data": true,
  "max_results": 200
}

Response:
{
  "data": {
    "matches": [
      {
        "individual1": { "xref": "@I1@", "name": "..." },
        "individual2": { "xref": "@I2@", "name": "..." },
        "similarity_score": 0.92,
        "confidence": "high",
        "matching_fields": ["name", "birth_date", "birth_place"],
        "differences": ["death_date"],
        "breakdown": {
          "name_score": 0.95,
          "date_score": 0.90,
          "place_score": 0.88,
          "sex_score": 1.0,
          "relationship_score": 0.85
        }
      }
    ],
    "total_comparisons": 12345,
    "processing_time_ms": 1234
  }
}
```

#### Find Duplicates Between Two Files
```
POST /api/v1/duplicates/compare

Request:
{
  "file1_id": "uuid1",
  "file2_id": "uuid2",
  "min_threshold": 0.70,
  ...
}

Response: (similar to above)
```

---

### 3.7 Validation

#### Validate File
```
POST /api/v1/files/{file_id}/validate

Request:
{
  "severity": "warning",  // error, warning, info
  "include_suggestions": true
}

Response:
{
  "data": {
    "valid": false,
    "errors": [
      {
        "type": "MISSING_BIRTH_DATE",
        "severity": "warning",
        "message": "Individual @I1@ is missing birth date",
        "xref": "@I1@",
        "suggestion": "Consider adding birth date if known"
      }
    ],
    "summary": {
      "total_errors": 5,
      "severe_errors": 2,
      "warnings": 3,
      "info": 0
    }
  }
}
```

#### Data Quality Report
```
GET /api/v1/files/{file_id}/quality

Response:
{
  "data": {
    "completeness": {
      "individuals_with_birth_date": 0.85,
      "individuals_with_death_date": 0.60,
      "individuals_with_place": 0.75
    },
    "consistency": {
      "date_conflicts": 12,
      "relationship_issues": 3
    },
    "recommendations": [
      "Consider adding birth dates for 15% of individuals",
      "Check date consistency for 12 individuals"
    ]
  }
}
```

---

### 3.8 Diff/Comparison

#### Compare Two Files
```
POST /api/v1/diff

Request:
{
  "file1_id": "uuid1",
  "file2_id": "uuid2",
  "matching_strategy": "hybrid",  // xref, content, hybrid
  "similarity_threshold": 0.85,
  "date_tolerance": 2,
  "detail_level": "field",  // summary, field, full
  "output_format": "json"  // json, text, html
}

Response:
{
  "data": {
    "summary": {
      "file1_stats": {
        "individuals": 1234,
        "families": 567
      },
      "file2_stats": {
        "individuals": 1250,
        "families": 570
      },
      "changes": {
        "added": 20,
        "removed": 4,
        "modified": 12
      }
    },
    "changes": {
      "added": [
        { "xref": "@I1000@", "type": "INDI", ... }
      ],
      "removed": [...],
      "modified": [
        {
          "xref": "@I1@",
          "type": "INDI",
          "changes": [
            {
              "field": "NAME",
              "old_value": "John /Doe/",
              "new_value": "John /Smith/",
              "type": "modified"
            }
          ]
        }
      ]
    },
    "statistics": {
      "processing_time_ms": 234,
      "records_compared": 1234
    }
  }
}
```

---

### 3.9 Export

#### Export File
```
POST /api/v1/files/{file_id}/export

Request:
{
  "format": "json",  // json, xml, yaml, csv, gedcom
  "include_individuals": true,
  "include_families": true,
  "include_notes": true,
  "include_sources": true,
  "filters": {
    "individuals": ["@I1@", "@I2@"],
    "families": ["@F1@"]
  }
}

Response:
{
  "data": {
    "export_id": "uuid",
    "format": "json",
    "size": 1048576,
    "download_url": "/api/v1/exports/{export_id}/download",
    "expires_at": "2025-01-28T10:00:00Z"
  }
}
```

#### Download Export
```
GET /api/v1/exports/{export_id}/download

Response: (file download with appropriate Content-Type)
```

#### Export Subtree (Ancestors/Descendants)
```
POST /api/v1/files/{file_id}/export/subtree

Request:
{
  "root_xref": "@I1@",
  "direction": "ancestors",  // ancestors, descendants, both
  "max_generations": 5,
  "format": "gedcom",
  "include_self": true
}

Response: (similar to export)
```

---

## 4. Authentication & Authorization

### 4.1 Authentication Methods

#### Option 1: API Keys (Simple)
```
Header: X-API-Key: your-api-key-here
```

#### Option 2: JWT Tokens (Recommended)
```
Header: Authorization: Bearer <jwt-token>
```

### 4.2 User Management (Optional)

```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/refresh
GET  /api/v1/auth/me
```

### 4.3 Rate Limiting

- **Free tier:** 100 requests/hour
- **Paid tier:** 1000 requests/hour
- **Enterprise:** Custom limits

Headers:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1643200000
```

---

## 5. Caching Strategy

### 5.1 Cache Layers

1. **Graph Cache (Redis):**
   - Cache parsed graph for each file_id
   - TTL: 1 hour (or until file updated)
   - Key: `graph:{file_id}`

2. **Query Result Cache (Redis):**
   - Cache expensive queries (ancestors, descendants, duplicates)
   - TTL: 15 minutes
   - Key: `query:{file_id}:{query_hash}`

3. **File Metadata Cache (Redis):**
   - Cache file stats, individual counts, etc.
   - TTL: 5 minutes
   - Key: `meta:{file_id}`

### 5.2 Cache Invalidation

- File upload/update → invalidate all caches for that file
- Manual cache clear endpoint: `DELETE /api/v1/files/{file_id}/cache`

---

## 6. File Storage

### 6.1 Storage Options

#### Option 1: Local Filesystem
- **Path:** `/var/lib/ligneous/files/{file_id}/`
- **Pros:** Simple, fast
- **Cons:** Not scalable, backup required

#### Option 2: S3-Compatible Storage
- **Provider:** AWS S3, MinIO, DigitalOcean Spaces
- **Path:** `s3://ligneous-bucket/files/{file_id}/`
- **Pros:** Scalable, durable, backup built-in
- **Cons:** Slightly slower, requires setup

### 6.2 File Organization

```
files/
  {file_id}/
    original.ged          # Original uploaded file
    parsed.json           # Parsed tree (optional, for faster loading)
    graph.cache           # Serialized graph (optional)
    exports/
      {export_id}.{ext}   # Generated exports
```

---

## 7. Database Schema (Optional)

### 7.1 PostgreSQL Tables

```sql
-- Users (if implementing user management)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- API Keys
CREATE TABLE api_keys (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP
);

-- Files
CREATE TABLE files (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255),
    size BIGINT NOT NULL,
    individuals_count INT,
    families_count INT,
    parse_status VARCHAR(50),  -- pending, parsed, error
    parse_errors INT DEFAULT 0,
    parse_warnings INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- File Metadata (for search/indexing)
CREATE TABLE file_metadata (
    file_id UUID REFERENCES files(id),
    key VARCHAR(255),
    value TEXT,
    PRIMARY KEY (file_id, key)
);

-- Exports
CREATE TABLE exports (
    id UUID PRIMARY KEY,
    file_id UUID REFERENCES files(id),
    format VARCHAR(50) NOT NULL,
    size BIGINT,
    status VARCHAR(50),  -- pending, completed, failed
    download_url VARCHAR(500),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Rate Limiting
CREATE TABLE rate_limits (
    api_key_id UUID REFERENCES api_keys(id),
    endpoint VARCHAR(255),
    count INT DEFAULT 0,
    window_start TIMESTAMP,
    PRIMARY KEY (api_key_id, endpoint, window_start)
);
```

---

## 8. Deployment Architecture

### 8.1 Server Setup

#### nginx Configuration (`/etc/nginx/sites-available/ligneous.org`)

```nginx
server {
    listen 80;
    server_name ligneous.org www.ligneous.org;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name ligneous.org www.ligneous.org;

    # SSL certificates (Let's Encrypt)
    ssl_certificate /etc/letsencrypt/live/ligneous.org/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/ligneous.org/privkey.pem;

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api_limit:10m rate=10r/s;
    limit_req zone=api_limit burst=20 nodelay;

    # File upload size limit
    client_max_body_size 100M;

    # API endpoints
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Timeouts for long-running operations
        proxy_read_timeout 300s;
        proxy_connect_timeout 10s;
    }

    # Static files (if serving frontend)
    location / {
        root /var/www/ligneous.org;
        try_files $uri $uri/ /index.html;
    }

    # Health check
    location /health {
        access_log off;
        proxy_pass http://localhost:8080/health;
    }
}
```

### 8.2 Go Server Setup

#### Systemd Service (`/etc/systemd/system/ligneous-api.service`)

```ini
[Unit]
Description=Ligneous GEDCOM API Server
After=network.target

[Service]
Type=simple
User=ligneous
WorkingDirectory=/opt/ligneous-api
ExecStart=/opt/ligneous-api/ligneous-api
Restart=always
RestartSec=5
Environment="PORT=8080"
Environment="REDIS_URL=redis://localhost:6379"
Environment="DATABASE_URL=postgres://user:pass@localhost/ligneous"
Environment="STORAGE_PATH=/var/lib/ligneous/files"
Environment="JWT_SECRET=your-secret-key"

[Install]
WantedBy=multi-user.target
```

### 8.3 Deployment Steps

1. **Server Setup:**
   ```bash
   # Install Go, nginx, PostgreSQL, Redis
   sudo apt update
   sudo apt install nginx postgresql redis-server
   
   # Create user and directories
   sudo useradd -r -s /bin/false ligneous
   sudo mkdir -p /var/lib/ligneous/files
   sudo mkdir -p /opt/ligneous-api
   sudo chown ligneous:ligneous /var/lib/ligneous
   ```

2. **SSL Certificate (Let's Encrypt):**
   ```bash
   sudo apt install certbot python3-certbot-nginx
   sudo certbot --nginx -d ligneous.org -d www.ligneous.org
   ```

3. **Build & Deploy:**
   ```bash
   # Build binary
   cd /apps/gedcom-go
   go build -o ligneous-api cmd/api/main.go
   
   # Copy to server
   scp ligneous-api user@ligneous.org:/opt/ligneous-api/
   
   # Start service
   sudo systemctl enable ligneous-api
   sudo systemctl start ligneous-api
   ```

4. **Monitor:**
   ```bash
   sudo systemctl status ligneous-api
   sudo journalctl -u ligneous-api -f
   ```

---

## 9. API Endpoints Summary

### 9.1 Complete Endpoint List

| Method | Endpoint | Description |
|--------|----------|-------------|
| **File Management** |
| POST | `/api/v1/files` | Upload & parse GEDCOM file |
| GET | `/api/v1/files` | List files |
| GET | `/api/v1/files/{file_id}` | Get file info |
| DELETE | `/api/v1/files/{file_id}` | Delete file |
| **Individuals** |
| GET | `/api/v1/files/{file_id}/individuals` | List individuals |
| GET | `/api/v1/files/{file_id}/individuals/{xref}` | Get individual |
| POST | `/api/v1/files/{file_id}/individuals/search` | Search individuals |
| **Relationships** |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/parents` | Get parents |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/children` | Get children |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/siblings` | Get siblings |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/spouses` | Get spouses |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/ancestors` | Get ancestors |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/descendants` | Get descendants |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/relationship/{xref2}` | Calculate relationship |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/paths/{xref2}` | Find paths |
| GET | `/api/v1/files/{file_id}/individuals/{xref}/common-ancestors/{xref2}` | Common ancestors |
| **Families** |
| GET | `/api/v1/files/{file_id}/families` | List families |
| GET | `/api/v1/files/{file_id}/families/{xref}` | Get family |
| **Analytics** |
| GET | `/api/v1/files/{file_id}/metrics` | Graph metrics |
| GET | `/api/v1/files/{file_id}/centrality` | Centrality measures |
| GET | `/api/v1/files/{file_id}/most-connected` | Most connected individuals |
| **Duplicate Detection** |
| POST | `/api/v1/files/{file_id}/duplicates` | Find duplicates |
| POST | `/api/v1/duplicates/compare` | Compare two files |
| **Validation** |
| POST | `/api/v1/files/{file_id}/validate` | Validate file |
| GET | `/api/v1/files/{file_id}/quality` | Data quality report |
| **Diff** |
| POST | `/api/v1/diff` | Compare two files |
| **Export** |
| POST | `/api/v1/files/{file_id}/export` | Export file |
| POST | `/api/v1/files/{file_id}/export/subtree` | Export subtree |
| GET | `/api/v1/exports/{export_id}/download` | Download export |
| **System** |
| GET | `/health` | Health check |
| GET | `/api/v1/version` | API version |

---

## 10. Performance Considerations

### 10.1 Optimization Strategies

1. **Graph Caching:**
   - Cache parsed graphs in Redis
   - Invalidate on file update
   - TTL: 1 hour

2. **Query Result Caching:**
   - Cache expensive queries (ancestors, descendants, duplicates)
   - Cache key includes query parameters
   - TTL: 15 minutes

3. **Async Processing:**
   - Large file parsing → background job
   - Duplicate detection → background job
   - Export generation → background job
   - Webhook notifications when complete

4. **Pagination:**
   - All list endpoints support pagination
   - Default limit: 100, max: 1000

5. **Connection Pooling:**
   - Database connection pool
   - Redis connection pool
   - HTTP client connection pool

### 10.2 Rate Limiting

- **Per API key:** 100 requests/hour (free), 1000/hour (paid)
- **Per IP:** 10 requests/second (burst: 20)
- **Per endpoint:** Different limits for expensive operations

---

## 11. Security Considerations

### 11.1 Input Validation

- Validate all XREF formats
- Sanitize file uploads
- Limit file size (100MB default)
- Validate date formats
- Check query parameters (max_generations, limits)

### 11.2 Authentication

- API keys stored as hashes (bcrypt)
- JWT tokens with expiration
- Refresh token rotation
- Rate limiting per API key

### 11.3 File Security

- Files stored with UUID names (not user-provided)
- Access control per file (user_id)
- File size limits
- Virus scanning (optional, ClamAV)

### 11.4 API Security

- HTTPS only (TLS 1.2+)
- CORS configuration
- Request size limits
- Timeout limits for long operations

---

## 12. Monitoring & Logging

### 12.1 Logging

- **Access logs:** nginx access.log
- **Application logs:** Structured JSON logs
- **Error logs:** Separate error log file
- **Query logs:** Log slow queries (>1s)

### 12.2 Metrics

- Request count per endpoint
- Response times (p50, p95, p99)
- Error rates
- Cache hit rates
- File processing times
- Active file count

### 12.3 Health Checks

```
GET /health

Response:
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "storage": "available",
  "version": "1.0.0"
}
```

---

## 13. Implementation Phases

### Phase 1: Core API (MVP)
- File upload & parsing
- Individual queries (get, list, search)
- Basic relationships (parents, children, siblings, spouses)
- Health check

### Phase 2: Advanced Queries
- Ancestors/descendants
- Relationship calculation
- Path finding
- Graph metrics

### Phase 3: Analysis Features
- Duplicate detection
- Validation
- Data quality reports
- Diff/comparison

### Phase 4: Export & Optimization
- Export functionality
- Caching layer
- Background jobs
- Performance optimization

### Phase 5: Production Features
- Authentication & authorization
- Rate limiting
- Monitoring & logging
- Documentation

---

## 14. API Documentation

### 14.1 OpenAPI/Swagger

- Generate OpenAPI 3.0 specification
- Interactive API docs at `/api/docs`
- Postman collection export

### 14.2 Example Requests

- Include curl examples
- Include code examples (Go, Python, JavaScript)
- Postman collection

---

## 15. Next Steps

1. **Create API package structure:**
   ```
   cmd/api/
     main.go
   api/
     handlers/
       files.go
       individuals.go
       relationships.go
       ...
     middleware/
       auth.go
       logging.go
       rate_limit.go
     models/
       request.go
       response.go
     storage/
       file_storage.go
       cache.go
   ```

2. **Set up development environment:**
   - Local nginx config
   - Local Redis instance
   - Local PostgreSQL (optional)
   - Environment variables

3. **Implement Phase 1 endpoints:**
   - Start with file upload
   - Basic individual queries
   - Health check

4. **Test with real GEDCOM files:**
   - Use testdata files
   - Test with various sizes
   - Performance testing

5. **Deploy to staging:**
   - Set up staging server
   - Test deployment process
   - Load testing

---

**Planning Complete** ✅

This plan provides a comprehensive foundation for building the REST API. The architecture is scalable, secure, and follows REST best practices.


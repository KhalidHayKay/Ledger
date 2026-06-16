# 📁 Ledger - Project Structure

_Generated on: 09/06/2026, 20:44:28_

## 📋 Quick Overview

| Metric           | Value    |
| ---------------- | -------- |
| 📄 Total Files   | 48       |
| 📁 Total Folders | 18       |
| 🌳 Max Depth     | 2 levels |
| 🛠️ Tech Stack    | Docker   |

## ⭐ Important Files

- 🟡 🚫 **.gitignore** - Git ignore rules
- 🟡 🐳 **Dockerfile** - Docker container

## 📊 File Statistics

### By File Type

- 📄 **.go** (Other files): 31 files (64.6%)
- ⚙️ **.yml** (YAML files): 3 files (6.3%)
- ⚙️ **.yaml** (YAML files): 3 files (6.3%)
- ⚙️ **.toml** (TOML files): 1 files (2.1%)
- 🐳 **.dockerignore** (Docker ignore): 1 files (2.1%)
- 📄 **.example** (Other files): 1 files (2.1%)
- 🚫 **.gitignore** (Git ignore): 1 files (2.1%)
- 🐳 **.dockerfile** (Docker files): 1 files (2.1%)
- 📄 **.txt** (Text files): 1 files (2.1%)
- 📄 **.mod** (Other files): 1 files (2.1%)
- 📄 **.sum** (Other files): 1 files (2.1%)
- 📄 **.** (Other files): 1 files (2.1%)
- 📖 **.md** (Markdown files): 1 files (2.1%)
- 📄 **.sh** (Other files): 1 files (2.1%)

### By Category

- **Other**: 36 files (75.0%)
- **Config**: 7 files (14.6%)
- **DevOps**: 3 files (6.3%)
- **Docs**: 2 files (4.2%)

### 📁 Largest Directories

- **root**: 48 files
- **internal**: 21 files
- **internal/paymentintent**: 10 files
- **app**: 6 files
- **internal/idempotency**: 5 files

## 🌳 Directory Structure

```
Ledger/
├── ⚙️ .air.toml
├── 🐳 .dockerignore
├── 📄 .env.example
├── 📂 .github/
│   └── 📂 workflows/
│   │   ├── ⚙️ build.yml
│   │   ├── ⚙️ deploy.yaml
│   │   ├── ⚙️ lint.yml
│   │   └── ⚙️ test.yml
├── 🟡 🚫 **.gitignore**
├── 🚀 app/
│   ├── ⚙️ config/
│   │   ├── 📄 env.go
│   │   └── 📄 validate.go
│   ├── 📂 middleware/
│   │   └── 📄 middleware.go
│   ├── 📂 render/
│   │   └── 📄 json.go
│   └── 📂 storage/
│   │   ├── 📄 pgsql.go
│   │   └── 📄 redis.go
├── 📂 cmd/
│   ├── 🚀 app/
│   │   └── 📄 main.go
│   └── 📂 cli/
│   │   └── 📄 main.go
├── ⚙️ compose.override.yaml
├── ⚙️ compose.yaml
├── 📄 coverage.txt
├── 🟡 🐳 **Dockerfile**
├── 📄 go.mod
├── 📄 go.sum
├── 📂 internal/
│   ├── 📂 bank/
│   │   ├── 📄 dto.go
│   │   ├── 📄 ficmart_bank_repo.go
│   │   ├── 📄 model.go
│   │   └── 📄 repository.go
│   ├── 📂 idempotency/
│   │   ├── 📄 model.go
│   │   ├── 📄 redis_repo.go
│   │   ├── 📄 repository.go
│   │   ├── 📄 service_test.go
│   │   └── 📄 service.go
│   ├── 📂 paymentintent/
│   │   ├── 📄 capture.go
│   │   ├── 📄 create.go
│   │   ├── 📄 dto.go
│   │   ├── 📄 error.go
│   │   ├── 📄 handler.go
│   │   ├── 📄 model.go
│   │   ├── 📄 pgsql_repo.go
│   │   ├── 📄 repository.go
│   │   ├── 📄 service_test.go
│   │   └── 📄 service.go
│   └── 📂 paymentevent/
│   │   ├── 📄 pgsql_repo.go
│   │   └── 📄 repository.go
├── 📄 makefile
├── 📂 pkg/
│   ├── 📂 uow/
│   │   └── 📄 uow.go
│   └── 🔧 utils/
│   │   └── 📄 utils.go
├── 📖 project_structure.md
└── 📄 start.sh
```

## 📖 Legend

### File Types

- ⚙️ Config: TOML files
- 🐳 DevOps: Docker ignore
- 📄 Other: Other files
- ⚙️ Config: YAML files
- ⚙️ Config: YAML files
- 🚫 DevOps: Git ignore
- 🐳 DevOps: Docker files
- 📄 Docs: Text files
- 📖 Docs: Markdown files

### Importance Levels

- 🔴 Critical: Essential project files
- 🟡 High: Important configuration files
- 🔵 Medium: Helpful but not essential files

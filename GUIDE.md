# Hướng dẫn sử dụng CodebaseGo

## Mục lục

- [1. Cài đặt & Khởi chạy nhanh](#1-cài-đặt--khởi-chạy-nhanh)
- [2. Cấu trúc dự án](#2-cấu-trúc-dự-án)
- [3. Cơ chế Module & File bắt buộc `register.go`](#3-cơ-chế-module--file-bắt-buộc-registergo)
- [4. Quản lý Bật / Tắt Module (`configs/modules.yaml`)](#4-quản-lý-bật--tắt-module-configsmodulesyaml)
- [5. Hướng dẫn thêm Module mới (Step-by-Step)](#5-hướng-dẫn-thêm-module-mới-step-by-step)
- [6. Tự động hóa Dependency Injection (`make setup`)](#6-tự-động-hóa-dependency-injection-make-setup)
- [7. API Reference](#7-api-reference)
- [8. Kết nối Database & Migration](#8-kết-nối-database--migration)
- [9. Swagger Documentation](#9-swagger-documentation)
- [10. Quy ước làm việc nhóm & Tránh conflict](#10-quy-ước-làm-việc-nhóm--tránh-conflict)
- [11. FAQ](#11-faq)

---

## 1. Cài đặt & Khởi chạy nhanh

### Yêu cầu hệ thống

- **Go:** >= 1.21
- **Make:** Có sẵn trên Linux/macOS (Windows: cài qua `choco install make`)

### Bước 1: Cài đặt dev tools và dependencies

```bash
# Cài đặt công cụ wire, swag, goose
make install-tools

# Tải dependencies
make deps
```

### Bước 2: Cấu hình ứng dụng

Copy file config mẫu và chỉnh sửa nếu cần:

```bash
cp configs/config.example.yaml configs/config.yaml
```

Các giá trị quan trọng:

| Key | Mô tả | Mặc định |
|-----|--------|----------|
| `server.port` | Port chạy HTTP server | `8080` |
| `server.mode` | Gin mode (`debug` / `release`) | `debug` |
| `server.trusted_proxies` | Danh sách IP / CIDR Proxy tin tưởng | `["127.0.0.1", "::1"]` |
| `jwt.secret` | Secret key cho JWT | `change-me-in-production` |
| `jwt.access_expire_minute` | Thời gian Access Token hết hạn (phút) | `15` |
| `jwt.refresh_expire_day` | Thời gian Refresh Token hết hạn (ngày) | `7` |
| `db.driver` | Database driver (`sqlite` / `mysql` / `postgres`) | `sqlite` |
| `db.dsn` | Chuỗi kết nối DB | `app.db` |
| `db.auto_migrate` | Bật/tắt GORM AutoMigrate khi khởi động | `true` |

### Bước 3: Đồng bộ Module & Khởi động Server

```bash
# Quét module, sinh mã Wire DI tự động
make setup

# Chạy server
make run
```

Server mặc định chạy tại: `http://localhost:8080`.

---

## 2. Cấu trúc dự án

```
CodebaseGo/
├── cmd/
│   ├── api/                     # Entry point & DI Composition Root
│   │   ├── main.go              # Hàm main, khởi tạo App & Graceful shutdown
│   │   ├── wire.go              # Wire injector template (không cần sửa tay)
│   │   ├── wire_gen.go          # Wire generated code (tự sinh bởi make setup)
│   │   └── modules_gen.go       # Module registry bindings (tự sinh bởi make setup)
│   │
│   └── tools/
│       └── modgen/              # Tool CLI tự động quét module & sinh mã DI
│           └── main.go
│
├── internal/
│   ├── modules/                 # ★ CÁC MODULE TÍNH NĂNG (Modular Monolith)
│   │   ├── auth/                # Module Xác thực (Register, Login, Token Rotation)
│   │   │   ├── register.go      # ★ File hợp đồng bắt buộc của module
│   │   │   ├── handler.go       # HTTP handlers
│   │   │   ├── service.go       # Business logic
│   │   │   ├── repository.go    # Data access interface + GORM impl
│   │   │   ├── entity.go        # GORM schema
│   │   │   └── dto.go           # Request/Response validation DTOs
│   │   │
│   │   ├── user/                # Module Quản lý User
│   │   │   ├── register.go      # ★ File hợp đồng bắt buộc
│   │   │   ├── handler.go
│   │   │   ├── service.go
│   │   │   ├── repository.go
│   │   │   ├── entity.go
│   │   │   └── dto.go
│   │   │
│   │   ├── health/              # Module Health Check
│   │   │   ├── register.go      # ★ File hợp đồng bắt buộc
│   │   │   └── handler.go
│   │   │
│   │   └── graphql/             # Module GraphQL API & Playground
│   │       ├── register.go      # ★ File hợp đồng bắt buộc
│   │       ├── resolver.go
│   │       └── schema.graphql
│   │
│   ├── platform/                # Hạ tầng kỹ thuật dùng chung
│   │   ├── config/              # Viper config loader
│   │   ├── database/            # GORM connection pool (SQLite / MySQL / Postgres)
│   │   ├── eventbus/            # In-Memory & Redis Pub/Sub EventBus
│   │   ├── logger/              # Zerolog structured logger
│   │   ├── redis/               # go-redis client wrapper
│   │   ├── response/            # Standard response envelope & error handler
│   │   └── server/              # Gin HTTP server, Rate limiter & Security headers
│   │
│   └── common/                  # Interfaces & Types dùng chung toàn hệ thống
│       ├── interfaces.go        # RouteRegistrar, Migrator, BackgroundWorker, EventBus
│       ├── errors.go            # AppError & Sentinel errors
│       └── pagination.go        # Offset & Keyset/Cursor pagination
│
├── configs/
│   ├── config.yaml              # Config ứng dụng (runtime)
│   ├── config.example.yaml      # File mẫu config ứng dụng
│   └── modules.yaml             # ★ Quản lý Bật/Tắt module & Thống kê
│
├── migrations/                  # SQL migrations (Goose)
├── Makefile                     # make setup, make run, make test, make swagger
└── README.md
```

---

## 3. Cơ chế Module & File bắt buộc `register.go`

Mỗi thư mục con trong `internal/modules/<tên_module>/` là một feature module độc lập. **Mỗi module bắt buộc phải có file `register.go`**.

File `register.go` đóng vai trò là **bản khai báo hợp đồng (Module Manifest)** cung cấp cho hệ thống:

```go
package product

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"codebasego/internal/common"
)

// Đảm bảo struct Module thỏa mãn các interface chuẩn của hệ thống
var (
	_ common.RouteRegistrar    = (*Module)(nil) // Đăng ký routes (tuỳ chọn)
	_ common.Migrator          = (*Module)(nil) // AutoMigrate DB (tuỳ chọn)
	_ common.BackgroundWorker  = (*Module)(nil) // Background job (tuỳ chọn)
)

// 1. Khai báo Wire ProviderSet cho module
var ProviderSet = wire.NewSet(
	NewGormRepository,
	wire.Bind(new(Repository), new(*GormRepository)),
	NewService,
	NewHandler,
	NewModule,
)

type Module struct {
	handler *Handler
}

func NewModule(h *Handler) *Module {
	return &Module{handler: h}
}

// 2. AutoMigrate (Tuỳ chọn) - GORM tự động tạo/cập nhật bảng
func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Product{})
}

// 3. RegisterRoutes (Tuỳ chọn) - Đăng ký HTTP endpoints
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/products")
	group.GET("", m.handler.List)
	group.GET("/:id", m.handler.GetByID)
	group.POST("", m.handler.Create)
}

// 4. StartBackground (Tuỳ chọn) - Khởi chạy worker chạy ngầm định kỳ với Graceful Shutdown
func (m *Module) StartBackground(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// xử lý background job
			}
		}
	}()
}
```

### Giao tiếp giữa các module (Cross-Module Interface Binding)

Nếu module A cần phụ thuộc vào service của module B (ví dụ module `order` cần gọi `user.Service`):
1. Module A tự định nghĩa interface `UserService` cục bộ trong package của mình.
2. Thêm comment annotation ngay trên `ProviderSet` trong `register.go`:
   ```go
   // @wire:bind target=UserService source=*user.Service
   ```
3. Khi chạy `make setup`, Wire DI sẽ tự động bind `user.Service` vào `UserService` của module A!

---

## 4. Quản lý Bật / Tắt Module (`configs/modules.yaml`)

File [`configs/modules.yaml`](file:///var/www/html/Learner/CodebaseGo/configs/modules.yaml) là trung tâm kiểm soát trạng thái của tất cả các module trong ứng dụng:

```yaml
# ==============================================================================
# MODULE REGISTRY & ACTIVATION CONFIGURATION
# ==============================================================================
# Đặt 'enabled: false' để tắt một module mà không cần xoá code.
# Chạy 'make setup' để cập nhật mã nguồn Wire DI.
# ==============================================================================

modules:
  auth:
    enabled: true
    description: "Authentication, JWT Token rotation & Reuse detection"
  user:
    enabled: true
    description: "User CRUD management & Cursor pagination"
  health:
    enabled: true
    description: "Health check & Database connectivity probe"
  graphql:
    enabled: true
    description: "GraphQL API queries, mutations & Playground UI"
  product:
    enabled: true
    description: "Product catalog management"
```

### Khi tắt một module (`enabled: false`):
1. Mở `configs/modules.yaml`, đổi `enabled: false` tại module mong muốn.
2. Chạy:
   ```bash
   make setup
   ```
3. Hệ thống sẽ tự động:
   * Loại bỏ ProviderSet của module khỏi `cmd/api/modules_gen.go`.
   * Tắt toàn bộ route HTTP, database migration, background worker của module đó.
   * Biên dịch lại `cmd/api/wire_gen.go` sạch sẽ, không tốn tài nguyên runtime.

---

## 5. Hướng dẫn thêm Module mới (Step-by-Step)

Ví dụ: Bạn muốn thêm module **`order`** (quản lý đơn hàng).

### Bước 1: Tạo thư mục module
```bash
mkdir -p internal/modules/order
```

### Bước 2: Tạo entity, DTO, repository, service, handler
* `entity.go`: Khai báo struct GORM.
* `dto.go`: Struct request/response có Gin validation tag.
* `repository.go`: Interface + implementation truy vấn database.
* `service.go`: Business logic xử lý nghiệp vụ đơn hàng.
* `handler.go`: Controller nhận request từ Gin context và trả về qua `response.Success()`.

### Bước 3: Tạo file bắt buộc `register.go`
```go
// internal/modules/order/register.go
package order

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"gorm.io/gorm"

	"codebasego/internal/common"
)

var (
	_ common.RouteRegistrar = (*Module)(nil)
	_ common.Migrator       = (*Module)(nil)
)

var ProviderSet = wire.NewSet(
	NewGormRepository,
	wire.Bind(new(Repository), new(*GormRepository)),
	NewService,
	NewHandler,
	NewModule,
)

type Module struct {
	handler *Handler
}

func NewModule(h *Handler) *Module {
	return &Module{handler: h}
}

func (m *Module) AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&Order{})
}

func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
	group := rg.Group("/orders")
	group.GET("", m.handler.List)
	group.GET("/:id", m.handler.GetByID)
	group.POST("", m.handler.Create)
}
```

### Bước 4: Chạy lệnh `make setup`
```bash
make setup
```

**Kết quả:**
* Module `order` tự động được thêm vào [`configs/modules.yaml`](file:///var/www/html/Learner/CodebaseGo/configs/modules.yaml).
* Bảng trạng thái hiển thị module `order` đang `[ENABLED]`.
* Wire DI tự động cập nhật `cmd/api/modules_gen.go` và `cmd/api/wire_gen.go`.

### Bước 5: Chạy test & khởi động server
```bash
make test
make run
```
Endpoint `/api/v1/orders` đã hoạt động ngay lập tức!

---

## 6. Tự động hóa Dependency Injection (`make setup`)

Codebase sử dụng công cụ generator thông minh [`cmd/tools/modgen/main.go`](file:///var/www/html/Learner/CodebaseGo/cmd/tools/modgen/main.go).

Mỗi khi bạn chạy `make setup` hoặc `make wire`:
1. **AST Scanner:** Đọc mã nguồn các file `register.go` trong `internal/modules/*/`.
2. **Config Sync:** Tự động đồng bộ với `configs/modules.yaml` (thêm module mới phát hiện).
3. **Status Report:** In bảng thống kê chi tiết trạng thái, migrator, route, background task của từng module.
4. **Code Generation:**
   * Sinh file [`cmd/api/modules_gen.go`](file:///var/www/html/Learner/CodebaseGo/cmd/api/modules_gen.go) chứa `EnabledModulesSet`, `ProvideMigrators`, `ProvideRouteRegistrars`, `ProvideBackgroundWorkers`.
   * Tự động gọi `wire` để sinh [`cmd/api/wire_gen.go`](file:///var/www/html/Learner/CodebaseGo/cmd/api/wire_gen.go).

---

## 7. API Reference

### Public endpoints (không cần token)

#### Health check
```bash
curl http://localhost:8080/api/v1/health
```
```json
{"success": true, "data": {"database": "up", "status": "ok"}}
```

#### Đăng ký tài khoản
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123", "name": "Nguyen Van A"}'
```
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "a1b2c3d4-...",
    "expires_in": 900,
    "token_type": "Bearer"
  }
}
```

#### Đăng nhập
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

#### Làm mới Token (Refresh Token Rotation)
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "a1b2c3d4-..."}'
```

#### Đăng xuất (Thu hồi Refresh token)
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Content-Type: application/json" \
  -d '{"refresh_token": "a1b2c3d4-..."}'
```

### Protected endpoints (cần Bearer Token)

```bash
TOKEN="eyJhbGciOiJIUzI1NiIs..."
```

#### Danh sách Users (Cursor Pagination & Offset Pagination)
```bash
# Cursor pagination (Khuyên dùng cho dữ liệu lớn)
curl "http://localhost:8080/api/v1/users?limit=10" \
  -H "Authorization: Bearer $TOKEN"

# Offset pagination truyền thống
curl "http://localhost:8080/api/v1/users?page=1&per_page=10" \
  -H "Authorization: Bearer $TOKEN"
```

#### Lấy chi tiết User theo ID
```bash
curl http://localhost:8080/api/v1/users/{id} \
  -H "Authorization: Bearer $TOKEN"
```

#### Cập nhật User
```bash
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Tên mới"}'
```

#### Đăng xuất khỏi mọi thiết bị (Logout All)
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout-all \
  -H "Authorization: Bearer $TOKEN"
```

### GraphQL API

* **GraphQL Endpoint:** `POST http://localhost:8080/api/v1/graphql`
* **Interactive Playground:** `GET http://localhost:8080/api/v1/playground` (chỉ bật trên môi trường Dev/Staging)

---

## 8. Kết nối Database & Migration

Codebase hỗ trợ 3 database engines qua GORM: **SQLite**, **MySQL**, **PostgreSQL**.

### 1. Chế độ Local / Development (SQLite mặc định)
Tự động tạo file `app.db` và tự động tạo bảng (GORM AutoMigrate) khi chạy `make run`:
```yaml
# configs/config.yaml
db:
  driver: "sqlite"
  dsn: "app.db"
  auto_migrate: true
```

### 2. Chế độ Production (PostgreSQL / MySQL)
Chuyển đổi DSN và tắt `auto_migrate` để sử dụng SQL Migrations chuẩn:
```yaml
# configs/config.yaml
db:
  driver: "postgres"
  dsn: "host=localhost user=postgres password=password dbname=app port=5432 sslmode=disable"
  auto_migrate: false
```

### 3. Quản lý SQL Migrations với Goose
```bash
# Chạy migration lên phiên bản mới nhất
make migrate-up

# Rollback 1 migration gần nhất
make migrate-down

# Kiểm tra trạng thái migration
make migrate-status
```

---

## 9. Swagger Documentation

### Tạo Swagger Docs từ comment annotations:
```bash
make swagger
```

Xem giao diện tương tác tại: `http://localhost:8080/swagger/index.html`

### Swagger Annotation mẫu trên Handler:
```go
// @Summary      Lấy danh sách sản phẩm
// @Description  Trả về danh sách sản phẩm phân trang
// @Tags         products
// @Accept       json
// @Produce      json
// @Param        limit  query     int  false  "Số lượng mỗi trang" default(10)
// @Success      200    {object}  response.Body{data=[]Response}
// @Failure      401    {object}  response.Body
// @Security     BearerAuth
// @Router       /products [get]
```

---

## 10. Quy ước làm việc nhóm & Tránh conflict

### Mô hình nhánh Git
```
main
  └── develop
        ├── feature/auth        ← Dev A (chỉ code trong modules/auth)
        ├── feature/product     ← Dev B (chỉ code trong modules/product)
        └── feature/order       ← Dev C (chỉ code trong modules/order)
```

### Không bao giờ gây conflict khi Merge
* Mỗi developer **chỉ tạo và sửa code trong thư mục module của mình** (`internal/modules/<tên>/`).
* Developer **không cần sửa `cmd/api/wire.go` hay `cmd/api/main.go`**.
* Khi pull code mới hoặc checkout branch, chỉ cần chạy:
  ```bash
  make setup
  ```
  Hệ thống sẽ tự động tổng hợp toàn bộ module của các thành viên khác!

---

## 11. FAQ

### Q: Tôi tạo module mới nhưng server không nhận route?
**A:** Hãy kiểm tra xem bạn đã chạy `make setup` chưa. Lệnh `make setup` sẽ tự động quét file `register.go` của bạn và nạp vào router.

### Q: Làm sao để tạm thời tắt một module đang phát triển dở dang?
**A:** Mở `configs/modules.yaml`, tìm tên module đó và đặt `enabled: false`, sau đó chạy `make setup`.

### Q: Có cần commit các file `wire_gen.go` và `modules_gen.go` lên Git không?
**A: Có.** Nên commit cả `modules_gen.go` và `wire_gen.go` để các thành viên khác hoặc CI/CD pipeline có thể build binary ngay mà không bắt buộc cài đặt công cụ `wire` CLI trên máy chủ build.

### Q: Làm sao để viết Unit Test cho một module?
**A:** Tạo file `*_test.go` cùng thư mục với module. Nhờ kiến trúc Clean Architecture & Repository Interface, bạn có thể dễ dàng mock repository và test service logic:
```bash
make test
```

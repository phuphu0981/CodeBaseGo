# CodebaseGo

Modular Go + Gin backend API with clean architecture.

## Architecture

```
cmd/api/          → Entry point + Wire DI
internal/
  modules/        → Feature modules (auth, user, health, ...)
  platform/       → Shared infrastructure (config, logger, server, response)
  common/         → Shared domain types (errors, pagination)
```

**Thêm module mới (Chỉ 2 bước):**
1. Tạo `internal/modules/<tên>/` với file bắt buộc `register.go` (chứa `ProviderSet`, `RegisterRoutes`, `AutoMigrate` nếu có).
2. Chạy lệnh:
```bash
make setup
```
Lệnh trên sẽ tự động phát hiện module mới, cập nhật danh sách vào `configs/modules.yaml`, đồng bộ và sinh code Wire DI hoàn toàn tự động!

## Quick Start

```bash
# Install dev tools
make install-tools

# Download dependencies
make deps

# Run the server
make run
```

Server sẽ chạy tại `http://localhost:8080`.

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | /api/v1/health | ✗ | Comprehensive health check (DB + Redis) |
| GET | /api/v1/health/live | ✗ | Liveness probe (K8s/LB) |
| GET | /api/v1/health/ready | ✗ | Readiness probe (K8s/LB) |
| POST | /api/v1/auth/register | ✗ | Register (trả Access + Refresh token) |
| POST | /api/v1/auth/login | ✗ | Login (trả Access + Refresh token) |
| POST | /api/v1/auth/refresh | ✗ | Lấy Access token mới từ Refresh token |
| POST | /api/v1/auth/logout | ✗ | Thu hồi (Revoke) Refresh token |
| GET | /api/v1/users | ✓ | List users (REST) |
| GET | /api/v1/users/:id | ✓ | Get user (REST) |
| PUT | /api/v1/users/:id | ✓ | Update user (REST) |
| DELETE | /api/v1/users/:id | ✓ | Delete user (REST) |
| POST | /api/v1/graphql | Tuỳ | GraphQL API Endpoint (Queries & Mutations) |
| GET | /api/v1/playground | ✗ | GraphQL Interactive Playground UI |



## Swagger

```bash
make swagger    # Generate docs
# Then visit http://localhost:8080/swagger/index.html
```

## Documentation

- 📘 [Tài liệu Kiến trúc & Coding Standard](file:///var/www/html/Learner/CodebaseGo/coding_standard_and_architecture.md)
- 🔄 [Tài liệu Luồng xử lý Request (Request Lifecycle)](file:///var/www/html/Learner/CodebaseGo/request_lifecycle.md)
- 📖 [Hướng dẫn phát triển chi tiết (GUIDE.md)](file:///var/www/html/Learner/CodebaseGo/GUIDE.md)

## Code Quality & Architecture Check

Kiểm tra tuân thủ kiến trúc Modular Monolith & Clean Architecture theo [coding_standard_and_architecture.md](file:///var/www/html/Learner/CodebaseGo/coding_standard_and_architecture.md):

```bash
make archcheck   # Kiểm tra kiến trúc 4 tầng, ranh giới import, DTO bảo mật, quy ước đặt tên
make check       # Chạy toàn bộ archcheck + unit tests
make lint        # Chạy golangci-lint (nếu đã cài đặt)
```

## Tech Stack

- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **DI:** [Google Wire](https://github.com/google/wire)
- **Config:** [Viper](https://github.com/spf13/viper)
- **Logger:** [Zerolog](https://github.com/rs/zerolog)
- **JWT:** [golang-jwt](https://github.com/golang-jwt/jwt)
- **Swagger:** [Swaggo](https://github.com/swaggo/swag)

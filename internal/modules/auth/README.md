# Module Auth - Authentication & Token Security

## 1. Giới thiệu tổng quan

Module **Auth** đảm nhiệm toàn bộ nghiệp vụ xác thực người dùng, bảo mật phiên đăng nhập và quản lý vòng đời JSON Web Token (JWT). Module được thiết kế tuân thủ nghiêm ngặt các tiêu chuẩn bảo mật hiện đại:
- **JWT Access Token:** Thời gian sống ngắn (Short-lived), chứa claims nhận diện người dùng.
- **Refresh Token Rotation (RTR):** Mỗi lần làm mới token thành công, Refresh Token cũ sẽ bị hủy ngay lập tức và cấp 1 cặp Token mới hoàn toàn.
- **Phát hiện tái sử dụng Token (Token Reuse Detection):** Nếu một Refresh Token đã thu hồi bị kẻ gian gửi lại sau thời gian ân hạn (grace period 15s), hệ thống tự động vô hiệu hóa toàn bộ phiên đăng nhập của người dùng đó (Revoke All Tokens).
- **Tự động dọn dẹp (Background Worker):** Chạy tác vụ ngầm định kỳ 24h/lần để xóa các token đã hết hạn khỏi database và hỗ trợ Graceful Shutdown.

---

## 2. Cấu hình & Biến môi trường

Module sử dụng các tham số cấu hình trong [`configs/config.yaml`](../../../configs/config.yaml):

| Khóa cấu hình | Biến môi trường | Mặc định | Mô tả |
| :--- | :--- | :--- | :--- |
| `jwt.secret` | `JWT_SECRET` | `change-me-in-production` | Khóa bí mật dùng để ký và xác thực JWT |
| `jwt.access_expire_minute` | `JWT_ACCESS_EXPIRE_MINUTE` | `15` | Thời gian sống của Access Token (phút) |
| `jwt.refresh_expire_day` | `JWT_REFRESH_EXPIRE_DAY` | `7` | Thời gian sống của Refresh Token (ngày) |

---

## 3. Cấu trúc Database (`refresh_tokens`)

| Cột | Kiểu dữ liệu | Mô tả |
| :--- | :--- | :--- |
| `id` | `VARCHAR(36)` | UUID khóa chính |
| `user_id` | `VARCHAR(36)` | Khóa ngoại tham chiếu `users(id)` |
| `token` | `VARCHAR(64)` | Giá trị Refresh Token (Unique Index) |
| `expires_at` | `DATETIME` | Thời điểm hết hạn |
| `revoked` | `BOOLEAN` | Cờ đánh dấu đã bị thu hồi (`true`/`false`) |
| `created_at`, `updated_at` | `DATETIME` | Thời gian tạo và cập nhật |

---

## 4. Danh sách API Endpoints

| Method | Endpoint | Quyền | Mô tả |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/register` | Public | Đăng ký tài khoản mới (phát sinh Event `user.registered`) |
| `POST` | `/api/v1/auth/login` | Public | Đăng nhập bằng Email & Password, trả về cặp Token |
| `POST` | `/api/v1/auth/refresh` | Public | Đổi Refresh Token lấy Access Token mới (Token Rotation) |
| `POST` | `/api/v1/auth/logout` | Public | Thu hồi Refresh Token hiện tại |
| `POST` | `/api/v1/auth/logout-all`| Bearer Token | Đăng xuất và thu hồi toàn bộ Refresh Token của user trên mọi thiết bị |

---

## 5. Tích hợp Auth Middleware cho các Module khác

Module Auth cung cấp `common.AuthMiddleware` để bảo vệ các endpoint cần quyền đăng nhập:

```go
// Trong file register.go của module cần bảo vệ (ví dụ: user, page, setting)
func (m *Module) RegisterRoutes(rg *gin.RouterGroup) {
    protected := rg.Group("/admin", gin.HandlerFunc(m.authMiddleware))
    protected.POST("/action", m.handler.DoAction)
}
```
Khi client gửi request kèm Header `Authorization: Bearer <access_token>`, middleware tự động:
1. Giải mã và kiểm tra chữ ký JWT.
2. Nạp `user_id` và `email` vào `c.Set("user_id", ...)` và `c.Set("email", ...)` của Gin Context.

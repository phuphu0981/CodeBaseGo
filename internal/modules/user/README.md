# Module User - User Account Management

## 1. Giới thiệu tổng quan

Module **User** quản lý thông tin tài khoản người dùng, hồ sơ cá nhân và phân quyền thao tác. Module hỗ trợ:
- **Bảo mật mật khẩu:** Băm mật khẩu một chiều với thuật toán `bcrypt` chuẩn công nghiệp.
- **Phân trang hiệu năng cao:** Hỗ trợ song song **Keyset / Cursor Pagination** (khuyên dùng cho bảng dữ liệu lớn) và **Offset Pagination** truyền thống.
- **Phát / Nhận Sự kiện (Domain Events):** Định nghĩa sự kiện `user.registered` khi có người dùng mới tạo tài khoản để các module khác lắng nghe (ví dụ gửi mail chào mừng, thống kê).

---

## 2. Cấu trúc Database (`users`)

| Cột | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :--- | :--- | :--- | :--- |
| `id` | `VARCHAR(36)` | Primary Key | UUID định danh người dùng |
| `email` | `VARCHAR(255)` | Unique Index, Not Null | Địa chỉ email đăng nhập |
| `password` | `VARCHAR(255)` | Not Null | Mật khẩu đã được băm (bcrypt) |
| `name` | `VARCHAR(255)` | Not Null | Họ và tên hiển thị |
| `created_at`| `DATETIME` | Index, Not Null | Thời gian tạo tài khoản |
| `updated_at`| `DATETIME` | Not Null | Thời gian cập nhật thông tin |

---

## 3. Danh sách API Endpoints

Tất cả các endpoint của module User đều yêu cầu Header xác thực: `Authorization: Bearer <access_token>`

| Method | Endpoint | Quyền | Mô tả |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/users` | Bearer Token | Danh sách người dùng (Cursor pagination: `?limit=10&cursor=...` hoặc Offset: `?page=1&per_page=10`) |
| `GET` | `/api/v1/users/:id` | Bearer Token | Xem chi tiết thông tin tài khoản theo ID |
| `PUT` | `/api/v1/users/:id` | Bearer Token | Cập nhật hồ sơ (Chỉ cho phép cập nhật tài khoản của chính mình) |
| `DELETE`| `/api/v1/users/:id` | Bearer Token | Xóa tài khoản (Chỉ cho phép xóa tài khoản của chính mình) |

---

## 4. Domain Events

Module xuất bản các sự kiện miền qua `EventBus`:

```go
// Tên sự kiện
const EventUserRegistered = "user.registered"

// Payload dữ liệu
type UserRegisteredPayload struct {
    UserID    string    `json:"user_id"`
    Email     string    `json:"email"`
    Name      string    `json:"name"`
    CreatedAt time.Time `json:"created_at"`
}
```
Các module khác có thể đăng ký lắng nghe sự kiện này:
```go
eventBus.Subscribe(user.EventUserRegistered, func(ctx context.Context, event common.Event) error {
    var payload user.UserRegisteredPayload
    _ = common.DecodeEventPayload(event.Payload, &payload)
    // Thực hiện gửi email kích hoạt, tặng điểm thưởng,...
    return nil
})
```

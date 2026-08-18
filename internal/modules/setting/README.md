# Module Setting - Dynamic System Configuration (`core_config_data`)

## 1. Giới thiệu tổng quan

Module **Setting** cung cấp cơ chế lưu trữ và quản lý cấu hình hệ thống động theo dạng đường dẫn phân cấp (`Path`), phạm vi (`Scope`), và giá trị (`Value`) tương tự như bảng `core_config_data` trong Magento 2.

Module này cho phép:
- Lưu trữ các cấu hình chung như `base_url`, `api_base_url`, `store_name`, `timezone`.
- Mở rộng cấu hình linh hoạt cho bất kỳ module mới nào trong tương lai mà không cần thay đổi schema database.
- Cung cấp API Public an toàn để Frontend Next.js đọc các cấu hình môi trường/hiển thị.

---

## 2. Cấu trúc bảng `core_config_data`

| Cột | Kiểu dữ liệu | Mặc định | Mô tả |
| :--- | :--- | :--- | :--- |
| `id` | `VARCHAR(36)` | Primary Key | UUID bản ghi |
| `scope` | `VARCHAR(20)` | `'default'` | Phạm vi cấu hình (`default`, `website`, `store`) |
| `scope_id` | `VARCHAR(36)` | `'0'` | Mã định danh của scope |
| `path` | `VARCHAR(255)` | Not Null | Đường dẫn phân cấp (ví dụ: `web/unsecure/base_url`) |
| `value` | `TEXT` | Nullable | Giá trị cấu hình |
| `created_at`| `DATETIME` | Not Null | Thời điểm tạo |
| `updated_at`| `DATETIME` | Not Null | Thời điểm cập nhật |

> **Ràng buộc duy nhất:** `UNIQUE(scope, scope_id, path)` đảm bảo không bị duplicate key trong cùng một phạm vi.

---

## 3. Các đường dẫn cấu hình mặc định (Default Paths)

| Path | Giá trị mặc định | Mục đích |
| :--- | :--- | :--- |
| `web/unsecure/base_url` | `http://localhost:3000` | Base URL của Frontend (Next.js) |
| `web/secure/base_url` | `http://localhost:3000` | Base URL bảo mật (HTTPS) |
| `web/api_base_url` | `http://localhost:8080/api/v1` | URL gốc của Backend API |
| `general/store_information/name` | `CodebaseGo Store` | Tên hệ thống / cửa hàng |
| `general/locale/timezone` | `Asia/Ho_Chi_Minh` | Múi giờ chuẩn |

---

## 4. Danh sách API Endpoints

### 4.1. Lấy cấu hình công khai (`GET /api/v1/settings/public` - Public)
Dành cho Next.js đọc nhanh các thông tin base URL và thông tin chung:
```json
{
  "success": true,
  "data": {
    "base_url": "http://localhost:3000",
    "api_base_url": "http://localhost:8080/api/v1",
    "store_name": "CodebaseGo Store",
    "timezone": "Asia/Ho_Chi_Minh"
  }
}
```

### 4.2. Lấy cấu hình theo Path (`GET /api/v1/settings/by-path?path=web/unsecure/base_url` - Public)
Trả về chi tiết 1 bản ghi cấu hình theo path.

### 4.3. Quản lý cấu hình (Protected - Bearer Token)
- `GET /api/v1/settings`: Danh sách toàn bộ cấu hình (hỗ trợ lọc `?path_prefix=web/` hoặc `?scope=default`).
- `GET /api/v1/settings/:id`: Chi tiết cấu hình theo ID.
- `POST /api/v1/settings`: Thêm hoặc cập nhật (Upsert) cấu hình theo path:
  ```json
  {
    "scope": "default",
    "scope_id": "0",
    "path": "payment/stripe/enabled",
    "value": "1"
  }
  ```
- `DELETE /api/v1/settings/:id`: Xóa cấu hình.

---

## 5. Sử dụng trong mã nguồn Go (Cross-Module Usage)

Các module khác có thể inject `setting.Service` để lấy giá trị cấu hình:

```go
// Lấy giá trị cấu hình Base URL
baseURL := settingService.GetBaseURL(ctx)

// Lấy giá trị bất kỳ với giá trị fallback mặc định
stripeKey := settingService.GetWithDefault(ctx, "payment/stripe/public_key", "")
```

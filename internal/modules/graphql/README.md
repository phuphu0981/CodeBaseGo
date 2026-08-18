# Module GraphQL - Flexible GraphQL API & Playground

## 1. Giới thiệu tổng quan

Module **GraphQL** cung cấp giao thức truy vấn dữ liệu linh hoạt dựa trên thư viện **gqlgen** (Code-first / Schema-first hàng đầu cho Golang). Module cho phép:
- Client (Frontend Web/Mobile) tự do định nghĩa các trường dữ liệu cần lấy (Over-fetching & Under-fetching prevention).
- Tích hợp sẵn **Interactive GraphQL Playground UI** trực quan để test query trên môi trường Development.
- **Bảo mật:** Tích hợp bộ giới hạn độ phức tạp câu truy vấn (Query Complexity Limiter = 100) để chống tấn công DoS/ReDoS.
- **Xác thực linh hoạt:** Tự động giải mã Bearer JWT token (nếu có) và nạp `user_id` vào GraphQL Context.

---

## 2. Danh sách Endpoints

| Method | Endpoint | Môi trường | Mô tả |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/graphql` | Toàn bộ | Endpoint chính tiếp nhận các câu truy vấn GraphQL (Query & Mutation) |
| `GET` | `/api/v1/playground` | Dev / Staging | Giao diện đồ họa tương tác và kiểm thử GraphQL (tự động tắt khi `server.mode: release`) |

---

## 3. Các truy vấn mẫu (GraphQL Examples)

### 3.1. Truy vấn thông tin người dùng đang đăng nhập (`me`)
```graphql
query GetMyProfile {
  me {
    id
    email
    name
    createdAt
  }
}
```
*Yêu cầu gửi kèm Header:* `Authorization: Bearer <access_token>`

### 3.2. Danh sách người dùng (`users`)
```graphql
query GetUsersList {
  users(page: 1, perPage: 10) {
    id
    email
    name
  }
}
```

### 3.3. Mutation Đăng ký tài khoản (`register`)
```graphql
mutation RegisterUser {
  register(input: {
    email: "test@example.com"
    password: "password123"
    name: "Nguyen Van A"
  }) {
    user {
      id
      email
      name
    }
    accessToken
    refreshToken
  }
}
```

---

## 4. Tự động sinh mã GraphQL (Codegen)

Khi chỉnh sửa file [`schema.graphql`](./schema.graphql), bạn có thể chạy:
```bash
go run github.com/99designs/gqlgen generate
```
Hệ thống sẽ tự động cập nhật các Resolver stubs và Executable Schema.

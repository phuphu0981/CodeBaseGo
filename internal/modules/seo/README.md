# Module SEO - Tài liệu Hướng dẫn & Tối ưu SEO

## 1. Giới thiệu tổng quan

Module **SEO** cung cấp giải pháp quản lý toàn diện siêu dữ liệu (metadata) cho từng trang (page) trong dự án theo đường dẫn định danh (`slug`). Module được thiết kế phục vụ tối ưu On-Page SEO, Open Graph cho mạng xã hội, Structured Data (JSON-LD) cho Google Rich Snippets và tương thích cao với các nền tảng Server-Side Rendering (SSR) như Next.js, NuxtJS, Remix.

---

## 2. Các tính năng tối ưu SEO đã tích hợp

### 2.1. Chuẩn hóa & Quản lý URL Slug (`NormalizeSlug`)
- **Tránh trùng lặp nội dung (Duplicate Content Penalty):**
  - Tự động cắt bỏ khoảng trắng thừa và dấu gạch chéo đầu/cuối (`/about/` $\rightarrow$ `about`).
  - Chuyển toàn bộ ký tự về chữ thường (`lowercase`).
  - Xử lý đường dẫn trang chủ rỗng hoặc `/` về slug chuẩn `home`.
- **Ràng buộc duy nhất (`Unique Index`):** Đảm bảo mỗi page chỉ có duy nhất 1 bản ghi cấu hình SEO.

### 2.2. Siêu dữ liệu chuẩn tìm kiếm (On-Page Meta Tags)
- **`title` & `description`:** Quản lý tiêu đề và mô tả chuẩn SEO hiển thị trực tiếp trên kết quả tìm kiếm (Google SERP).
- **`canonical_url`:** Khai báo liên kết chuẩn gốc nhằm giải quyết triệt để lỗi trùng lặp nội dung khi URL chứa các tham số theo dõi (UTM campaign, pagination, affiliate parameters).
- **`robots`:** Điều hướng các Search Engine Bot (mặc định: `index, follow`), có thể tùy biến thành `noindex, nofollow` cho các trang riêng tư/thanh toán.
- **`keywords`:** Lưu trữ từ khóa phục vụ phân loại và đo lường nội dung.

### 2.3. Tối ưu Mạng xã hội (Open Graph & Twitter Cards)
- **Open Graph Protocol (`og_title`, `og_description`, `og_image`, `og_type`):** Đảm bảo hiển thị Rich Preview chuyên nghiệp với hình ảnh banner, tiêu đề và tóm tắt rõ ràng khi chia sẻ link lên Facebook, Zalo, LinkedIn,...
- **Twitter Cards (`twitter_card`):** Mặc định sử dụng thẻ `summary_large_image` hiển thị ảnh xem trước kích thước lớn.

### 2.4. Dữ liệu có cấu trúc Structured Data (JSON-LD / Schema.org)
- **`structured_data`:** Trường lưu chuỗi JSON-LD hợp lệ.
- Hỗ trợ các schema chuẩn Schema.org như `Article`, `Product`, `Organization`, `BreadcrumbList`, `FAQPage`, `LocalBusiness`,...
- Tăng khả năng hiển thị **Google Rich Snippets** (đánh giá sao, giá sản phẩm, khối câu hỏi thường gặp FAQ).

### 2.5. Tối ưu Hiệu năng & Tích hợp SSR
- **Truy vấn cực nhanh:** Đánh chỉ mục riêng `idx_seos_slug` và `idx_seos_created_at`, cho phép Frontend SSR lấy metadata chỉ trong vài mili-giây.
- **Hỗ trợ phân trang 2 chế độ:** Hỗ trợ Keyset/Cursor Pagination (hiệu năng cao cho tập dữ liệu lớn) và Offset Pagination truyền thống.

---

## 3. Cấu trúc trường dữ liệu (Schema Definition)

| Trường | Kiểu dữ liệu | Ràng buộc | Mô tả |
| :--- | :--- | :--- | :--- |
| `id` | `VARCHAR(36)` | Primary Key | Mã định danh UUID |
| `slug` | `VARCHAR(255)` | Unique Index, Not Null | Đường dẫn định danh của page (ví dụ: `home`, `about-us`) |
| `title` | `VARCHAR(255)` | Not Null | Tiêu đề trang (`<title>` và SERP title) |
| `description` | `TEXT` | Nullable | Mô tả meta description |
| `keywords` | `VARCHAR(500)` | Nullable | Danh sách từ khóa SEO |
| `canonical_url`| `VARCHAR(500)` | Nullable | Đường dẫn URL chuẩn gốc |
| `og_title` | `VARCHAR(255)` | Nullable | Tiêu đề hiển thị Open Graph |
| `og_description`| `TEXT` | Nullable | Mô tả hiển thị Open Graph |
| `og_image` | `VARCHAR(500)` | Nullable | Link ảnh banner đại diện |
| `og_type` | `VARCHAR(50)` | Default `'website'` | Loại nội dung (`website`, `article`, `product`) |
| `twitter_card` | `VARCHAR(50)` | Default `'summary_large_image'` | Định dạng thẻ Twitter |
| `robots` | `VARCHAR(100)`| Default `'index, follow'` | Chỉ thị cho bot thu thập dữ liệu |
| `structured_data`| `TEXT` | Nullable | Payload JSON-LD dữ liệu có cấu trúc |
| `created_at` | `DATETIME` | Index, Not Null | Thời điểm tạo bản ghi |
| `updated_at` | `DATETIME` | Not Null | Thời điểm cập nhật bản ghi |

---

## 4. Danh sách API Endpoints

### 4.1. Lấy thông tin SEO theo Slug (Public)
Dùng cho Frontend / SSR Server để lấy metadata trước khi render HTML.

- **Endpoint:** `GET /api/v1/seo/by-slug?slug=about-us`
- **Response mẫu (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "e9b28a2a-4b92-41ce-8ceb-50da7d92bf7e",
    "slug": "about-us",
    "title": "Về chúng tôi | Công ty ABC",
    "description": "Tìm hiểu thêm về tầm nhìn, sứ mệnh và đội ngũ của ABC.",
    "keywords": "about us, cong ty abc, gioi thieu",
    "canonical_url": "https://example.com/about-us",
    "og_title": "Về chúng tôi | Công ty ABC",
    "og_description": "Tìm hiểu thêm về tầm nhìn, sứ mệnh và đội ngũ của ABC.",
    "og_image": "https://example.com/images/about-og.jpg",
    "og_type": "website",
    "twitter_card": "summary_large_image",
    "robots": "index, follow",
    "structured_data": "{\"@context\":\"https://schema.org\",\"@type\":\"AboutPage\",\"name\":\"Về chúng tôi\"}",
    "created_at": "2026-08-18T10:00:00Z",
    "updated_at": "2026-08-18T10:00:00Z"
  }
}
```

### 4.2. Danh sách tất cả bản ghi SEO (Public)
Hỗ trợ cả phân trang Cursor và Offset:
- **Cursor pagination:** `GET /api/v1/seo?limit=10&cursor=...`
- **Offset pagination:** `GET /api/v1/seo?page=1&per_page=10`

### 4.3. Lấy thông tin SEO theo ID (Public)
- **Endpoint:** `GET /api/v1/seo/:id`

### 4.4. Tạo mới cấu hình SEO (Protected - Bearer Token)
- **Endpoint:** `POST /api/v1/seo`
- **Headers:** `Authorization: Bearer <token>`
- **Body:**
```json
{
  "slug": "san-pham/dien-thoai-xyz",
  "title": "Điện thoại XYZ Chính Hãng | Giá Tốt Nhất",
  "description": "Mua Điện thoại XYZ bảo hành 12 tháng, giao hàng hỏa tốc trong 2h.",
  "canonical_url": "https://example.com/san-pham/dien-thoai-xyz",
  "og_title": "Điện thoại XYZ Chính Hãng",
  "og_image": "https://example.com/images/xyz.jpg",
  "og_type": "product",
  "structured_data": "{\"@context\":\"https://schema.org\",\"@type\":\"Product\",\"name\":\"Điện thoại XYZ\",\"offers\":{\"@type\":\"Offer\",\"price\":\"12000000\",\"priceCurrency\":\"VND\"}}"
}
```

### 4.5. Cập nhật cấu hình SEO (Protected - Bearer Token)
- **Endpoint:** `PUT /api/v1/seo/:id`
- **Headers:** `Authorization: Bearer <token>`

### 4.6. Xóa cấu hình SEO (Protected - Bearer Token)
- **Endpoint:** `DELETE /api/v1/seo/:id`
- **Headers:** `Authorization: Bearer <token>`

---

## 5. Ví dụ tích hợp với Next.js (App Router / Pages Router)

### Next.js App Router (`generateMetadata`)
```typescript
import { Metadata } from 'next';

export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const res = await fetch(`https://api.example.com/api/v1/seo/by-slug?slug=${params.slug || 'home'}`);
  const { data: seo } = await res.json();

  if (!seo) return {};

  return {
    title: seo.title,
    description: seo.description,
    keywords: seo.keywords ? seo.keywords.split(', ') : undefined,
    alternates: {
      canonical: seo.canonical_url || undefined,
    },
    robots: seo.robots,
    openGraph: {
      title: seo.og_title || seo.title,
      description: seo.og_description || seo.description,
      images: seo.og_image ? [{ url: seo.og_image }] : [],
      type: seo.og_type || 'website',
    },
    twitter: {
      card: seo.twitter_card || 'summary_large_image',
      title: seo.og_title || seo.title,
      description: seo.og_description || seo.description,
      images: seo.og_image ? [seo.og_image] : [],
    },
  };
}
```

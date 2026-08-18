# Module Page - Dynamic CMS & Page Builder Integration

## 1. Giới thiệu tổng quan

Module **Page** quản lý nội dung, template layout và trạng thái xuất bản của các trang tĩnh / landing pages / custom pages trong hệ thống. Module được thiết kế để kết hợp hoàn hảo với **Next.js App Router** và Module [`seo`](../seo/README.md).

---

## 2. Kiến trúc dữ liệu Block-based Component

Trường `content` trong bảng `pages` lưu trữ cấu trúc JSON đại diện cho các Blocks/Sections giao diện:

```json
{
  "sections": [
    {
      "type": "hero_banner",
      "data": {
        "title": "Nền tảng thương mại điện tử hiện đại",
        "subtitle": "Xây dựng trên Go và Next.js",
        "cta_text": "Bắt đầu ngay",
        "cta_link": "/register",
        "background_image": "https://example.com/hero.jpg"
      }
    },
    {
      "type": "features_grid",
      "data": {
        "columns": 3,
        "items": [
          {"icon": "zap", "title": "Siêu nhanh", "desc": "TTFB < 50ms nhờ Edge ISR"},
          {"icon": "shield", "title": "Bảo mật", "desc": "JWT Rotation & Type safety"},
          {"icon": "search", "title": "Chuẩn SEO", "desc": "Tích hợp sẵn JSON-LD & OG"}
        ]
      }
    }
  ]
}
```

---

## 3. Danh sách API Endpoints

| Method | Endpoint | Quyền | Mô tả |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/pages/slugs` | Public | Danh sách tất cả slug đã publish (dùng cho Next.js `generateStaticParams`) |
| `GET` | `/api/v1/pages/by-slug?slug=home` | Public | Lấy chi tiết nội dung trang theo slug |
| `GET` | `/api/v1/pages` | Public | Danh sách pages (hỗ trợ phân trang Cursor & Offset) |
| `GET` | `/api/v1/pages/:id` | Public | Lấy chi tiết trang theo ID |
| `POST` | `/api/v1/pages` | Bearer Token | Tạo trang mới |
| `PUT` | `/api/v1/pages/:id` | Bearer Token | Cập nhật trang |
| `DELETE`| `/api/v1/pages/:id` | Bearer Token | Xóa trang |

---

## 4. Hướng dẫn tích hợp với Next.js App Router

### `app/[slug]/page.tsx`
```typescript
import { notFound } from 'next/navigation';
import { Metadata } from 'next';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

// 1. Pre-render tất cả static pages khi Build
export async function generateStaticParams() {
  const res = await fetch(`${API_BASE}/pages/slugs`, { next: { revalidate: 3600 } });
  const { data: slugs } = await res.json();
  return (slugs || []).map((slug: string) => ({ slug }));
}

// 2. Tự động lấy SEO Metadata từ Module SEO
export async function generateMetadata({ params }: { params: { slug: string } }): Promise<Metadata> {
  const slug = params.slug || 'home';
  const res = await fetch(`${API_BASE}/seo/by-slug?slug=${slug}`, {
    next: { tags: [`seo-${slug}`], revalidate: 86400 }
  });
  const { data: seo } = await res.json();
  if (!seo) return {};

  return {
    title: seo.title,
    description: seo.description,
    alternates: { canonical: seo.canonical_url },
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
    }
  };
}

// 3. Render nội dung Server Component
export default async function Page({ params }: { params: { slug: string } }) {
  const slug = params.slug || 'home';
  const [pageRes, seoRes] = await Promise.all([
    fetch(`${API_BASE}/pages/by-slug?slug=${slug}`, { next: { tags: [`page-${slug}`] } }),
    fetch(`${API_BASE}/seo/by-slug?slug=${slug}`, { next: { tags: [`seo-${slug}`] } }),
  ]);

  const { data: page } = await pageRes.json();
  const { data: seo } = await seoRes.json();

  if (!page || page.status !== 'published') notFound();

  let contentData: any = {};
  try {
    contentData = JSON.parse(page.content || '{}');
  } catch (e) {}

  return (
    <article className="min-h-screen">
      {/* Google Rich Snippets JSON-LD */}
      {seo?.structured_data && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: seo.structured_data }}
        />
      )}

      {/* Render Dynamic Blocks */}
      {contentData.sections?.map((section: any, idx: number) => (
        <BlockRenderer key={idx} section={section} />
      ))}
    </article>
  );
}

function BlockRenderer({ section }: { section: any }) {
  switch (section.type) {
    case 'hero_banner':
      return <div className="hero">{section.data?.title}</div>;
    case 'features_grid':
      return <div className="features">{section.data?.items?.length} items</div>;
    default:
      return null;
  }
}
```

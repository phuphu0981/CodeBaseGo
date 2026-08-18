# Module Health - System Probes & Connectivity Checks

## 1. Giới thiệu tổng quan

Module **Health** cung cấp các cổng kiểm tra trạng thái hoạt động (Health Checks) phục vụ cho hệ thống giám sát hạ tầng (Monitoring), Load Balancer (Nginx, AWS ALB, Cloudflare) và cơ chế tự phục hồi của **Kubernetes / Docker Swarm (Liveness & Readiness Probes)**.

---

## 2. Các điểm kiểm tra trạng thái (Endpoints)

Tất cả các endpoint của module Health là **Public** và không yêu cầu xác thực để các công cụ giám sát có thể thăm dò định kỳ.

### 2.1. Kiểm tra toàn diện (`GET /api/v1/health`)
Kiểm tra kết nối tới cả Database (GORM) và Redis (nếu bật):
- **Status 200 OK:**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "database": "up",
    "redis": "up"
  }
}
```
- **Status 503 Service Unavailable:** Khi mất kết nối Database hoặc Redis.

### 2.2. Liveness Probe (`GET /api/v1/health/live`)
- **Mục đích:** Báo cáo xem tiến trình Go HTTP Server có đang sống hay không (Process Alive).
- Dùng cho Kubernetes Liveness Probe để tự động khởi động lại container (Restart Container) nếu ứng dụng bị treo (Deadlock).

### 2.3. Readiness Probe (`GET /api/v1/health/ready`)
- **Mục đích:** Báo cáo xem ứng dụng đã sẵn sàng nhận lưu lượng truy cập (Traffic) hay chưa.
- Kiểm tra kết nối mạng tới Database & Redis. Nếu DB chưa sẵn sàng, Kubernetes hoặc Load Balancer sẽ tạm thời ngắt định tuyến traffic vào node này.

---

## 3. Cấu hình mẫu trên Kubernetes Deployment

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health/live
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /api/v1/health/ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

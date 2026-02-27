# Casheer Menu Service

Menu and Order management service for Casheer POS system.

## Features

- 📋 Menu management (CRUD)
- 🖼️ Menu images upload to Cloudinary
- 🔄 Menu variations (size, level, etc.)
- 📦 Stock management
- 🧾 Order processing
- 📊 Best seller analytics
- 🔌 RabbitMQ integration for events

## Tech Stack

- Go 1.21
- Fiber v2
- GORM
- PostgreSQL
- Cloudinary
- RabbitMQ
- JWT

## API Endpoints

### Menu Endpoints
- `POST /api/v1/menus` - Create menu
- `GET /api/v1/menus` - Get all menus
- `GET /api/v1/menus/bestseller` - Get best seller
- `GET /api/v1/menus/:id` - Get menu by ID
- `PUT /api/v1/menus/:id` - Update menu
- `DELETE /api/v1/menus/:id` - Delete menu
- `POST /api/v1/menus/:id/images` - Upload image
- `DELETE /api/v1/menus/:id/images/:imageId` - Delete image

### Variation Endpoints
- `POST /api/v1/variations` - Create variation
- `PUT /api/v1/variations/:id` - Update variation
- `DELETE /api/v1/variations/:id` - Delete variation
- `PATCH /api/v1/variations/:id/stock` - Update stock

### Order Endpoints
- `POST /api/v1/orders` - Create order
- `GET /api/v1/orders` - Get all orders
- `GET /api/v1/orders/:id` - Get order by ID
- `GET /api/v1/orders/tenant/:tenantId` - Get orders by tenant

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| DB_HOST | Database host | localhost |
| DB_PORT | Database port | 5432 |
| DB_USER | Database user | postgres |
| DB_PASSWORD | Database password | postgres |
| DB_NAME | Database name | casheer_db |
| PORT | Service port | 3002 |
| JWT_SECRET | JWT secret key | - |
| CLOUDINARY_CLOUD_NAME | Cloudinary cloud name | - |
| CLOUDINARY_API_KEY | Cloudinary API key | - |
| CLOUDINARY_API_SECRET | Cloudinary API secret | - |
| RABBITMQ_URL | RabbitMQ connection URL | amqp://localhost:5672 |

## Running the Service

```bash
go mod tidy
go run cmd/main.go
```

### **23. `Tiltfile`**
```python
# Tiltfile for Casheer Menu Service

print("🚀 Starting Casheer Menu Service...")

# Build configuration
docker_build(
    'casheer-menu-service',
    '.',
    dockerfile='Dockerfile',
    build_args={
        'GO_VERSION': '1.21'
    }
)

# Kubernetes deployment
k8s_yaml('k8s/deployment.yaml')

# Port forwarding for local development
k8s_resource(
    'casheer-menu-service',
    port_forwards='3002:3002',
    labels=['menu-service']
)

# Watch for changes
watch_file('internal/**/*.go')
watch_file('cmd/**/*.go')
watch_file('pkg/**/*.go')
watch_file('.env')

print("✅ Menu Service configuration loaded")
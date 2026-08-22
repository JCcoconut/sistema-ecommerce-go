# Referencia de la API REST

Base local: `http://localhost:8080`. Todas las respuestas de la API utilizan `application/json`.

## Catálogo e inventario

| Método | Ruta | Éxito | Descripción |
|---|---|---:|---|
| GET | `/api/health` | 200 | Resumen de salud y conteos |
| GET | `/api/productos?q=audio` | 200 | Lista o filtra productos |
| POST | `/api/productos` | 201 | Crea un producto |
| GET | `/api/productos/{id}` | 200 | Obtiene un producto |
| PUT | `/api/productos/{id}` | 200 | Actualiza datos y stock |
| DELETE | `/api/productos/{id}` | 204 | Elimina si no está en el carrito |
| GET | `/api/productos/bajo-stock?limite=5` | 200 | Filtra por umbral de stock |

```bash
curl http://localhost:8080/api/productos

curl -X POST http://localhost:8080/api/productos \
  -H 'Content-Type: application/json' \
  -d '{
    "id":"P100","nombre":"Webcam 4K","categoria":"Video",
    "precio":99.90,"stock":12,
    "proveedor_nombre":"Digital Supply","proveedor_email":"ventas@digital.ec"
  }'
```

## Carrito y promociones

| Método | Ruta | Éxito | Descripción |
|---|---|---:|---|
| GET | `/api/cupones` | 200 | Lista códigos disponibles |
| GET | `/api/carrito` | 200 | Ítems y totales |
| POST | `/api/carrito/items` | 201 | Agrega/acumula un producto |
| PUT | `/api/carrito/items/{id}` | 200 | Reemplaza la cantidad |
| DELETE | `/api/carrito/items/{id}` | 200 | Retira la línea |
| POST | `/api/carrito/cupon` | 200 | Aplica una estrategia de descuento |

```bash
curl -X POST http://localhost:8080/api/carrito/items \
  -H 'Content-Type: application/json' \
  -d '{"producto_id":"P001","cantidad":2}'

curl -X POST http://localhost:8080/api/carrito/cupon \
  -H 'Content-Type: application/json' \
  -d '{"codigo":"GO15"}'
```

## Clientes, pedidos y facturas

| Método | Ruta | Éxito | Descripción |
|---|---|---:|---|
| GET / POST | `/api/clientes` | 200/201 | Lista o registra clientes |
| GET | `/api/clientes/{email}/pedidos` | 200 | Historial de un cliente |
| GET / POST | `/api/pedidos` | 200/201 | Lista o confirma una compra |
| GET | `/api/pedidos/{id}` | 200 | Consulta un pedido |
| PUT | `/api/pedidos/{id}/estado` | 200 | Avanza el flujo del pedido |
| POST | `/api/pedidos/{id}/cancelar` | 200 | Cancela y repone existencias |
| GET | `/api/facturas` | 200 | Lista comprobantes simulados |
| GET | `/api/facturas/{id}` | 200 | Consulta un comprobante simulado |

```bash
curl -X POST http://localhost:8080/api/pedidos \
  -H 'Content-Type: application/json' \
  -d '{
    "nombre":"Jorge Cano","email":"jorge@correo.ec",
    "ciudad":"Quito","sector":"Norte","referencia":"Casa azul"
  }'

curl -X PUT http://localhost:8080/api/pedidos/PED-001/estado \
  -H 'Content-Type: application/json' \
  -d '{"estado":"Enviado"}'
```

## Errores

Formato uniforme:

```json
{
  "error": {
    "codigo": "no_encontrado",
    "mensaje": "producto no encontrado"
  }
}
```

| Estado | Uso |
|---:|---|
| 400 | JSON o datos inválidos |
| 404 | Recurso inexistente |
| 409 | Duplicado, stock insuficiente o transición inválida |
| 500 | Error interno o de persistencia |

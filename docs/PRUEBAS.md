# Plan y resultados de pruebas

## Estrategia

Se aplica una pirámide sencilla: muchas pruebas unitarias para reglas pequeñas, pruebas de integración para la coordinación entre paquetes y una prueba de aceptación que recorre el flujo principal por HTTP.

## Comandos

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

Para generar cobertura:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Matriz de pruebas

| ID | Tipo | Caso | Resultado esperado | Estado |
|---|---|---|---|---|
| U-01 | Unitaria | Crear producto válido | Objeto con ID normalizado | Aprobado |
| U-02 | Unitaria | Precio o stock inválido | Retorna error | Aprobado |
| U-03 | Unitaria | Agregar dos veces al carrito | Una línea y cantidad acumulada | Aprobado |
| U-04 | Unitaria | Stock insuficiente | Error sin modificar cantidad | Aprobado |
| U-05 | Table-driven | Transiciones de pedido | Solo flujo autorizado | Aprobado |
| U-06 | Unitaria | Factura con total 115 | Base 100, IVA 15 | Aprobado |
| P-01 | Persistencia | Guardar/cargar estado | Datos equivalentes | Aprobado |
| P-02 | Seguridad | Campo JSON desconocido | Error de decodificación | Aprobado |
| P-03 | Seguridad | Permisos del archivo | Sin permisos para grupo/otros | Aprobado |
| I-01 | Integración | Reiniciar Tienda | Carrito restaurado desde JSON | Aprobado |
| I-02 | Integración HTTP | Recurso inexistente | Estado 404 y error JSON | Aprobado |
| A-01 | Aceptación | Carrito + cupón + checkout | Pedido y factura creados | Aprobado |
| C-01 | Concurrencia | 30 POST simultáneos | 30 unidades, cero carreras | Aprobado |

## Evidencia de la ejecución

Comando ejecutado:

```text
go test -race ./...
```

Resultado resumido:

```text
ok  internal/aplicacion
ok  internal/carrito
ok  internal/catalogo
ok  internal/facturacion
ok  internal/modelo
ok  internal/pedido
ok  internal/persistencia
ok  internal/promocion
ok  internal/webapi
```

Cobertura global obtenida con `go test -coverprofile=coverage.out ./...`: **42,5 % de sentencias**. Este porcentaje incluye los dos puntos de entrada `cmd/`, que deliberadamente se verifican mediante integración y ejecución manual; el objetivo del proyecto es demostrar variedad y calidad de casos, no alcanzar una métrica artificial.

## Errores detectados durante el desarrollo

| Problema | Causa | Corrección |
|---|---|---|
| Posible pérdida de cantidades concurrentes | Varias goroutines modificaban el mismo map | Operaciones de escritura bajo `sync.RWMutex` |
| Estados de pedido arbitrarios | Método aceptaba cualquier estado conocido | Tabla explícita de transiciones permitidas |
| JSON podía aceptar errores tipográficos | Decoder permisivo | `DisallowUnknownFields` y documento único |
| Archivo podía quedar parcial | Escritura directa | Archivo temporal + `Sync` + `Rename` |
| Eliminación de producto usado | Referencia desde carrito | Respuesta 409 y conservación del recurso |

## Prueba manual de aceptación

1. Ejecutar `go run ./cmd/web`.
2. Abrir `http://localhost:8080`.
3. Agregar dos productos.
4. Aplicar `GO15`.
5. Completar los datos del cliente.
6. Confirmar el pedido.
7. Verificar el pedido en el historial y la factura simulada en pantalla.
8. Detener y reiniciar el servidor.
9. Confirmar que pedidos, stock y factura se conservan.

# Sistema de gestión de e-commerce en Go

Avance significativo del proyecto académico correspondiente a la Unidad 3. El programa funciona en consola y modela el catálogo, carrito, promociones, pedidos, clientes, direcciones, proveedores e inventario de **AudioCyber Store**.

## Funcionalidades demostrables

- Listar y buscar productos por nombre, categoría o identificador.
- Mostrar una vista rápida construida con una subslice.
- Agregar, acumular y retirar productos del carrito.
- Calcular cantidades, subtotales, descuentos y total de compra.
- Aplicar cupones mediante distintas implementaciones de una interfaz.
- Confirmar pedidos, validar existencias y descontar inventario.
- Consultar el historial de pedidos creados durante la ejecución.
- Reponer stock mediante un método con receptor por puntero.
- Informar entradas inválidas y reglas de negocio incumplidas sin detener abruptamente el programa.

## Ejecución

Requisitos: Go 1.22 o una versión posterior compatible.

```bash
go run ./cmd/tienda
```

## Verificación

```bash
gofmt -w .
go test ./...
go vet ./...
go run ./cmd/tienda
```

## Estructura

```text
cmd/tienda/             Aplicación de consola y menús
internal/modelo/        Producto, proveedor, cliente y dirección
internal/catalogo/      Catálogo, búsquedas, map, slice y subslice
internal/carrito/       Ítems y operaciones del carrito
internal/promocion/     Interfaz Descuento y gestión de cupones
internal/pedido/        Interfaz FuenteCarrito y creación de pedidos
```

## Dependencias

El proyecto no usa paquetes de terceros. Solo importa paquetes de la biblioteca estándar de Go: `bufio`, `errors`, `fmt`, `os`, `sort`, `strconv`, `strings`, `testing` y `time`.

La explicación detallada del avance está en [AVANCE_UNIDAD_3.md](AVANCE_UNIDAD_3.md).

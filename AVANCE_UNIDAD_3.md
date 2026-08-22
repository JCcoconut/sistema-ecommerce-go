# Evidencia histórica del avance — Unidad 3

> Este archivo conserva el estado del proyecto antes de incorporar la API REST, persistencia, concurrencia y pruebas finales. Para la entrega vigente consulte [README.md](README.md) e [INFORME_FINAL.md](INFORME_FINAL.md).

## Objetivo del avance

Construir una versión funcional en consola del sistema de gestión de e-commerce, aplicando las estructuras y mecanismos de orientación a objetos disponibles en Go: composición de `structs`, encapsulación por paquetes, constructores, métodos, interfaces y manejo explícito de errores.

## Alcance implementado

| Módulo | Responsabilidad | Funcionalidades terminadas |
|---|---|---|
| Modelo | Representar objetos del negocio | Productos, proveedores, clientes y direcciones validados y encapsulados |
| Catálogo | Gestionar productos disponibles | Alta inicial, listado, búsqueda por ID, búsqueda textual y vista mediante subslice |
| Carrito | Gestionar una compra temporal | Agregar, acumular, eliminar, vaciar, contar unidades y calcular subtotal |
| Promociones | Aplicar reglas intercambiables | Sin descuento, porcentaje y búsqueda de cupones en un `map` |
| Pedidos | Confirmar la compra | Validación en dos fases, descuento de stock, líneas inmutables e historial en memoria |
| Consola | Permitir la demostración | Menú, lectura de datos y mensajes de error comprensibles |

## Relación entre requisitos y código

| Tema de la Unidad 3 | Implementación | Ubicación principal |
|---|---|---|
| `structs` | `Producto`, `Proveedor`, `Cliente`, `Item`, `Pedido` | `internal/*` |
| `structs` anidados | `Cliente` contiene `DireccionEntrega`; `Producto` contiene `Proveedor`; `Pedido` contiene `Cliente` | `internal/modelo`, `internal/pedido` |
| Encapsulación | Campos en minúscula, constructores y métodos de consulta/modificación | Todos los paquetes internos |
| Receptor por valor | Métodos de consulta como `Nombre`, `Stock`, `Subtotal`, `Estado` | `internal/modelo`, `internal/pedido` |
| Receptor por puntero | `DescontarStock`, `ReponerStock`, `Agregar`, `Vaciar`, `CambiarEstado` | `internal/modelo`, `internal/carrito`, `internal/pedido` |
| Constructores | `NuevoProducto`, `NuevoCliente`, `NuevoCatalogo`, `NuevoPedido` | Todos los módulos |
| Interfaces | `Buscador`, `Descuento` y `FuenteCarrito` | Catálogo, promoción y pedido |
| Manejo de errores | Validaciones, errores centinela, `errors.Is` y envoltura con `%w` | Todos los módulos y `main` |
| `map` | Productos del catálogo, ítems del carrito, cupones y estados válidos | Catálogo, carrito, promoción y pedido |
| Array | Códigos y porcentajes fijos de cupones | `internal/promocion/descuento.go` |
| Slice | Orden de catálogo/carrito, líneas e historial | Catálogo, carrito, pedido y `main` |
| Subslice | Selección de los primeros productos con copia defensiva | `Catalogo.Primeros` |
| Función anónima y closure | Generación consecutiva de identificadores de pedido | `pedido.NuevoGeneradorID` |
| Comentarios | Explicación de validación en dos fases, interfaces y copias defensivas | Funciones con lógica menos evidente |

## Manejo de errores

Los constructores impiden que se creen objetos con precios negativos, identificadores vacíos, correos inválidos o direcciones incompletas. Las operaciones rechazan cantidades no positivas, stock insuficiente, carritos vacíos, cupones inexistentes y estados de pedido desconocidos. El programa principal presenta el problema al usuario y continúa funcionando.

Se utilizan errores centinela cuando el llamador necesita distinguir una causa concreta, por ejemplo `ErrProductoNoEncontrado`, `ErrItemNoEncontrado` y `ErrCuponNoEncontrado`. Los errores que necesitan contexto se envuelven con `fmt.Errorf` y `%w`.

## Decisión importante: creación del pedido en dos fases

`NuevoPedido` primero valida todos los ítems, calcula el subtotal y aplica el descuento. Solo después crea las líneas y descuenta el inventario. Con esto se evita modificar parte del stock si alguna línea presenta un problema antes de completar la confirmación.

## Pruebas incluidas

Las pruebas automatizadas verifican:

- constructores y encapsulación de objetos;
- cambios mediante receptores por puntero;
- búsqueda por `map` y creación de subslices;
- acumulación y errores del carrito;
- implementaciones de la interfaz `Descuento`;
- creación del pedido, descuento de stock y cambio de estado;
- funcionamiento del closure que genera identificadores.

## Limitaciones del avance

- Los datos se mantienen únicamente durante la ejecución; todavía no existe base de datos.
- No se implementan usuarios, autenticación, pagos reales, envíos ni interfaz web.
- El inventario no contempla accesos concurrentes.
- Los pedidos no se guardan al cerrar el programa.

Estas limitaciones delimitan el avance y dejan una ruta clara para las siguientes unidades.

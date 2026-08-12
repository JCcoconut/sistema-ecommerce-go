// Tienda ejecuta la demostración del avance correspondiente a la Unidad 3.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/catalogo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/promocion"
)

func main() {
	lector := bufio.NewReader(os.Stdin)
	catalogoTienda, err := catalogo.CrearInicial()
	if err != nil {
		fmt.Println("No se pudo iniciar el catálogo:", err)
		return
	}
	gestorCupones, err := promocion.NuevoGestorCupones()
	if err != nil {
		fmt.Println("No se pudieron iniciar los cupones:", err)
		return
	}

	carro := carrito.NuevoCarrito()
	descuentoActivo := promocion.Descuento(promocion.SinDescuento{})
	generarID := pedido.NuevoGeneradorID()
	pedidos := make([]*pedido.Pedido, 0)

	mostrarEncabezado()
	for {
		mostrarMenu()
		opcion := leerTexto(lector, "Seleccione una opción: ")

		switch opcion {
		case "1":
			mostrarProductos(catalogoTienda.Listar())
		case "2":
			texto := leerTexto(lector, "Nombre o categoría: ")
			mostrarResultadoBusqueda(catalogoTienda, texto)
		case "3":
			fmt.Println("Vista rápida mediante subslice:")
			mostrarProductos(catalogoTienda.Primeros(3))
		case "4":
			agregarAlCarrito(lector, catalogoTienda, carro)
		case "5":
			id := leerTexto(lector, "ID del producto a retirar: ")
			if err := carro.Eliminar(id); err != nil {
				fmt.Println("No se pudo retirar:", err)
				continue
			}
			fmt.Println("Producto retirado correctamente.")
		case "6":
			mostrarCarrito(carro, descuentoActivo)
		case "7":
			fmt.Println("Cupones disponibles:", strings.Join(gestorCupones.Codigos(), ", "))
			codigo := leerTexto(lector, "Código: ")
			nuevoDescuento, err := gestorCupones.Buscar(codigo)
			if err != nil {
				fmt.Println("No se pudo aplicar:", err)
				continue
			}
			descuentoActivo = nuevoDescuento
			fmt.Println("Descuento aplicado:", descuentoActivo.Nombre())
		case "8":
			nuevoPedido, err := confirmarPedido(
				lector, generarID(), carro, descuentoActivo,
			)
			if err != nil {
				fmt.Println("No se pudo confirmar el pedido:", err)
				continue
			}
			pedidos = append(pedidos, nuevoPedido)
			carro.Vaciar()
			descuentoActivo = promocion.SinDescuento{}
			fmt.Printf("Pedido %s confirmado. Total: $%.2f\n", nuevoPedido.ID(), nuevoPedido.Total())
		case "9":
			mostrarPedidos(pedidos)
		case "10":
			reponerInventario(lector, catalogoTienda)
		case "0":
			fmt.Println("Gracias por utilizar AudioCyber Store.")
			return
		default:
			fmt.Println("Opción no válida.")
		}
	}
}

func mostrarEncabezado() {
	fmt.Println("================================================")
	fmt.Println("      AUDIOCYBER STORE - AVANCE UNIDAD 3")
	fmt.Println("================================================")
}

func mostrarMenu() {
	fmt.Println("\n1. Listar catálogo")
	fmt.Println("2. Buscar productos mediante interfaz")
	fmt.Println("3. Mostrar los primeros tres productos (subslice)")
	fmt.Println("4. Agregar producto al carrito")
	fmt.Println("5. Retirar producto del carrito")
	fmt.Println("6. Ver carrito")
	fmt.Println("7. Aplicar cupón")
	fmt.Println("8. Confirmar pedido")
	fmt.Println("9. Ver historial de pedidos")
	fmt.Println("10. Reponer inventario")
	fmt.Println("0. Salir")
}

// mostrarResultadoBusqueda recibe la interfaz Buscador, no el tipo Catalogo.
func mostrarResultadoBusqueda(buscador catalogo.Buscador, texto string) {
	mostrarProductos(buscador.Buscar(texto))
}

func agregarAlCarrito(lector *bufio.Reader, catalogoTienda *catalogo.Catalogo, carro *carrito.Carrito) {
	id := leerTexto(lector, "ID del producto: ")
	producto, err := catalogoTienda.BuscarPorID(id)
	if err != nil {
		if errors.Is(err, catalogo.ErrProductoNoEncontrado) {
			fmt.Println("El producto solicitado no existe.")
			return
		}
		fmt.Println("Error al buscar:", err)
		return
	}

	cantidad, err := leerEntero(lector, "Cantidad: ")
	if err != nil {
		fmt.Println("Cantidad inválida:", err)
		return
	}
	if err := carro.Agregar(producto, cantidad); err != nil {
		fmt.Println("No se pudo agregar:", err)
		return
	}
	fmt.Println("Producto agregado correctamente.")
}

func confirmarPedido(
	lector *bufio.Reader,
	id string,
	carro *carrito.Carrito,
	descuento promocion.Descuento,
) (*pedido.Pedido, error) {
	nombre := leerTexto(lector, "Nombre del cliente: ")
	email := leerTexto(lector, "Correo: ")
	ciudad := leerTexto(lector, "Ciudad: ")
	sector := leerTexto(lector, "Sector: ")
	referencia := leerTexto(lector, "Referencia de entrega: ")

	direccion, err := modelo.NuevaDireccionEntrega(ciudad, sector, referencia)
	if err != nil {
		return nil, fmt.Errorf("crear dirección: %w", err)
	}
	cliente, err := modelo.NuevoCliente(nombre, email, direccion)
	if err != nil {
		return nil, fmt.Errorf("crear cliente: %w", err)
	}

	return pedido.NuevoPedido(id, time.Now(), cliente, carro, descuento)
}

func reponerInventario(lector *bufio.Reader, catalogoTienda *catalogo.Catalogo) {
	id := leerTexto(lector, "ID del producto: ")
	producto, err := catalogoTienda.BuscarPorID(id)
	if err != nil {
		fmt.Println("No se encontró el producto:", err)
		return
	}
	cantidad, err := leerEntero(lector, "Unidades a reponer: ")
	if err != nil {
		fmt.Println("Cantidad inválida:", err)
		return
	}
	if err := producto.ReponerStock(cantidad); err != nil {
		fmt.Println("No se pudo reponer:", err)
		return
	}
	fmt.Printf("Stock actualizado de %s: %d unidades.\n", producto.Nombre(), producto.Stock())
}

func leerTexto(lector *bufio.Reader, mensaje string) string {
	fmt.Print(mensaje)
	texto, _ := lector.ReadString('\n')
	return strings.TrimSpace(texto)
}

func leerEntero(lector *bufio.Reader, mensaje string) (int, error) {
	texto := leerTexto(lector, mensaje)
	valor, err := strconv.Atoi(texto)
	if err != nil {
		return 0, fmt.Errorf("%q no es un número entero", texto)
	}
	return valor, nil
}

func mostrarProductos(productos []modelo.Producto) {
	if len(productos) == 0 {
		fmt.Println("No se encontraron productos.")
		return
	}

	fmt.Printf("\n%-6s %-24s %-13s %9s %7s %-18s\n",
		"ID", "PRODUCTO", "CATEGORÍA", "PRECIO", "STOCK", "PROVEEDOR")
	fmt.Println(strings.Repeat("-", 91))
	for _, producto := range productos {
		fmt.Printf("%-6s %-24s %-13s $%8.2f %7d %-18s\n",
			producto.ID(), producto.Nombre(), producto.Categoria(),
			producto.Precio(), producto.Stock(), producto.Proveedor().Nombre())
	}
}

func mostrarCarrito(carro *carrito.Carrito, descuento promocion.Descuento) {
	if carro.Vacio() {
		fmt.Println("El carrito está vacío.")
		return
	}

	fmt.Printf("\n%-6s %-25s %8s %12s\n", "ID", "PRODUCTO", "CANT.", "TOTAL")
	fmt.Println(strings.Repeat("-", 57))
	for _, item := range carro.Items() {
		subtotal, err := item.Subtotal()
		if err != nil {
			fmt.Println("Error en una línea:", err)
			continue
		}
		fmt.Printf("%-6s %-25s %8d $%11.2f\n",
			item.Producto().ID(), item.Producto().Nombre(), item.Cantidad(), subtotal)
	}

	subtotal, err := carro.Subtotal()
	if err != nil {
		fmt.Println("No se pudo calcular el carrito:", err)
		return
	}
	total, err := descuento.Aplicar(subtotal)
	if err != nil {
		fmt.Println("No se pudo aplicar el descuento:", err)
		return
	}
	fmt.Printf("Unidades: %d | Promoción: %s\n", carro.CantidadTotal(), descuento.Nombre())
	fmt.Printf("Subtotal: $%.2f | Descuento: $%.2f | Total: $%.2f\n",
		subtotal, subtotal-total, total)
}

func mostrarPedidos(pedidos []*pedido.Pedido) {
	if len(pedidos) == 0 {
		fmt.Println("Todavía no existen pedidos confirmados.")
		return
	}

	for _, orden := range pedidos {
		direccion := orden.Cliente().Direccion()
		fmt.Printf("%s | %s | %s | %s, %s | %d líneas | $%.2f | %s\n",
			orden.ID(), orden.Fecha().Format("02/01/2006 15:04"),
			orden.Cliente().Nombre(), direccion.Ciudad(), direccion.Sector(),
			len(orden.Lineas()), orden.Total(), orden.Estado())
	}
}

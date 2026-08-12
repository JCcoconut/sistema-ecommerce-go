package pedido

import (
	"testing"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/promocion"
)

func datosPedido(t *testing.T) (*modelo.Producto, modelo.Cliente) {
	t.Helper()
	proveedor, err := modelo.NuevoProveedor("Proveedor", "ventas@proveedor.ec")
	if err != nil {
		t.Fatal(err)
	}
	producto, err := modelo.NuevoProducto("P001", "Audífonos", "Audio", 10, 5, proveedor)
	if err != nil {
		t.Fatal(err)
	}
	direccion, err := modelo.NuevaDireccionEntrega("Quito", "Norte", "Casa azul")
	if err != nil {
		t.Fatal(err)
	}
	cliente, err := modelo.NuevoCliente("Jorge", "jorge@correo.com", direccion)
	if err != nil {
		t.Fatal(err)
	}
	return producto, cliente
}

func TestNuevoPedidoDescuentaStockYAplicaInterfaz(t *testing.T) {
	producto, cliente := datosPedido(t)
	carro := carrito.NuevoCarrito()
	if err := carro.Agregar(producto, 2); err != nil {
		t.Fatal(err)
	}
	regla, err := promocion.NuevoPorcentaje("DIEZ", 0.10)
	if err != nil {
		t.Fatal(err)
	}

	orden, err := NuevoPedido("PED-001", time.Now(), cliente, carro, regla)
	if err != nil {
		t.Fatal(err)
	}
	if orden.Total() != 18 || orden.Descuento() != 2 {
		t.Fatalf("totales inesperados: total %.2f, descuento %.2f", orden.Total(), orden.Descuento())
	}
	if producto.Stock() != 3 {
		t.Fatalf("se esperaba stock 3; se obtuvo %d", producto.Stock())
	}
	if orden.Cliente().Direccion().Ciudad() != "Quito" {
		t.Fatal("no se conservó el struct anidado del cliente")
	}
	if err := orden.CambiarEstado(EstadoEnviado); err != nil {
		t.Fatal(err)
	}
	if orden.Estado() != EstadoEnviado {
		t.Fatal("el receptor por puntero no actualizó el estado")
	}
}

func TestPedidoRechazaCarritoVacio(t *testing.T) {
	_, cliente := datosPedido(t)
	_, err := NuevoPedido("PED-001", time.Now(), cliente, carrito.NuevoCarrito(), nil)
	if err == nil {
		t.Fatal("se esperaba un error por carrito vacío")
	}
}

func TestClosureGeneraIDsConsecutivos(t *testing.T) {
	generar := NuevoGeneradorID()
	if generar() != "PED-001" || generar() != "PED-002" {
		t.Fatal("el closure no conservó correctamente el contador")
	}
}

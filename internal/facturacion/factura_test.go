package facturacion

import (
	"strings"
	"testing"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
)

func TestEmisorSimuladoDesglosaIVA(t *testing.T) {
	proveedor, _ := modelo.NuevoProveedor("Proveedor", "ventas@proveedor.ec")
	producto, _ := modelo.NuevoProducto("P001", "Producto", "Prueba", 115, 2, proveedor)
	direccion, _ := modelo.NuevaDireccionEntrega("Quito", "Norte", "")
	cliente, _ := modelo.NuevoCliente("Ana", "ana@correo.ec", direccion)
	carro := carrito.NuevoCarrito()
	if err := carro.Agregar(producto, 1); err != nil {
		t.Fatal(err)
	}
	orden, err := pedido.NuevoPedido("PED-001", time.Now(), cliente, carro, nil)
	if err != nil {
		t.Fatal(err)
	}
	factura, err := (EmisorSimulado{}).Emitir(*orden)
	if err != nil {
		t.Fatal(err)
	}
	if factura.BaseImponible() != 100 || factura.IVA() != 15 || factura.Total() != 115 {
		t.Fatalf("desglose inesperado: base %.2f, IVA %.2f, total %.2f", factura.BaseImponible(), factura.IVA(), factura.Total())
	}
	if !strings.Contains(factura.Advertencia(), "SIN VALIDEZ TRIBUTARIA") {
		t.Fatal("falta la advertencia legal")
	}
}

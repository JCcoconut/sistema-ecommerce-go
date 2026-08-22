package aplicacion

import (
	"path/filepath"
	"testing"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/facturacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/persistencia"
)

func TestTiendaPersisteYRestauraCarrito(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	repo, err := persistencia.NuevoArchivoJSON(ruta)
	if err != nil {
		t.Fatal(err)
	}
	primera, err := NuevaTienda(repo, facturacion.EmisorSimulado{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := primera.AgregarAlCarrito("P001", 2); err != nil {
		t.Fatal(err)
	}
	segunda, err := NuevaTienda(repo, facturacion.EmisorSimulado{})
	if err != nil {
		t.Fatal(err)
	}
	carro, err := segunda.VerCarrito()
	if err != nil {
		t.Fatal(err)
	}
	if carro.CantidadTotal != 2 || len(carro.Items) != 1 {
		t.Fatalf("carrito no restaurado: %#v", carro)
	}
}

func TestCompraCierraMVP(t *testing.T) {
	tienda, err := NuevaTienda(nil, facturacion.EmisorSimulado{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tienda.AgregarAlCarrito("P002", 2); err != nil {
		t.Fatal(err)
	}
	if _, err := tienda.AplicarCupon("GO15"); err != nil {
		t.Fatal(err)
	}
	compra, err := tienda.ConfirmarCompra(ClienteEntrada{Nombre: "Jorge", Email: "jorge@correo.ec", Ciudad: "Quito", Sector: "Norte", Referencia: "Edificio azul"})
	if err != nil {
		t.Fatal(err)
	}
	if compra.Pedido.ID != "PED-001" || compra.Pedido.Total <= 0 {
		t.Fatalf("pedido inesperado: %#v", compra.Pedido)
	}
	if compra.Factura.PedidoID != compra.Pedido.ID || compra.Factura.Advertencia == "" {
		t.Fatalf("factura inesperada: %#v", compra.Factura)
	}
	carro, _ := tienda.VerCarrito()
	if carro.CantidadTotal != 0 {
		t.Fatal("el carrito debe vaciarse tras la compra")
	}
}

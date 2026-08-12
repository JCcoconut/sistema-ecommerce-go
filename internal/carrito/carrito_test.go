package carrito

import (
	"errors"
	"testing"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
)

func productoPrueba(t *testing.T) *modelo.Producto {
	t.Helper() // marca esta función como un helper para que los errores se reporten en la línea de la prueba que la llamó
	proveedor, err := modelo.NuevoProveedor("Proveedor", "ventas@proveedor.ec")
	if err != nil {
		t.Fatal(err)
	}
	producto, err := modelo.NuevoProducto("P001", "Teclado", "Periféricos", 25, 5, proveedor)
	if err != nil {
		t.Fatal(err)
	}
	return producto
}

func TestCarritoAcumulaItemsEnMap(t *testing.T) {
	producto := productoPrueba(t)
	carro := NuevoCarrito()
	if err := carro.Agregar(producto, 2); err != nil {
		t.Fatal(err)
	}
	if err := carro.Agregar(producto, 1); err != nil {
		t.Fatal(err)
	}

	if len(carro.Items()) != 1 || carro.CantidadTotal() != 3 {
		t.Fatalf("carrito inesperado: %d líneas, %d unidades", len(carro.Items()), carro.CantidadTotal())
	}
	subtotal, err := carro.Subtotal()
	if err != nil {
		t.Fatal(err)
	}
	if subtotal != 75 {
		t.Fatalf("se esperaba 75; se obtuvo %.2f", subtotal)
	}
}

func TestCarritoManejaErrores(t *testing.T) {
	carro := NuevoCarrito()
	if err := carro.Agregar(productoPrueba(t), 99); err == nil {
		t.Fatal("se esperaba error por stock insuficiente")
	}
	if err := carro.Eliminar("P999"); !errors.Is(err, ErrItemNoEncontrado) {
		t.Fatalf("se esperaba ErrItemNoEncontrado; se obtuvo %v", err)
	}
}

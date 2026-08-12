package catalogo

import (
	"errors"
	"testing"
)

func TestCatalogoUsaMapBusquedaYSusblice(t *testing.T) {
	catalogo, err := CrearInicial()
	if err != nil {
		t.Fatal(err)
	}

	producto, err := catalogo.BuscarPorID("p001")
	if err != nil {
		t.Fatal(err)
	}
	if producto.ID() != "P001" {
		t.Fatalf("ID inesperado: %s", producto.ID())
	}
	if len(catalogo.Buscar("audio")) != 2 {
		t.Fatal("se esperaban dos productos de audio")
	}
	if len(catalogo.Primeros(3)) != 3 {
		t.Fatal("la subslice debe contener tres productos")
	}
	if len(catalogo.Primeros(100)) != catalogo.Cantidad() {
		t.Fatal("una cantidad mayor debe limitarse al tamaño del catálogo")
	}
}

func TestBuscarProductoInexistente(t *testing.T) {
	catalogo, err := CrearInicial()
	if err != nil {
		t.Fatal(err)
	}
	_, err = catalogo.BuscarPorID("NO-EXISTE")
	if !errors.Is(err, ErrProductoNoEncontrado) {
		t.Fatalf("se esperaba ErrProductoNoEncontrado; se obtuvo %v", err)
	}
}

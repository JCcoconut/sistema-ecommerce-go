package modelo

import "testing"

func TestProductoEncapsuladoYReceptores(t *testing.T) {
	proveedor, err := NuevoProveedor("AudioTech", "ventas@audio.ec")
	if err != nil {
		t.Fatal(err)
	}
	producto, err := NuevoProducto("p001", "Audífonos", "Audio", 50, 5, proveedor)
	if err != nil {
		t.Fatal(err)
	}

	if producto.ID() != "P001" || producto.Stock() != 5 {
		t.Fatalf("producto inesperado: %s, stock %d", producto.ID(), producto.Stock())
	}
	if err := producto.DescontarStock(2); err != nil {
		t.Fatal(err)
	}
	if err := producto.ReponerStock(1); err != nil {
		t.Fatal(err)
	}
	if producto.Stock() != 4 {
		t.Fatalf("se esperaba stock 4; se obtuvo %d", producto.Stock())
	}
}

func TestConstructoresRechazanDatosInvalidos(t *testing.T) {
	if _, err := NuevoProveedor("", "correo@ejemplo.com"); err == nil {
		t.Fatal("se esperaba un error por proveedor sin nombre")
	}
	proveedor, _ := NuevoProveedor("Proveedor", "correo@ejemplo.com")
	if _, err := NuevoProducto("P001", "Producto", "Audio", -1, 3, proveedor); err == nil {
		t.Fatal("se esperaba un error por precio inválido")
	}
}

func TestClienteContieneDireccionAnidada(t *testing.T) {
	direccion, err := NuevaDireccionEntrega("Quito", "Centro", "Frente al parque")
	if err != nil {
		t.Fatal(err)
	}
	cliente, err := NuevoCliente("Ana", "ana@correo.com", direccion)
	if err != nil {
		t.Fatal(err)
	}
	if cliente.Direccion().Ciudad() != "Quito" {
		t.Fatalf("ciudad inesperada: %s", cliente.Direccion().Ciudad())
	}
}

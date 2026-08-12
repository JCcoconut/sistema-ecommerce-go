package catalogo

import (
	"fmt"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
)

// CrearInicial centraliza la construcción explícita de los objetos de demostración.
func CrearInicial() (*Catalogo, error) {
	proveedorAudio, err := modelo.NuevoProveedor("AudioTech Ecuador", "ventas@audiotech.ec")
	if err != nil {
		return nil, fmt.Errorf("crear proveedor de audio: %w", err)
	}
	proveedorDigital, err := modelo.NuevoProveedor("Digital Supply", "contacto@digitalsupply.ec")
	if err != nil {
		return nil, fmt.Errorf("crear proveedor digital: %w", err)
	}

	datos := []struct {
		id, nombre, categoria string
		precio                float64
		stock                 int
		proveedor             modelo.Proveedor
	}{
		{"P001", "Teclado mecánico", "Periféricos", 65.90, 16, proveedorDigital},
		{"P002", "Mouse inalámbrico", "Periféricos", 28.50, 22, proveedorDigital},
		{"P003", "Audífonos de estudio", "Audio", 89.99, 12, proveedorAudio},
		{"P004", "Micrófono USB", "Audio", 74.25, 7, proveedorAudio},
		{"P005", "Hub USB-C", "Accesorios", 39.80, 24, proveedorDigital},
		{"P006", "Soporte para laptop", "Accesorios", 32.00, 14, proveedorDigital},
	}

	productos := make([]*modelo.Producto, 0, len(datos))
	for _, dato := range datos {
		producto, err := modelo.NuevoProducto(
			dato.id, dato.nombre, dato.categoria,
			dato.precio, dato.stock, dato.proveedor,
		)
		if err != nil {
			return nil, fmt.Errorf("crear producto %s: %w", dato.id, err)
		}
		productos = append(productos, producto)
	}

	return NuevoCatalogo(productos)
}

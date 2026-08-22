// Package catalogo administra los productos utilizando un map y un slice.
package catalogo

import (
	"errors"
	"strings"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
)

var (
	// ErrProductoNoEncontrado permite identificar este error con errors.Is.
	ErrProductoNoEncontrado   = errors.New("producto no encontrado")
	ErrProductoDuplicado      = errors.New("el producto ya existe")
	ErrCatalogoNoInicializado = errors.New("el catálogo no está inicializado")
)

// Buscador define el comportamiento mínimo de un catálogo consultable.
// Main utiliza esta interfaz en lugar de depender de una implementación concreta.
type Buscador interface {
	BuscarPorID(id string) (*modelo.Producto, error)
	Buscar(texto string) []modelo.Producto
}

// Verificación en compilación: Catalogo debe satisfacer la interfaz Buscador.
var _ Buscador = (*Catalogo)(nil)

// Catalogo encapsula un map para búsquedas rápidas y un slice para conservar orden.
type Catalogo struct {
	productos map[string]*modelo.Producto
	orden     []string
}

// NuevoCatalogo construye un catálogo y rechaza identificadores duplicados.
func NuevoCatalogo(productos []*modelo.Producto) (*Catalogo, error) {
	catalogo := &Catalogo{
		productos: make(map[string]*modelo.Producto),
		orden:     make([]string, 0, len(productos)),
	}

	for _, producto := range productos {
		if err := catalogo.Agregar(producto); err != nil {
			return nil, err
		}
	}
	return catalogo, nil
}

// Agregar usa receptor por puntero porque modifica el map y el orden del catálogo.
func (c *Catalogo) Agregar(producto *modelo.Producto) error {
	if c == nil {
		return ErrCatalogoNoInicializado
	}
	if producto == nil {
		return errors.New("el producto debe existir")
	}
	if _, existe := c.productos[producto.ID()]; existe {
		return ErrProductoDuplicado
	}
	c.productos[producto.ID()] = producto
	c.orden = append(c.orden, producto.ID())
	return nil
}

// Eliminar retira un producto y conserva consistente el slice de orden.
func (c *Catalogo) Eliminar(id string) error {
	if c == nil {
		return ErrCatalogoNoInicializado
	}
	id = strings.ToUpper(strings.TrimSpace(id))
	if _, existe := c.productos[id]; !existe {
		return ErrProductoNoEncontrado
	}
	delete(c.productos, id)
	nuevoOrden := make([]string, 0, len(c.orden)-1)
	for _, actual := range c.orden {
		if actual != id {
			nuevoOrden = append(nuevoOrden, actual)
		}
	}
	c.orden = nuevoOrden
	return nil
}

// BajoStock filtra productos mediante un recorrido sobre el slice ordenado.
func (c Catalogo) BajoStock(limite int) []modelo.Producto {
	if limite < 0 {
		limite = 0
	}
	resultado := make([]modelo.Producto, 0)
	for _, producto := range c.Listar() {
		if producto.Stock() <= limite {
			resultado = append(resultado, producto)
		}
	}
	return resultado
}

// BuscarPorID usa el map para recuperar un producto en forma directa.
func (c Catalogo) BuscarPorID(id string) (*modelo.Producto, error) {
	id = strings.ToUpper(strings.TrimSpace(id))
	producto, existe := c.productos[id]
	if !existe {
		return nil, ErrProductoNoEncontrado
	}
	return producto, nil
}

// Listar devuelve copias por valor; así el slice interno permanece encapsulado.
func (c Catalogo) Listar() []modelo.Producto {
	resultado := make([]modelo.Producto, 0, len(c.orden))
	for _, id := range c.orden {
		resultado = append(resultado, *c.productos[id])
	}
	return resultado
}

// Buscar devuelve productos cuyo nombre o categoría contenga el texto indicado.
func (c Catalogo) Buscar(texto string) []modelo.Producto {
	texto = strings.ToLower(strings.TrimSpace(texto))
	resultado := make([]modelo.Producto, 0)

	for _, producto := range c.Listar() {
		nombre := strings.ToLower(producto.Nombre())
		categoria := strings.ToLower(producto.Categoria())
		if strings.Contains(nombre, texto) || strings.Contains(categoria, texto) {
			resultado = append(resultado, producto)
		}
	}
	return resultado
}

// Primeros demuestra el uso de una subslice y luego crea una copia independiente.
// La copia evita que quien recibe el resultado modifique accidentalmente el slice base.
func (c Catalogo) Primeros(cantidad int) []modelo.Producto {
	productos := c.Listar()
	if cantidad <= 0 {
		return []modelo.Producto{}
	}
	if cantidad > len(productos) {
		cantidad = len(productos)
	}
	subSlice := productos[:cantidad]
	return append([]modelo.Producto(nil), subSlice...)
}

// Cantidad informa el número de productos registrados.
func (c Catalogo) Cantidad() int {
	return len(c.productos)
}

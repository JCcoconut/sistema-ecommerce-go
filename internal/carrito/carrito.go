// Package carrito administra los productos seleccionados por el cliente.
package carrito

import (
	"errors"
	"fmt"
	"strings"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
)

var ErrItemNoEncontrado = errors.New("el producto no está en el carrito")

// Item encapsula el producto y la cantidad seleccionada.
type Item struct {
	producto *modelo.Producto
	cantidad int
}

// NuevoItem valida la cantidad antes de crear el objeto.
func NuevoItem(producto *modelo.Producto, cantidad int) (*Item, error) {
	if producto == nil {
		return nil, errors.New("el producto es obligatorio")
	}
	if !producto.Disponible(cantidad) {
		return nil, errors.New("cantidad inválida o stock insuficiente")
	}
	return &Item{producto: producto, cantidad: cantidad}, nil
}

// Producto y Cantidad son métodos por valor porque únicamente consultan datos.
func (i Item) Producto() *modelo.Producto { return i.producto }
func (i Item) Cantidad() int              { return i.cantidad }

// Subtotal delega el cálculo al objeto Producto.
func (i Item) Subtotal() (float64, error) {
	return i.producto.CalcularSubtotal(i.cantidad)
}

// Aumentar usa receptor por puntero porque cambia la cantidad del Item.
func (i *Item) Aumentar(cantidad int) error {
	if i == nil || i.producto == nil {
		return errors.New("el ítem no es válido")
	}
	if cantidad <= 0 {
		return errors.New("la cantidad debe ser mayor que cero")
	}
	nuevaCantidad := i.cantidad + cantidad
	if !i.producto.Disponible(nuevaCantidad) {
		return errors.New("la nueva cantidad supera el stock disponible")
	}
	i.cantidad = nuevaCantidad
	return nil
}

// CambiarCantidad reemplaza la cantidad de una línea tras validar el stock.
func (i *Item) CambiarCantidad(cantidad int) error {
	if i == nil || i.producto == nil {
		return errors.New("el ítem no es válido")
	}
	if !i.producto.Disponible(cantidad) {
		return errors.New("la cantidad debe ser positiva y no superar el stock")
	}
	i.cantidad = cantidad
	return nil
}

// Carrito combina un map para localizar ítems y un slice para conservar el orden.
type Carrito struct {
	items map[string]*Item
	orden []string
}

// NuevoCarrito crea un carrito vacío listo para recibir productos.
func NuevoCarrito() *Carrito {
	return &Carrito{
		items: make(map[string]*Item),
		orden: make([]string, 0),
	}
}

// Agregar acumula cantidades cuando el producto ya existe.
// La validación se realiza antes de modificar el map para evitar estados inválidos.
func (c *Carrito) Agregar(producto *modelo.Producto, cantidad int) error {
	if c == nil {
		return errors.New("el carrito no está inicializado")
	}
	if producto == nil {
		return errors.New("el producto es obligatorio")
	}

	if item, existe := c.items[producto.ID()]; existe {
		if err := item.Aumentar(cantidad); err != nil {
			return fmt.Errorf("actualizar %s: %w", producto.ID(), err)
		}
		return nil
	}

	item, err := NuevoItem(producto, cantidad)
	if err != nil {
		return fmt.Errorf("agregar %s: %w", producto.ID(), err)
	}
	c.items[producto.ID()] = item
	c.orden = append(c.orden, producto.ID())
	return nil
}

// Eliminar retira el elemento del map y reconstruye el orden del carrito.
func (c *Carrito) Eliminar(productoID string) error {
	if c == nil {
		return errors.New("el carrito no está inicializado")
	}
	productoID = strings.ToUpper(strings.TrimSpace(productoID))
	if _, existe := c.items[productoID]; !existe {
		return ErrItemNoEncontrado
	}
	delete(c.items, productoID)

	nuevoOrden := make([]string, 0, len(c.orden)-1)
	for _, id := range c.orden {
		if id != productoID {
			nuevoOrden = append(nuevoOrden, id)
		}
	}
	c.orden = nuevoOrden
	return nil
}

// CambiarCantidad actualiza una línea ya existente.
func (c *Carrito) CambiarCantidad(productoID string, cantidad int) error {
	if c == nil {
		return errors.New("el carrito no está inicializado")
	}
	productoID = strings.ToUpper(strings.TrimSpace(productoID))
	item, existe := c.items[productoID]
	if !existe {
		return ErrItemNoEncontrado
	}
	return item.CambiarCantidad(cantidad)
}

// Items devuelve copias por valor para proteger el slice interno.
func (c Carrito) Items() []Item {
	resultado := make([]Item, 0, len(c.orden))
	for _, id := range c.orden {
		resultado = append(resultado, *c.items[id])
	}
	return resultado
}

// CantidadTotal calcula todas las unidades mediante iteración sobre el slice.
func (c Carrito) CantidadTotal() int {
	total := 0
	for _, item := range c.Items() {
		total += item.Cantidad()
	}
	return total
}

// Subtotal agrega el valor de cada línea y propaga los errores encontrados.
func (c Carrito) Subtotal() (float64, error) {
	var total float64
	for _, item := range c.Items() {
		subtotal, err := item.Subtotal()
		if err != nil {
			return 0, fmt.Errorf("calcular subtotal del carrito: %w", err)
		}
		total += subtotal
	}
	return total, nil
}

func (c Carrito) Vacio() bool {
	return len(c.items) == 0
}

// Vaciar usa receptor por puntero para reemplazar las colecciones internas.
func (c *Carrito) Vaciar() {
	c.items = make(map[string]*Item)
	c.orden = make([]string, 0)
}

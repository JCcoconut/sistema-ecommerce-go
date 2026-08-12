// Package pedido construye compras confirmadas a partir del carrito.
package pedido

import (
	"errors"
	"fmt"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/promocion"
)

const (
	EstadoConfirmado = "Confirmado"
	EstadoEnviado    = "Enviado"
	EstadoEntregado  = "Entregado"
	EstadoCancelado  = "Cancelado"
)

// FuenteCarrito es una interfaz consumida por el módulo de pedidos.
// Permite crear pedidos desde cualquier tipo que proporcione estas operaciones.
type FuenteCarrito interface {
	Vacio() bool
	Items() []carrito.Item
	Subtotal() (float64, error)
}

var _ FuenteCarrito = (*carrito.Carrito)(nil)

// Linea representa la fotografía de un producto al momento de comprarlo.
type Linea struct {
	productoID     string
	nombre         string
	cantidad       int
	precioUnitario float64
	total          float64
}

func (l Linea) ProductoID() string      { return l.productoID }
func (l Linea) Nombre() string          { return l.nombre }
func (l Linea) Cantidad() int           { return l.cantidad }
func (l Linea) PrecioUnitario() float64 { return l.precioUnitario }
func (l Linea) Total() float64          { return l.total }

// Pedido contiene un Cliente con una DireccionEntrega anidada y un slice de líneas.
type Pedido struct {
	id        string
	fecha     time.Time
	cliente   modelo.Cliente
	lineas    []Linea
	subtotal  float64
	descuento float64
	total     float64
	estado    string
}

// NuevoPedido valida primero todas las líneas y solo después descuenta stock.
// Esta ejecución en dos fases evita modificar el inventario cuando alguna línea
// del carrito es inválida o no tiene existencias suficientes.
func NuevoPedido(
	id string,
	fecha time.Time,
	cliente modelo.Cliente,
	carro FuenteCarrito,
	regla promocion.Descuento,
) (*Pedido, error) {
	if id == "" {
		return nil, errors.New("el ID del pedido es obligatorio")
	}
	if carro == nil || carro.Vacio() {
		return nil, errors.New("no se puede confirmar un carrito vacío")
	}
	if regla == nil {
		regla = promocion.SinDescuento{}
	}

	items := carro.Items()
	for _, item := range items {
		producto := item.Producto()
		if producto == nil || !producto.Disponible(item.Cantidad()) {
			return nil, fmt.Errorf("stock insuficiente para el producto %s", productoSeguroID(producto))
		}
	}

	subtotal, err := carro.Subtotal()
	if err != nil {
		return nil, fmt.Errorf("calcular subtotal del pedido: %w", err)
	}
	total, err := regla.Aplicar(subtotal)
	if err != nil {
		return nil, fmt.Errorf("aplicar descuento: %w", err)
	}
	if total < 0 || total > subtotal {
		return nil, errors.New("la regla de descuento produjo un total inválido")
	}

	lineas := make([]Linea, 0, len(items))
	for _, item := range items {
		producto := item.Producto()
		totalLinea, err := item.Subtotal()
		if err != nil {
			return nil, fmt.Errorf("crear línea %s: %w", producto.ID(), err)
		}
		lineas = append(lineas, Linea{
			productoID:     producto.ID(),
			nombre:         producto.Nombre(),
			cantidad:       item.Cantidad(),
			precioUnitario: producto.Precio(),
			total:          totalLinea,
		})
	}

	// Todos los ítems ya fueron validados, por lo que la mutación se realiza al final.
	for _, item := range items {
		if err := item.Producto().DescontarStock(item.Cantidad()); err != nil {
			return nil, fmt.Errorf("actualizar inventario: %w", err)
		}
	}

	return &Pedido{
		id:        id,
		fecha:     fecha,
		cliente:   cliente,
		lineas:    lineas,
		subtotal:  subtotal,
		descuento: subtotal - total,
		total:     total,
		estado:    EstadoConfirmado,
	}, nil
}

func productoSeguroID(producto *modelo.Producto) string {
	if producto == nil {
		return "desconocido"
	}
	return producto.ID()
}

// Métodos por valor para consultar el estado encapsulado del pedido.
func (p Pedido) ID() string              { return p.id }
func (p Pedido) Fecha() time.Time        { return p.fecha }
func (p Pedido) Cliente() modelo.Cliente { return p.cliente }
func (p Pedido) Subtotal() float64       { return p.subtotal }
func (p Pedido) Descuento() float64      { return p.descuento }
func (p Pedido) Total() float64          { return p.total }
func (p Pedido) Estado() string          { return p.estado }

// Lineas devuelve una copia para preservar la encapsulación del slice interno.
func (p Pedido) Lineas() []Linea {
	return append([]Linea(nil), p.lineas...)
}

// CambiarEstado usa receptor por puntero porque modifica el pedido original.
func (p *Pedido) CambiarEstado(nuevoEstado string) error {
	if p == nil {
		return errors.New("el pedido no existe")
	}
	estadosValidos := map[string]bool{
		EstadoConfirmado: true,
		EstadoEnviado:    true,
		EstadoEntregado:  true,
		EstadoCancelado:  true,
	}
	if !estadosValidos[nuevoEstado] {
		return errors.New("el estado solicitado no es válido")
	}
	p.estado = nuevoEstado
	return nil
}

// NuevoGeneradorID crea un closure que conserva el contador entre llamadas.
func NuevoGeneradorID() func() string {
	contador := 0
	return func() string {
		contador++
		return fmt.Sprintf("PED-%03d", contador)
	}
}

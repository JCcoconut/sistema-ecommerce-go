// Package facturacion genera comprobantes académicos sin validez tributaria.
package facturacion

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
)

const (
	TarifaIVA        = 0.15
	AdvertenciaLegal = "FACTURA ACADÉMICA SIMULADA — SIN VALIDEZ TRIBUTARIA"
)

// Emisor permite sustituir la facturación simulada por otra implementación.
// El proyecto no almacena tarjetas ni consume servicios reales del SRI.
type Emisor interface {
	Emitir(orden pedido.Pedido) (*Factura, error)
}

// EmisorSimulado implementa Emisor para fines demostrativos.
type EmisorSimulado struct{}

var _ Emisor = EmisorSimulado{}

// Linea es la representación inmutable de un producto facturado.
type Linea struct {
	productoID     string
	descripcion    string
	cantidad       int
	precioUnitario float64
	total          float64
}

func (l Linea) ProductoID() string      { return l.productoID }
func (l Linea) Descripcion() string     { return l.descripcion }
func (l Linea) Cantidad() int           { return l.cantidad }
func (l Linea) PrecioUnitario() float64 { return l.precioUnitario }
func (l Linea) Total() float64          { return l.total }

// Factura conserva una fotografía del pedido y el desglose académico de IVA.
// Se supone que el total del pedido ya incluye impuestos.
type Factura struct {
	id            string
	pedidoID      string
	fecha         time.Time
	emisor        string
	cliente       string
	email         string
	lineas        []Linea
	subtotal      float64
	descuento     float64
	baseImponible float64
	iva           float64
	total         float64
	claveSimulada string
	advertencia   string
}

func redondear(valor float64) float64 { return math.Round(valor*100) / 100 }

func (EmisorSimulado) Emitir(orden pedido.Pedido) (*Factura, error) {
	if orden.ID() == "" || orden.Total() <= 0 || len(orden.Lineas()) == 0 {
		return nil, errors.New("no se puede facturar un pedido incompleto")
	}
	lineas := make([]Linea, 0, len(orden.Lineas()))
	for _, linea := range orden.Lineas() {
		lineas = append(lineas, Linea{
			productoID: linea.ProductoID(), descripcion: linea.Nombre(),
			cantidad: linea.Cantidad(), precioUnitario: linea.PrecioUnitario(), total: linea.Total(),
		})
	}
	fecha := time.Now().UTC()
	base := redondear(orden.Total() / (1 + TarifaIVA))
	iva := redondear(orden.Total() - base)
	semilla := fmt.Sprintf("%s|%s|%.2f", orden.ID(), fecha.Format(time.RFC3339Nano), orden.Total())
	resumen := fmt.Sprintf("%X", sha256.Sum256([]byte(semilla)))
	return &Factura{
		id: "FAC-" + strings.TrimPrefix(orden.ID(), "PED-"), pedidoID: orden.ID(), fecha: fecha,
		emisor: "AudioCyber Store", cliente: orden.Cliente().Nombre(), email: orden.Cliente().Email(),
		lineas: lineas, baseImponible: base, iva: iva, total: redondear(orden.Total()),
		subtotal: redondear(orden.Subtotal()), descuento: redondear(orden.Descuento()),
		claveSimulada: resumen[:24], advertencia: AdvertenciaLegal,
	}, nil
}

func (f Factura) ID() string             { return f.id }
func (f Factura) PedidoID() string       { return f.pedidoID }
func (f Factura) Fecha() time.Time       { return f.fecha }
func (f Factura) Emisor() string         { return f.emisor }
func (f Factura) Cliente() string        { return f.cliente }
func (f Factura) Email() string          { return f.email }
func (f Factura) BaseImponible() float64 { return f.baseImponible }
func (f Factura) Subtotal() float64      { return f.subtotal }
func (f Factura) Descuento() float64     { return f.descuento }
func (f Factura) IVA() float64           { return f.iva }
func (f Factura) Total() float64         { return f.total }
func (f Factura) ClaveSimulada() string  { return f.claveSimulada }
func (f Factura) Advertencia() string    { return f.advertencia }

func (f Factura) Lineas() []Linea { return append([]Linea(nil), f.lineas...) }

// RestaurarFactura reconstruye un comprobante validado desde el archivo JSON.
func RestaurarFactura(
	id, pedidoID string,
	fecha time.Time,
	emisor, cliente, email string,
	lineas []Linea,
	subtotal, descuento float64,
	baseImponible, iva, total float64,
	clave string,
) (*Factura, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(pedidoID) == "" || fecha.IsZero() ||
		strings.TrimSpace(cliente) == "" || len(lineas) == 0 || total <= 0 {
		return nil, errors.New("la factura persistida está incompleta")
	}
	if math.Abs((baseImponible+iva)-total) > 0.02 {
		return nil, errors.New("el desglose de la factura persistida no coincide")
	}
	if math.Abs((subtotal-descuento)-total) > 0.02 {
		return nil, errors.New("el descuento de la factura persistida no coincide")
	}
	return &Factura{
		id: id, pedidoID: pedidoID, fecha: fecha, emisor: emisor, cliente: cliente, email: email,
		lineas: append([]Linea(nil), lineas...), subtotal: subtotal, descuento: descuento,
		baseImponible: baseImponible, iva: iva,
		total: total, claveSimulada: clave, advertencia: AdvertenciaLegal,
	}, nil
}

// NuevaLineaRestaurada crea una línea proveniente de persistencia.
func NuevaLineaRestaurada(productoID, descripcion string, cantidad int, precioUnitario, total float64) (Linea, error) {
	if strings.TrimSpace(productoID) == "" || strings.TrimSpace(descripcion) == "" ||
		cantidad <= 0 || precioUnitario <= 0 || total <= 0 {
		return Linea{}, errors.New("la línea de factura persistida no es válida")
	}
	return Linea{productoID: productoID, descripcion: descripcion, cantidad: cantidad, precioUnitario: precioUnitario, total: total}, nil
}

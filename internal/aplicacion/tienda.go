// Package aplicacion coordina los módulos del dominio y protege el estado compartido.
package aplicacion

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/catalogo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/facturacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/persistencia"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/promocion"
)

var (
	ErrClienteNoEncontrado = errors.New("cliente no encontrado")
	ErrPedidoNoEncontrado  = errors.New("pedido no encontrado")
	ErrFacturaNoEncontrada = errors.New("factura no encontrada")
	ErrProductoEnCarrito   = errors.New("el producto está siendo utilizado por el carrito")
)

// Tienda es el agregado principal de la aplicación.
// net/http atiende cada solicitud en una goroutine; por ello, el RWMutex es
// imprescindible para impedir carreras sobre maps, slices, stock y carrito.
type Tienda struct {
	mu sync.RWMutex

	catalogo      *catalogo.Catalogo
	carro         *carrito.Carrito
	cupones       *promocion.GestorCupones
	descuento     promocion.Descuento
	cuponActivo   string
	clientes      map[string]modelo.Cliente
	ordenClientes []string
	pedidos       map[string]*pedido.Pedido
	ordenPedidos  []string
	facturas      map[string]*facturacion.Factura
	ordenFacturas []string
	contador      int

	repositorio persistencia.Repositorio
	emisor      facturacion.Emisor
}

func NuevaTienda(repositorio persistencia.Repositorio, emisor facturacion.Emisor) (*Tienda, error) {
	cupones, err := promocion.NuevoGestorCupones()
	if err != nil {
		return nil, fmt.Errorf("inicializar cupones: %w", err)
	}
	if emisor == nil {
		emisor = facturacion.EmisorSimulado{}
	}
	tienda := &Tienda{
		carro: carrito.NuevoCarrito(), cupones: cupones, descuento: promocion.SinDescuento{},
		clientes: make(map[string]modelo.Cliente), pedidos: make(map[string]*pedido.Pedido),
		facturas: make(map[string]*facturacion.Factura), repositorio: repositorio, emisor: emisor,
	}

	if repositorio != nil {
		var estado EstadoPersistido
		err := repositorio.Cargar(&estado)
		switch {
		case err == nil:
			if err := tienda.restaurar(estado); err != nil {
				return nil, fmt.Errorf("restaurar estado: %w", err)
			}
			return tienda, nil
		case !errors.Is(err, persistencia.ErrSinDatos):
			return nil, fmt.Errorf("cargar persistencia: %w", err)
		}
	}

	tienda.catalogo, err = catalogo.CrearInicial()
	if err != nil {
		return nil, fmt.Errorf("crear catálogo inicial: %w", err)
	}
	if err := tienda.persistirSinBloqueo(); err != nil {
		return nil, err
	}
	return tienda, nil
}

func (t *Tienda) Resumen() ResumenDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return ResumenDTO{
		Estado: "ok", Aplicacion: "AudioCyber Store API",
		Productos: t.catalogo.Cantidad(), Clientes: len(t.clientes), Pedidos: len(t.pedidos),
		Facturas: len(t.facturas), Persistencia: t.repositorio != nil,
	}
}

func (t *Tienda) ListarProductos(busqueda string) []ProductoDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	productos := t.catalogo.Listar()
	if strings.TrimSpace(busqueda) != "" {
		productos = t.catalogo.Buscar(busqueda)
	}
	return productosDTO(productos)
}

func (t *Tienda) Producto(id string) (ProductoDTO, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	producto, err := t.catalogo.BuscarPorID(id)
	if err != nil {
		return ProductoDTO{}, err
	}
	return productoDTO(*producto), nil
}

func (t *Tienda) ProductosBajoStock(limite int) []ProductoDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return productosDTO(t.catalogo.BajoStock(limite))
}

func (t *Tienda) CrearProducto(entrada ProductoEntrada) (ProductoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	proveedor, err := modelo.NuevoProveedor(entrada.ProveedorNombre, entrada.ProveedorEmail)
	if err != nil {
		return ProductoDTO{}, fmt.Errorf("crear proveedor: %w", err)
	}
	producto, err := modelo.NuevoProducto(entrada.ID, entrada.Nombre, entrada.Categoria, entrada.Precio, entrada.Stock, proveedor)
	if err != nil {
		return ProductoDTO{}, err
	}
	if err := t.catalogo.Agregar(producto); err != nil {
		return ProductoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return ProductoDTO{}, err
	}
	return productoDTO(*producto), nil
}

func (t *Tienda) ActualizarProducto(id string, entrada ProductoEntrada) (ProductoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	producto, err := t.catalogo.BuscarPorID(id)
	if err != nil {
		return ProductoDTO{}, err
	}
	for _, item := range t.carro.Items() {
		if item.Producto().ID() == producto.ID() && entrada.Stock < item.Cantidad() {
			return ProductoDTO{}, ErrProductoEnCarrito
		}
	}
	proveedor, err := modelo.NuevoProveedor(entrada.ProveedorNombre, entrada.ProveedorEmail)
	if err != nil {
		return ProductoDTO{}, err
	}
	if err := producto.Actualizar(entrada.Nombre, entrada.Categoria, entrada.Precio, proveedor); err != nil {
		return ProductoDTO{}, err
	}
	diferencia := entrada.Stock - producto.Stock()
	if diferencia > 0 {
		err = producto.ReponerStock(diferencia)
	} else if diferencia < 0 {
		err = producto.DescontarStock(-diferencia)
	}
	if err != nil {
		return ProductoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return ProductoDTO{}, err
	}
	return productoDTO(*producto), nil
}

func (t *Tienda) EliminarProducto(id string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, item := range t.carro.Items() {
		if strings.EqualFold(item.Producto().ID(), id) {
			return ErrProductoEnCarrito
		}
	}
	if err := t.catalogo.Eliminar(id); err != nil {
		return err
	}
	return t.persistirSinBloqueo()
}

func (t *Tienda) VerCarrito() (CarritoDTO, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.carritoDTOSinBloqueo()
}

func (t *Tienda) AgregarAlCarrito(id string, cantidad int) (CarritoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	producto, err := t.catalogo.BuscarPorID(id)
	if err != nil {
		return CarritoDTO{}, err
	}
	if err := t.carro.Agregar(producto, cantidad); err != nil {
		return CarritoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return CarritoDTO{}, err
	}
	return t.carritoDTOSinBloqueo()
}

func (t *Tienda) CambiarCantidadCarrito(id string, cantidad int) (CarritoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.carro.CambiarCantidad(id, cantidad); err != nil {
		return CarritoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return CarritoDTO{}, err
	}
	return t.carritoDTOSinBloqueo()
}

func (t *Tienda) EliminarDelCarrito(id string) (CarritoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if err := t.carro.Eliminar(id); err != nil {
		return CarritoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return CarritoDTO{}, err
	}
	return t.carritoDTOSinBloqueo()
}

func (t *Tienda) AplicarCupon(codigo string) (CarritoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	descuento, err := t.cupones.Buscar(codigo)
	if err != nil {
		return CarritoDTO{}, err
	}
	t.descuento = descuento
	t.cuponActivo = strings.ToUpper(strings.TrimSpace(codigo))
	if err := t.persistirSinBloqueo(); err != nil {
		return CarritoDTO{}, err
	}
	return t.carritoDTOSinBloqueo()
}

func (t *Tienda) Cupones() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cupones.Codigos()
}

func (t *Tienda) CrearCliente(entrada ClienteEntrada) (ClienteDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cliente, err := construirCliente(entrada)
	if err != nil {
		return ClienteDTO{}, err
	}
	clave := strings.ToLower(cliente.Email())
	if _, existe := t.clientes[clave]; !existe {
		t.ordenClientes = append(t.ordenClientes, clave)
	}
	t.clientes[clave] = cliente
	if err := t.persistirSinBloqueo(); err != nil {
		return ClienteDTO{}, err
	}
	return clienteDTO(cliente), nil
}

func (t *Tienda) ListarClientes() []ClienteDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	resultado := make([]ClienteDTO, 0, len(t.ordenClientes))
	for _, email := range t.ordenClientes {
		resultado = append(resultado, clienteDTO(t.clientes[email]))
	}
	return resultado
}

func (t *Tienda) ConfirmarCompra(entrada ClienteEntrada) (CompraDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	cliente, err := construirCliente(entrada)
	if err != nil {
		return CompraDTO{}, err
	}
	if t.carro.Vacio() {
		return CompraDTO{}, errors.New("el carrito está vacío")
	}
	t.contador++
	id := fmt.Sprintf("PED-%03d", t.contador)
	orden, err := pedido.NuevoPedido(id, time.Now().UTC(), cliente, t.carro, t.descuento)
	if err != nil {
		t.contador--
		return CompraDTO{}, err
	}
	factura, err := t.emisor.Emitir(*orden)
	if err != nil {
		return CompraDTO{}, fmt.Errorf("emitir factura simulada: %w", err)
	}
	claveCliente := strings.ToLower(cliente.Email())
	if _, existe := t.clientes[claveCliente]; !existe {
		t.ordenClientes = append(t.ordenClientes, claveCliente)
	}
	t.clientes[claveCliente] = cliente
	t.pedidos[orden.ID()] = orden
	t.ordenPedidos = append(t.ordenPedidos, orden.ID())
	t.facturas[factura.ID()] = factura
	t.ordenFacturas = append(t.ordenFacturas, factura.ID())
	t.carro.Vaciar()
	t.descuento = promocion.SinDescuento{}
	t.cuponActivo = ""
	if err := t.persistirSinBloqueo(); err != nil {
		return CompraDTO{}, err
	}
	return CompraDTO{Pedido: pedidoDTO(*orden), Factura: facturaDTO(*factura)}, nil
}

func (t *Tienda) ListarPedidos(email string) []PedidoDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	email = strings.ToLower(strings.TrimSpace(email))
	resultado := make([]PedidoDTO, 0)
	for _, id := range t.ordenPedidos {
		orden := t.pedidos[id]
		if email == "" || strings.EqualFold(orden.Cliente().Email(), email) {
			resultado = append(resultado, pedidoDTO(*orden))
		}
	}
	return resultado
}

func (t *Tienda) Pedido(id string) (PedidoDTO, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	orden, existe := t.pedidos[strings.TrimSpace(id)]
	if !existe {
		return PedidoDTO{}, ErrPedidoNoEncontrado
	}
	return pedidoDTO(*orden), nil
}

func (t *Tienda) CambiarEstadoPedido(id, estado string) (PedidoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	orden, existe := t.pedidos[strings.TrimSpace(id)]
	if !existe {
		return PedidoDTO{}, ErrPedidoNoEncontrado
	}
	if err := orden.CambiarEstado(estado); err != nil {
		return PedidoDTO{}, err
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return PedidoDTO{}, err
	}
	return pedidoDTO(*orden), nil
}

func (t *Tienda) CancelarPedido(id string) (PedidoDTO, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	orden, existe := t.pedidos[strings.TrimSpace(id)]
	if !existe {
		return PedidoDTO{}, ErrPedidoNoEncontrado
	}
	if orden.Estado() != pedido.EstadoConfirmado {
		return PedidoDTO{}, fmt.Errorf("%w: solo se cancela un pedido confirmado", pedido.ErrTransicionInvalida)
	}
	if err := orden.CambiarEstado(pedido.EstadoCancelado); err != nil {
		return PedidoDTO{}, err
	}
	for _, linea := range orden.Lineas() {
		producto, err := t.catalogo.BuscarPorID(linea.ProductoID())
		if err == nil {
			_ = producto.ReponerStock(linea.Cantidad())
		}
	}
	if err := t.persistirSinBloqueo(); err != nil {
		return PedidoDTO{}, err
	}
	return pedidoDTO(*orden), nil
}

func (t *Tienda) ListarFacturas() []FacturaDTO {
	t.mu.RLock()
	defer t.mu.RUnlock()
	resultado := make([]FacturaDTO, 0, len(t.ordenFacturas))
	for _, id := range t.ordenFacturas {
		resultado = append(resultado, facturaDTO(*t.facturas[id]))
	}
	return resultado
}

func (t *Tienda) Factura(id string) (FacturaDTO, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	factura, existe := t.facturas[strings.TrimSpace(id)]
	if !existe {
		return FacturaDTO{}, ErrFacturaNoEncontrada
	}
	return facturaDTO(*factura), nil
}

func (t *Tienda) carritoDTOSinBloqueo() (CarritoDTO, error) {
	items := make([]ItemCarritoDTO, 0, len(t.carro.Items()))
	for _, item := range t.carro.Items() {
		total, err := item.Subtotal()
		if err != nil {
			return CarritoDTO{}, err
		}
		items = append(items, ItemCarritoDTO{ProductoID: item.Producto().ID(), Nombre: item.Producto().Nombre(), Cantidad: item.Cantidad(), Precio: item.Producto().Precio(), Total: total})
	}
	subtotal, err := t.carro.Subtotal()
	if err != nil {
		return CarritoDTO{}, err
	}
	total, err := t.descuento.Aplicar(subtotal)
	if err != nil {
		return CarritoDTO{}, err
	}
	return CarritoDTO{Items: items, CantidadTotal: t.carro.CantidadTotal(), Subtotal: subtotal, Cupon: t.cuponActivo, Descuento: subtotal - total, Total: total}, nil
}

func productoDTO(p modelo.Producto) ProductoDTO {
	return ProductoDTO{ID: p.ID(), Nombre: p.Nombre(), Categoria: p.Categoria(), Precio: p.Precio(), Stock: p.Stock(), Proveedor: ProveedorDTO{Nombre: p.Proveedor().Nombre(), Email: p.Proveedor().Email()}}
}

func productosDTO(productos []modelo.Producto) []ProductoDTO {
	resultado := make([]ProductoDTO, 0, len(productos))
	for _, producto := range productos {
		resultado = append(resultado, productoDTO(producto))
	}
	return resultado
}

func clienteDTO(c modelo.Cliente) ClienteDTO {
	d := c.Direccion()
	return ClienteDTO{Nombre: c.Nombre(), Email: c.Email(), Direccion: DireccionDTO{Ciudad: d.Ciudad(), Sector: d.Sector(), Referencia: d.Referencia()}}
}

func construirCliente(e ClienteEntrada) (modelo.Cliente, error) {
	direccion, err := modelo.NuevaDireccionEntrega(e.Ciudad, e.Sector, e.Referencia)
	if err != nil {
		return modelo.Cliente{}, err
	}
	return modelo.NuevoCliente(e.Nombre, e.Email, direccion)
}

func pedidoDTO(p pedido.Pedido) PedidoDTO {
	lineas := make([]LineaPedidoDTO, 0, len(p.Lineas()))
	for _, l := range p.Lineas() {
		lineas = append(lineas, LineaPedidoDTO{ProductoID: l.ProductoID(), Nombre: l.Nombre(), Cantidad: l.Cantidad(), PrecioUnitario: l.PrecioUnitario(), Total: l.Total()})
	}
	return PedidoDTO{ID: p.ID(), Fecha: p.Fecha(), Cliente: clienteDTO(p.Cliente()), Lineas: lineas, Subtotal: p.Subtotal(), Descuento: p.Descuento(), Total: p.Total(), Estado: p.Estado()}
}

func facturaDTO(f facturacion.Factura) FacturaDTO {
	lineas := make([]LineaFacturaDTO, 0, len(f.Lineas()))
	for _, l := range f.Lineas() {
		lineas = append(lineas, LineaFacturaDTO{ProductoID: l.ProductoID(), Descripcion: l.Descripcion(), Cantidad: l.Cantidad(), PrecioUnitario: l.PrecioUnitario(), Total: l.Total()})
	}
	return FacturaDTO{ID: f.ID(), PedidoID: f.PedidoID(), Fecha: f.Fecha(), Emisor: f.Emisor(), Cliente: f.Cliente(), Email: f.Email(), Lineas: lineas, Subtotal: f.Subtotal(), Descuento: f.Descuento(), BaseImponible: f.BaseImponible(), TarifaIVA: facturacion.TarifaIVA, IVA: f.IVA(), Total: f.Total(), ClaveSimulada: f.ClaveSimulada(), Advertencia: f.Advertencia()}
}

func (t *Tienda) IDsOrdenados() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ids := append([]string(nil), t.ordenPedidos...)
	sort.Strings(ids)
	return ids
}

func contadorDesdeID(id string) int {
	partes := strings.Split(id, "-")
	if len(partes) != 2 {
		return 0
	}
	valor, _ := strconv.Atoi(partes[1])
	return valor
}

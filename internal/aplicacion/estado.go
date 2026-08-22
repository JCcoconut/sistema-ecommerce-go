package aplicacion

import (
	"fmt"
	"strings"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/catalogo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/facturacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/promocion"
)

func (t *Tienda) persistirSinBloqueo() error {
	if t.repositorio == nil {
		return nil
	}
	estado, err := t.estadoSinBloqueo()
	if err != nil {
		return fmt.Errorf("preparar persistencia: %w", err)
	}
	if err := t.repositorio.Guardar(estado); err != nil {
		return fmt.Errorf("guardar persistencia: %w", err)
	}
	return nil
}

func (t *Tienda) estadoSinBloqueo() (EstadoPersistido, error) {
	carro := make([]CantidadEntrada, 0, len(t.carro.Items()))
	for _, item := range t.carro.Items() {
		carro = append(carro, CantidadEntrada{ProductoID: item.Producto().ID(), Cantidad: item.Cantidad()})
	}
	clientes := make([]ClienteDTO, 0, len(t.ordenClientes))
	for _, email := range t.ordenClientes {
		clientes = append(clientes, clienteDTO(t.clientes[email]))
	}
	pedidos := make([]PedidoDTO, 0, len(t.ordenPedidos))
	for _, id := range t.ordenPedidos {
		pedidos = append(pedidos, pedidoDTO(*t.pedidos[id]))
	}
	facturas := make([]FacturaDTO, 0, len(t.ordenFacturas))
	for _, id := range t.ordenFacturas {
		facturas = append(facturas, facturaDTO(*t.facturas[id]))
	}
	return EstadoPersistido{
		Version: 1, GuardadoEn: time.Now().UTC(), Productos: productosDTO(t.catalogo.Listar()),
		Carrito: carro, Cupon: t.cuponActivo, Clientes: clientes, Pedidos: pedidos,
		Facturas: facturas, ContadorPedido: t.contador,
	}, nil
}

func (t *Tienda) restaurar(estado EstadoPersistido) error {
	if estado.Version != 1 {
		return fmt.Errorf("versión de persistencia no soportada: %d", estado.Version)
	}
	productos := make([]*modelo.Producto, 0, len(estado.Productos))
	for _, dto := range estado.Productos {
		proveedor, err := modelo.NuevoProveedor(dto.Proveedor.Nombre, dto.Proveedor.Email)
		if err != nil {
			return fmt.Errorf("restaurar proveedor de %s: %w", dto.ID, err)
		}
		producto, err := modelo.NuevoProducto(dto.ID, dto.Nombre, dto.Categoria, dto.Precio, dto.Stock, proveedor)
		if err != nil {
			return fmt.Errorf("restaurar producto %s: %w", dto.ID, err)
		}
		productos = append(productos, producto)
	}
	catalogoRestaurado, err := catalogo.NuevoCatalogo(productos)
	if err != nil {
		return err
	}
	t.catalogo = catalogoRestaurado
	t.carro = carrito.NuevoCarrito()
	for _, item := range estado.Carrito {
		producto, err := t.catalogo.BuscarPorID(item.ProductoID)
		if err != nil {
			return fmt.Errorf("restaurar carrito: %w", err)
		}
		if err := t.carro.Agregar(producto, item.Cantidad); err != nil {
			return fmt.Errorf("restaurar carrito: %w", err)
		}
	}
	t.descuento = promocion.SinDescuento{}
	t.cuponActivo = ""
	if strings.TrimSpace(estado.Cupon) != "" {
		descuento, err := t.cupones.Buscar(estado.Cupon)
		if err != nil {
			return fmt.Errorf("restaurar cupón: %w", err)
		}
		t.descuento = descuento
		t.cuponActivo = estado.Cupon
	}

	t.clientes = make(map[string]modelo.Cliente)
	t.ordenClientes = nil
	for _, dto := range estado.Clientes {
		cliente, err := construirCliente(ClienteEntrada{Nombre: dto.Nombre, Email: dto.Email, Ciudad: dto.Direccion.Ciudad, Sector: dto.Direccion.Sector, Referencia: dto.Direccion.Referencia})
		if err != nil {
			return fmt.Errorf("restaurar cliente %s: %w", dto.Email, err)
		}
		clave := strings.ToLower(cliente.Email())
		t.clientes[clave] = cliente
		t.ordenClientes = append(t.ordenClientes, clave)
	}

	t.pedidos = make(map[string]*pedido.Pedido)
	t.ordenPedidos = nil
	maximoContador := estado.ContadorPedido
	for _, dto := range estado.Pedidos {
		cliente, err := construirCliente(ClienteEntrada{Nombre: dto.Cliente.Nombre, Email: dto.Cliente.Email, Ciudad: dto.Cliente.Direccion.Ciudad, Sector: dto.Cliente.Direccion.Sector, Referencia: dto.Cliente.Direccion.Referencia})
		if err != nil {
			return fmt.Errorf("restaurar cliente de pedido %s: %w", dto.ID, err)
		}
		lineas := make([]pedido.Linea, 0, len(dto.Lineas))
		for _, lineaDTO := range dto.Lineas {
			linea, err := pedido.NuevaLineaRestaurada(lineaDTO.ProductoID, lineaDTO.Nombre, lineaDTO.Cantidad, lineaDTO.PrecioUnitario)
			if err != nil {
				return fmt.Errorf("restaurar línea de %s: %w", dto.ID, err)
			}
			lineas = append(lineas, linea)
		}
		orden, err := pedido.RestaurarPedido(dto.ID, dto.Fecha, cliente, lineas, dto.Subtotal, dto.Descuento, dto.Total, dto.Estado)
		if err != nil {
			return err
		}
		t.pedidos[dto.ID] = orden
		t.ordenPedidos = append(t.ordenPedidos, dto.ID)
		if valor := contadorDesdeID(dto.ID); valor > maximoContador {
			maximoContador = valor
		}
	}
	t.contador = maximoContador

	t.facturas = make(map[string]*facturacion.Factura)
	t.ordenFacturas = nil
	for _, dto := range estado.Facturas {
		lineas := make([]facturacion.Linea, 0, len(dto.Lineas))
		for _, lineaDTO := range dto.Lineas {
			linea, err := facturacion.NuevaLineaRestaurada(lineaDTO.ProductoID, lineaDTO.Descripcion, lineaDTO.Cantidad, lineaDTO.PrecioUnitario, lineaDTO.Total)
			if err != nil {
				return err
			}
			lineas = append(lineas, linea)
		}
		factura, err := facturacion.RestaurarFactura(dto.ID, dto.PedidoID, dto.Fecha, dto.Emisor, dto.Cliente, dto.Email, lineas, dto.Subtotal, dto.Descuento, dto.BaseImponible, dto.IVA, dto.Total, dto.ClaveSimulada)
		if err != nil {
			return err
		}
		t.facturas[dto.ID] = factura
		t.ordenFacturas = append(t.ordenFacturas, dto.ID)
	}
	return nil
}

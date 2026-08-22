package aplicacion

import "time"

type ProveedorDTO struct {
	Nombre string `json:"nombre"`
	Email  string `json:"email"`
}

type ProductoDTO struct {
	ID        string       `json:"id"`
	Nombre    string       `json:"nombre"`
	Categoria string       `json:"categoria"`
	Precio    float64      `json:"precio"`
	Stock     int          `json:"stock"`
	Proveedor ProveedorDTO `json:"proveedor"`
}

type ProductoEntrada struct {
	ID              string  `json:"id"`
	Nombre          string  `json:"nombre"`
	Categoria       string  `json:"categoria"`
	Precio          float64 `json:"precio"`
	Stock           int     `json:"stock"`
	ProveedorNombre string  `json:"proveedor_nombre"`
	ProveedorEmail  string  `json:"proveedor_email"`
}

type DireccionDTO struct {
	Ciudad     string `json:"ciudad"`
	Sector     string `json:"sector"`
	Referencia string `json:"referencia,omitempty"`
}

type ClienteDTO struct {
	Nombre    string       `json:"nombre"`
	Email     string       `json:"email"`
	Direccion DireccionDTO `json:"direccion"`
}

type ClienteEntrada struct {
	Nombre     string `json:"nombre"`
	Email      string `json:"email"`
	Ciudad     string `json:"ciudad"`
	Sector     string `json:"sector"`
	Referencia string `json:"referencia"`
}

type ItemCarritoDTO struct {
	ProductoID string  `json:"producto_id"`
	Nombre     string  `json:"nombre"`
	Cantidad   int     `json:"cantidad"`
	Precio     float64 `json:"precio_unitario"`
	Total      float64 `json:"total"`
}

type CarritoDTO struct {
	Items         []ItemCarritoDTO `json:"items"`
	CantidadTotal int              `json:"cantidad_total"`
	Subtotal      float64          `json:"subtotal"`
	Cupon         string           `json:"cupon,omitempty"`
	Descuento     float64          `json:"descuento"`
	Total         float64          `json:"total"`
}

type CantidadEntrada struct {
	ProductoID string `json:"producto_id,omitempty"`
	Cantidad   int    `json:"cantidad"`
}

type CuponEntrada struct {
	Codigo string `json:"codigo"`
}

type LineaPedidoDTO struct {
	ProductoID     string  `json:"producto_id"`
	Nombre         string  `json:"nombre"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Total          float64 `json:"total"`
}

type PedidoDTO struct {
	ID        string           `json:"id"`
	Fecha     time.Time        `json:"fecha"`
	Cliente   ClienteDTO       `json:"cliente"`
	Lineas    []LineaPedidoDTO `json:"lineas"`
	Subtotal  float64          `json:"subtotal"`
	Descuento float64          `json:"descuento"`
	Total     float64          `json:"total"`
	Estado    string           `json:"estado"`
}

type EstadoEntrada struct {
	Estado string `json:"estado"`
}

type LineaFacturaDTO struct {
	ProductoID     string  `json:"producto_id"`
	Descripcion    string  `json:"descripcion"`
	Cantidad       int     `json:"cantidad"`
	PrecioUnitario float64 `json:"precio_unitario"`
	Total          float64 `json:"total"`
}

type FacturaDTO struct {
	ID            string            `json:"id"`
	PedidoID      string            `json:"pedido_id"`
	Fecha         time.Time         `json:"fecha"`
	Emisor        string            `json:"emisor"`
	Cliente       string            `json:"cliente"`
	Email         string            `json:"email"`
	Lineas        []LineaFacturaDTO `json:"lineas"`
	Subtotal      float64           `json:"subtotal_productos"`
	Descuento     float64           `json:"descuento"`
	BaseImponible float64           `json:"base_imponible"`
	TarifaIVA     float64           `json:"tarifa_iva"`
	IVA           float64           `json:"iva"`
	Total         float64           `json:"total"`
	ClaveSimulada string            `json:"clave_simulada"`
	Advertencia   string            `json:"advertencia"`
}

type CompraDTO struct {
	Pedido  PedidoDTO  `json:"pedido"`
	Factura FacturaDTO `json:"factura"`
}

type EstadoPersistido struct {
	Version        int               `json:"version"`
	GuardadoEn     time.Time         `json:"guardado_en"`
	Productos      []ProductoDTO     `json:"productos"`
	Carrito        []CantidadEntrada `json:"carrito"`
	Cupon          string            `json:"cupon,omitempty"`
	Clientes       []ClienteDTO      `json:"clientes"`
	Pedidos        []PedidoDTO       `json:"pedidos"`
	Facturas       []FacturaDTO      `json:"facturas"`
	ContadorPedido int               `json:"contador_pedido"`
}

type ResumenDTO struct {
	Estado       string `json:"estado"`
	Aplicacion   string `json:"aplicacion"`
	Productos    int    `json:"productos"`
	Clientes     int    `json:"clientes"`
	Pedidos      int    `json:"pedidos"`
	Facturas     int    `json:"facturas"`
	Persistencia bool   `json:"persistencia"`
}

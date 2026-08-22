package webapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/aplicacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/carrito"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/catalogo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/modelo"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/pedido"
)

const maximoCuerpo = 1 << 20 // 1 MiB

type errorDTO struct {
	Error struct {
		Codigo  string `json:"codigo"`
		Mensaje string `json:"mensaje"`
	} `json:"error"`
}

func escribirJSON(w http.ResponseWriter, estado int, valor any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(estado)
	if err := json.NewEncoder(w).Encode(valor); err != nil {
		return
	}
}

func escribirError(w http.ResponseWriter, estado int, codigo, mensaje string) {
	respuesta := errorDTO{}
	respuesta.Error.Codigo = codigo
	respuesta.Error.Mensaje = mensaje
	escribirJSON(w, estado, respuesta)
}

func decodificarJSON(w http.ResponseWriter, r *http.Request, destino any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maximoCuerpo)
	decodificador := json.NewDecoder(r.Body)
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		return fmt.Errorf("JSON inválido: %w", err)
	}
	if err := decodificador.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("el cuerpo debe contener un único documento JSON")
	}
	return nil
}

func responderErrorDominio(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, catalogo.ErrProductoNoEncontrado),
		errors.Is(err, carrito.ErrItemNoEncontrado),
		errors.Is(err, aplicacion.ErrClienteNoEncontrado),
		errors.Is(err, aplicacion.ErrPedidoNoEncontrado),
		errors.Is(err, aplicacion.ErrFacturaNoEncontrada):
		escribirError(w, http.StatusNotFound, "no_encontrado", err.Error())
	case errors.Is(err, catalogo.ErrProductoDuplicado),
		errors.Is(err, aplicacion.ErrProductoEnCarrito),
		errors.Is(err, pedido.ErrTransicionInvalida),
		errors.Is(err, modelo.ErrStockInsuficiente):
		escribirError(w, http.StatusConflict, "conflicto", err.Error())
	case strings.Contains(err.Error(), "persistencia"):
		escribirError(w, http.StatusInternalServerError, "persistencia", "no se pudo guardar el estado")
	default:
		escribirError(w, http.StatusBadRequest, "datos_invalidos", err.Error())
	}
}

func (s *Servidor) health(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { escribirJSON(w, http.StatusOK, t.Resumen()) }
}

func (s *Servidor) listarProductos(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, t.ListarProductos(r.URL.Query().Get("q")))
	}
}

func (s *Servidor) crearProducto(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.ProductoEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		producto, err := t.CrearProducto(entrada)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		w.Header().Set("Location", "/api/productos/"+url.PathEscape(producto.ID))
		escribirJSON(w, http.StatusCreated, producto)
	}
}

func (s *Servidor) obtenerProducto(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		producto, err := t.Producto(r.PathValue("id"))
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, producto)
	}
}

func (s *Servidor) actualizarProducto(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.ProductoEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		producto, err := t.ActualizarProducto(r.PathValue("id"), entrada)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, producto)
	}
}

func (s *Servidor) eliminarProducto(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := t.EliminarProducto(r.PathValue("id")); err != nil {
			responderErrorDominio(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Servidor) bajoStock(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limite := 5
		if valor := r.URL.Query().Get("limite"); valor != "" {
			convertido, err := strconv.Atoi(valor)
			if err != nil || convertido < 0 {
				escribirError(w, http.StatusBadRequest, "limite_invalido", "el límite debe ser un entero no negativo")
				return
			}
			limite = convertido
		}
		escribirJSON(w, http.StatusOK, t.ProductosBajoStock(limite))
	}
}

func (s *Servidor) listarCupones(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]any{"cupones": t.Cupones()})
	}
}

func (s *Servidor) verCarrito(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		carro, err := t.VerCarrito()
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, carro)
	}
}

func (s *Servidor) agregarCarrito(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.CantidadEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		carro, err := t.AgregarAlCarrito(entrada.ProductoID, entrada.Cantidad)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusCreated, carro)
	}
}

func (s *Servidor) cambiarCantidad(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.CantidadEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		carro, err := t.CambiarCantidadCarrito(r.PathValue("id"), entrada.Cantidad)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, carro)
	}
}

func (s *Servidor) eliminarCarrito(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		carro, err := t.EliminarDelCarrito(r.PathValue("id"))
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, carro)
	}
}

func (s *Servidor) aplicarCupon(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.CuponEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		carro, err := t.AplicarCupon(entrada.Codigo)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, carro)
	}
}

func (s *Servidor) listarClientes(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { escribirJSON(w, http.StatusOK, t.ListarClientes()) }
}

func (s *Servidor) crearCliente(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.ClienteEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		cliente, err := t.CrearCliente(entrada)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusCreated, cliente)
	}
}

func (s *Servidor) pedidosCliente(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, t.ListarPedidos(r.PathValue("email")))
	}
}

func (s *Servidor) listarPedidos(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		escribirJSON(w, http.StatusOK, t.ListarPedidos(r.URL.Query().Get("email")))
	}
}

func (s *Servidor) confirmarPedido(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.ClienteEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		compra, err := t.ConfirmarCompra(entrada)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		w.Header().Set("Location", "/api/pedidos/"+url.PathEscape(compra.Pedido.ID))
		escribirJSON(w, http.StatusCreated, compra)
	}
}

func (s *Servidor) obtenerPedido(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orden, err := t.Pedido(r.PathValue("id"))
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, orden)
	}
}

func (s *Servidor) cambiarEstado(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var entrada aplicacion.EstadoEntrada
		if err := decodificarJSON(w, r, &entrada); err != nil {
			escribirError(w, http.StatusBadRequest, "json_invalido", err.Error())
			return
		}
		orden, err := t.CambiarEstadoPedido(r.PathValue("id"), entrada.Estado)
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, orden)
	}
}

func (s *Servidor) cancelarPedido(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orden, err := t.CancelarPedido(r.PathValue("id"))
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, orden)
	}
}

func (s *Servidor) listarFacturas(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) { escribirJSON(w, http.StatusOK, t.ListarFacturas()) }
}

func (s *Servidor) obtenerFactura(t *aplicacion.Tienda) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		factura, err := t.Factura(r.PathValue("id"))
		if err != nil {
			responderErrorDominio(w, err)
			return
		}
		escribirJSON(w, http.StatusOK, factura)
	}
}

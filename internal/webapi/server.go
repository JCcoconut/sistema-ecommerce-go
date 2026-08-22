// Package webapi expone la aplicación mediante servicios REST y JSON.
package webapi

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/aplicacion"
)

//go:embed static/*
var archivosWeb embed.FS

type Servidor struct {
	handler  http.Handler
	contador atomic.Uint64
}

func NuevoServidor(tienda *aplicacion.Tienda) (*Servidor, error) {
	if tienda == nil {
		return nil, fmt.Errorf("la aplicación de tienda es obligatoria")
	}
	s := &Servidor{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health(tienda))
	mux.HandleFunc("GET /api/productos", s.listarProductos(tienda))
	mux.HandleFunc("POST /api/productos", s.crearProducto(tienda))
	mux.HandleFunc("GET /api/productos/bajo-stock", s.bajoStock(tienda))
	mux.HandleFunc("GET /api/productos/{id}", s.obtenerProducto(tienda))
	mux.HandleFunc("PUT /api/productos/{id}", s.actualizarProducto(tienda))
	mux.HandleFunc("DELETE /api/productos/{id}", s.eliminarProducto(tienda))
	mux.HandleFunc("GET /api/cupones", s.listarCupones(tienda))
	mux.HandleFunc("GET /api/carrito", s.verCarrito(tienda))
	mux.HandleFunc("POST /api/carrito/items", s.agregarCarrito(tienda))
	mux.HandleFunc("PUT /api/carrito/items/{id}", s.cambiarCantidad(tienda))
	mux.HandleFunc("DELETE /api/carrito/items/{id}", s.eliminarCarrito(tienda))
	mux.HandleFunc("POST /api/carrito/cupon", s.aplicarCupon(tienda))
	mux.HandleFunc("GET /api/clientes", s.listarClientes(tienda))
	mux.HandleFunc("POST /api/clientes", s.crearCliente(tienda))
	mux.HandleFunc("GET /api/clientes/{email}/pedidos", s.pedidosCliente(tienda))
	mux.HandleFunc("GET /api/pedidos", s.listarPedidos(tienda))
	mux.HandleFunc("POST /api/pedidos", s.confirmarPedido(tienda))
	mux.HandleFunc("GET /api/pedidos/{id}", s.obtenerPedido(tienda))
	mux.HandleFunc("PUT /api/pedidos/{id}/estado", s.cambiarEstado(tienda))
	mux.HandleFunc("POST /api/pedidos/{id}/cancelar", s.cancelarPedido(tienda))
	mux.HandleFunc("GET /api/facturas", s.listarFacturas(tienda))
	mux.HandleFunc("GET /api/facturas/{id}", s.obtenerFactura(tienda))

	sub, err := fs.Sub(archivosWeb, "static")
	if err != nil {
		return nil, fmt.Errorf("preparar archivos web: %w", err)
	}
	mux.Handle("GET /", http.FileServer(http.FS(sub)))

	// Los middlewares se aplican de adentro hacia afuera.
	s.handler = s.recuperarPanicos(s.encabezadosSeguros(s.registrar(mux)))
	return s, nil
}

func (s *Servidor) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

func (s *Servidor) encabezadosSeguros(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'")
		siguiente.ServeHTTP(w, r)
	})
}

func (s *Servidor) registrar(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inicio := time.Now()
		id := s.contador.Add(1)
		w.Header().Set("X-Request-ID", fmt.Sprintf("REQ-%06d", id))
		siguiente.ServeHTTP(w, r)
		log.Printf("request_id=REQ-%06d metodo=%s ruta=%s duracion=%s", id, r.Method, r.URL.Path, time.Since(inicio).Round(time.Millisecond))
	})
}

func (s *Servidor) recuperarPanicos(siguiente http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if valor := recover(); valor != nil {
				log.Printf("pánico recuperado: %v\n%s", valor, debug.Stack())
				escribirError(w, http.StatusInternalServerError, "error_interno", "ocurrió un error interno")
			}
		}()
		siguiente.ServeHTTP(w, r)
	})
}

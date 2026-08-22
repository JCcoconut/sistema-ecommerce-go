package webapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/aplicacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/facturacion"
)

func servidorPrueba(t *testing.T) (*aplicacion.Tienda, *httptest.Server) {
	t.Helper()
	tienda, err := aplicacion.NuevaTienda(nil, facturacion.EmisorSimulado{})
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NuevoServidor(tienda)
	if err != nil {
		t.Fatal(err)
	}
	servidor := httptest.NewServer(handler)
	t.Cleanup(servidor.Close)
	return tienda, servidor
}

func solicitar(t *testing.T, cliente *http.Client, metodo, ruta string, cuerpo []byte) *http.Response {
	t.Helper()
	peticion, err := http.NewRequest(metodo, ruta, bytes.NewReader(cuerpo))
	if err != nil {
		t.Fatal(err)
	}
	peticion.Header.Set("Content-Type", "application/json")
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		t.Fatal(err)
	}
	return respuesta
}

func TestAPIIntegracionYAceptacion(t *testing.T) {
	_, servidor := servidorPrueba(t)
	respuesta := solicitar(t, servidor.Client(), http.MethodPost, servidor.URL+"/api/carrito/items", []byte(`{"producto_id":"P001","cantidad":2}`))
	if respuesta.StatusCode != http.StatusCreated {
		cuerpo, _ := io.ReadAll(respuesta.Body)
		respuesta.Body.Close()
		t.Fatalf("agregar: %d %s", respuesta.StatusCode, cuerpo)
	}
	respuesta.Body.Close()
	respuesta = solicitar(t, servidor.Client(), http.MethodPost, servidor.URL+"/api/carrito/cupon", []byte(`{"codigo":"GO15"}`))
	if respuesta.StatusCode != http.StatusOK {
		t.Fatalf("cupón: %d", respuesta.StatusCode)
	}
	respuesta.Body.Close()
	cuerpo := []byte(`{"nombre":"Jorge","email":"jorge@correo.ec","ciudad":"Quito","sector":"Norte","referencia":"Casa azul"}`)
	respuesta = solicitar(t, servidor.Client(), http.MethodPost, servidor.URL+"/api/pedidos", cuerpo)
	defer respuesta.Body.Close()
	if respuesta.StatusCode != http.StatusCreated {
		datos, _ := io.ReadAll(respuesta.Body)
		t.Fatalf("checkout: %d %s", respuesta.StatusCode, datos)
	}
	var compra aplicacion.CompraDTO
	if err := json.NewDecoder(respuesta.Body).Decode(&compra); err != nil {
		t.Fatal(err)
	}
	if compra.Pedido.ID != "PED-001" || compra.Factura.PedidoID != "PED-001" {
		t.Fatalf("compra inesperada: %#v", compra)
	}

	respuesta = solicitar(t, servidor.Client(), http.MethodGet, servidor.URL+"/api/facturas/"+compra.Factura.ID, nil)
	if respuesta.StatusCode != http.StatusOK {
		t.Fatalf("factura: %d", respuesta.StatusCode)
	}
	respuesta.Body.Close()
}

func TestAPIErroresYJSONEstricto(t *testing.T) {
	_, servidor := servidorPrueba(t)
	respuesta := solicitar(t, servidor.Client(), http.MethodGet, servidor.URL+"/api/productos/NO-EXISTE", nil)
	if respuesta.StatusCode != http.StatusNotFound {
		t.Fatalf("esperado 404; obtenido %d", respuesta.StatusCode)
	}
	respuesta.Body.Close()
	respuesta = solicitar(t, servidor.Client(), http.MethodPost, servidor.URL+"/api/carrito/items", []byte(`{"producto_id":"P001","cantidad":1,"campo_desconocido":true}`))
	if respuesta.StatusCode != http.StatusBadRequest {
		t.Fatalf("esperado 400; obtenido %d", respuesta.StatusCode)
	}
	respuesta.Body.Close()
}

func TestAPIConcurrenciaProtegida(t *testing.T) {
	tienda, servidor := servidorPrueba(t)
	entrada := aplicacion.ProductoEntrada{ID: "PX", Nombre: "Producto concurrente", Categoria: "Pruebas", Precio: 1, Stock: 100, ProveedorNombre: "Proveedor", ProveedorEmail: "proveedor@correo.ec"}
	if _, err := tienda.CrearProducto(entrada); err != nil {
		t.Fatal(err)
	}
	const solicitudes = 30
	errores := make(chan error, solicitudes)
	var grupo sync.WaitGroup
	for i := 0; i < solicitudes; i++ {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			peticion, err := http.NewRequest(http.MethodPost, servidor.URL+"/api/carrito/items", bytes.NewBufferString(`{"producto_id":"PX","cantidad":1}`))
			if err != nil {
				errores <- err
				return
			}
			peticion.Header.Set("Content-Type", "application/json")
			respuesta, err := servidor.Client().Do(peticion)
			if err != nil {
				errores <- err
				return
			}
			defer respuesta.Body.Close()
			if respuesta.StatusCode != http.StatusCreated {
				errores <- fmt.Errorf("estado %d", respuesta.StatusCode)
			}
		}()
	}
	grupo.Wait()
	close(errores)
	for err := range errores {
		if err != nil {
			t.Fatal(err)
		}
	}
	carro, err := tienda.VerCarrito()
	if err != nil {
		t.Fatal(err)
	}
	if carro.CantidadTotal != solicitudes {
		t.Fatalf("se esperaban %d unidades; se obtuvieron %d", solicitudes, carro.CantidadTotal)
	}
}

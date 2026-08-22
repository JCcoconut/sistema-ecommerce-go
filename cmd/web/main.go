package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JCcoconut/sistema-ecommerce-go/internal/aplicacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/facturacion"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/persistencia"
	"github.com/JCcoconut/sistema-ecommerce-go/internal/webapi"
)

func main() {
	rutaDatos := valorEntorno("ECOMMERCE_DATA", "data/estado.json")
	puerto := valorEntorno("PORT", "8080")
	repositorio, err := persistencia.NuevoArchivoJSON(rutaDatos)
	if err != nil {
		log.Fatal(err)
	}
	tienda, err := aplicacion.NuevaTienda(repositorio, facturacion.EmisorSimulado{})
	if err != nil {
		log.Fatal(err)
	}
	handler, err := webapi.NuevoServidor(tienda)
	if err != nil {
		log.Fatal(err)
	}

	servidor := &http.Server{
		Addr: ":" + puerto, Handler: handler,
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	ctx, detener := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer detener()
	go func() {
		log.Printf("AudioCyber Store disponible en http://localhost:%s", puerto)
		if err := servidor.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("servidor HTTP: %v", err)
		}
	}()
	<-ctx.Done()
	contextoCierre, cancelar := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelar()
	if err := servidor.Shutdown(contextoCierre); err != nil {
		log.Printf("cierre incompleto: %v", err)
	}
}

func valorEntorno(nombre, defecto string) string {
	if valor := strings.TrimSpace(os.Getenv(nombre)); valor != "" {
		return valor
	}
	return defecto
}

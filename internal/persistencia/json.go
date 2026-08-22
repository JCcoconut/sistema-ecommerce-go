// Package persistencia guarda el estado del e-commerce en un archivo JSON.
package persistencia

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const maximoArchivo = 10 << 20 // 10 MiB

var ErrSinDatos = errors.New("todavía no existe un archivo de datos")

// Repositorio abstrae el mecanismo de almacenamiento utilizado por la aplicación.
type Repositorio interface {
	Cargar(destino any) error
	Guardar(origen any) error
}

// ArchivoJSON implementa persistencia local con escrituras atómicas.
type ArchivoJSON struct {
	ruta string
	mu   sync.Mutex
}

var _ Repositorio = (*ArchivoJSON)(nil)

func NuevoArchivoJSON(ruta string) (*ArchivoJSON, error) {
	if ruta == "" {
		return nil, errors.New("la ruta del archivo JSON es obligatoria")
	}
	return &ArchivoJSON{ruta: filepath.Clean(ruta)}, nil
}

func (a *ArchivoJSON) Cargar(destino any) error {
	if a == nil || destino == nil {
		return errors.New("el repositorio y el destino deben existir")
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	archivo, err := os.Open(a.ruta)
	if errors.Is(err, os.ErrNotExist) {
		return ErrSinDatos
	}
	if err != nil {
		return fmt.Errorf("abrir datos: %w", err)
	}
	defer archivo.Close()

	info, err := archivo.Stat()
	if err != nil {
		return fmt.Errorf("consultar archivo de datos: %w", err)
	}
	if info.Size() > maximoArchivo {
		return fmt.Errorf("el archivo de datos supera el límite de %d bytes", maximoArchivo)
	}

	decodificador := json.NewDecoder(io.LimitReader(archivo, maximoArchivo+1))
	decodificador.DisallowUnknownFields()
	if err := decodificador.Decode(destino); err != nil {
		return fmt.Errorf("decodificar datos JSON: %w", err)
	}
	var extra any
	if err := decodificador.Decode(&extra); !errors.Is(err, io.EOF) {
		return errors.New("el archivo JSON contiene más de un documento")
	}
	return nil
}

func (a *ArchivoJSON) Guardar(origen any) error {
	if a == nil || origen == nil {
		return errors.New("el repositorio y los datos deben existir")
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	directorio := filepath.Dir(a.ruta)
	if err := os.MkdirAll(directorio, 0o700); err != nil {
		return fmt.Errorf("crear directorio de datos: %w", err)
	}
	temporal, err := os.CreateTemp(directorio, ".estado-*.tmp")
	if err != nil {
		return fmt.Errorf("crear archivo temporal: %w", err)
	}
	nombreTemporal := temporal.Name()
	defer os.Remove(nombreTemporal)

	if err := temporal.Chmod(0o600); err != nil {
		temporal.Close()
		return fmt.Errorf("proteger archivo temporal: %w", err)
	}
	codificador := json.NewEncoder(temporal)
	codificador.SetIndent("", "  ")
	codificador.SetEscapeHTML(true)
	if err := codificador.Encode(origen); err != nil {
		temporal.Close()
		return fmt.Errorf("codificar datos JSON: %w", err)
	}
	if err := temporal.Sync(); err != nil {
		temporal.Close()
		return fmt.Errorf("sincronizar datos: %w", err)
	}
	if err := temporal.Close(); err != nil {
		return fmt.Errorf("cerrar archivo temporal: %w", err)
	}
	if err := os.Rename(nombreTemporal, a.ruta); err != nil {
		return fmt.Errorf("reemplazar archivo de datos: %w", err)
	}
	return nil
}

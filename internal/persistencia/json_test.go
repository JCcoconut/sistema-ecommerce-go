package persistencia

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestArchivoJSONGuardaYCarga(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "datos", "estado.json")
	repo, err := NuevoArchivoJSON(ruta)
	if err != nil {
		t.Fatal(err)
	}
	tipoEstado := struct {
		Nombre  string `json:"nombre"`
		Valores []int  `json:"valores"`
	}{}
	if err := repo.Cargar(&tipoEstado); !errors.Is(err, ErrSinDatos) {
		t.Fatalf("se esperaba ErrSinDatos; se obtuvo %v", err)
	}
	esperado := struct {
		Nombre  string `json:"nombre"`
		Valores []int  `json:"valores"`
	}{Nombre: "prueba", Valores: []int{1, 2, 3}}
	if err := repo.Guardar(esperado); err != nil {
		t.Fatal(err)
	}
	var obtenido struct {
		Nombre  string `json:"nombre"`
		Valores []int  `json:"valores"`
	}
	if err := repo.Cargar(&obtenido); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(esperado, obtenido) {
		t.Fatalf("esperado %#v, obtenido %#v", esperado, obtenido)
	}
	info, err := os.Stat(ruta)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("el archivo no debe conceder permisos a grupo/otros: %v", info.Mode().Perm())
	}
}

func TestArchivoJSONRechazaCamposDesconocidos(t *testing.T) {
	ruta := filepath.Join(t.TempDir(), "estado.json")
	if err := os.WriteFile(ruta, []byte(`{"nombre":"ok","intruso":true}`), 0o600); err != nil {
		t.Fatal(err)
	}
	repo, _ := NuevoArchivoJSON(ruta)
	var destino struct {
		Nombre string `json:"nombre"`
	}
	if err := repo.Cargar(&destino); err == nil {
		t.Fatal("se esperaba error por campo desconocido")
	}
}

package promocion

import (
	"errors"
	"reflect"
	"testing"
)

func TestImplementacionesDeInterfazDescuento(t *testing.T) {
	porcentaje, err := NuevoPorcentaje("Prueba", 0.15)
	if err != nil {
		t.Fatal(err)
	}

	casos := []struct {
		regla    Descuento
		esperado float64
	}{
		{regla: SinDescuento{}, esperado: 100},
		{regla: porcentaje, esperado: 85},
	}
	for _, caso := range casos {
		total, err := caso.regla.Aplicar(100)
		if err != nil {
			t.Fatal(err)
		}
		if total != caso.esperado {
			t.Fatalf("%s: se esperaba %.2f; se obtuvo %.2f", caso.regla.Nombre(), caso.esperado, total)
		}
	}
}

func TestGestorCuponesUsaMap(t *testing.T) {
	gestor, err := NuevoGestorCupones()
	if err != nil {
		t.Fatal(err)
	}
	esperados := []string{"AUDIO10", "GO15", "UIDE20"}
	if !reflect.DeepEqual(gestor.Codigos(), esperados) {
		t.Fatalf("códigos inesperados: %v", gestor.Codigos())
	}
	if _, err := gestor.Buscar("NOEXISTE"); !errors.Is(err, ErrCuponNoEncontrado) {
		t.Fatalf("se esperaba ErrCuponNoEncontrado; se obtuvo %v", err)
	}
}

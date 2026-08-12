// Package promocion define descuentos intercambiables mediante interfaces.
package promocion

import (
	"errors"
	"sort"
	"strings"
)

var ErrCuponNoEncontrado = errors.New("el cupón no existe")

// Descuento es una interfaz: cualquier tipo con estos métodos puede usarse
// como política de descuento sin modificar el módulo de pedidos.
type Descuento interface {
	Nombre() string
	Aplicar(subtotal float64) (float64, error)
}

// SinDescuento implementa la interfaz conservando el subtotal.
type SinDescuento struct{}

var _ Descuento = SinDescuento{}

func (SinDescuento) Nombre() string { return "Sin descuento" }

func (SinDescuento) Aplicar(subtotal float64) (float64, error) {
	if subtotal < 0 {
		return 0, errors.New("el subtotal no puede ser negativo")
	}
	return subtotal, nil
}

// Porcentaje implementa la interfaz aplicando un porcentaje encapsulado.
type Porcentaje struct {
	nombre     string
	porcentaje float64
}

var _ Descuento = (*Porcentaje)(nil)

// NuevoPorcentaje construye una política de descuento válida.
func NuevoPorcentaje(nombre string, porcentaje float64) (*Porcentaje, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		return nil, errors.New("el nombre del descuento es obligatorio")
	}
	if porcentaje <= 0 || porcentaje >= 1 {
		return nil, errors.New("el porcentaje debe estar entre 0 y 1")
	}
	return &Porcentaje{nombre: nombre, porcentaje: porcentaje}, nil
}

func (p Porcentaje) Nombre() string { return p.nombre }

func (p Porcentaje) Aplicar(subtotal float64) (float64, error) {
	if subtotal < 0 {
		return 0, errors.New("el subtotal no puede ser negativo")
	}
	return subtotal * (1 - p.porcentaje), nil
}

// GestorCupones encapsula el map que relaciona códigos con implementaciones
// de la interfaz Descuento.
type GestorCupones struct {
	cupones map[string]Descuento
}

// NuevoGestorCupones usa un array porque los códigos iniciales son fijos.
func NuevoGestorCupones() (*GestorCupones, error) {
	codigos := [3]string{"AUDIO10", "GO15", "UIDE20"}
	porcentajes := [3]float64{0.10, 0.15, 0.20}

	gestor := &GestorCupones{cupones: make(map[string]Descuento)}
	for i, codigo := range codigos {
		descuento, err := NuevoPorcentaje(codigo, porcentajes[i])
		if err != nil {
			return nil, err
		}
		gestor.cupones[codigo] = descuento
	}
	return gestor, nil
}

// Buscar retorna la interfaz correspondiente al código ingresado.
func (g GestorCupones) Buscar(codigo string) (Descuento, error) {
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	descuento, existe := g.cupones[codigo]
	if !existe {
		return nil, ErrCuponNoEncontrado
	}
	return descuento, nil
}

// Codigos devuelve un slice nuevo para no exponer el map interno.
func (g GestorCupones) Codigos() []string {
	resultado := make([]string, 0, len(g.cupones))
	for codigo := range g.cupones {
		resultado = append(resultado, codigo)
	}
	sort.Strings(resultado)
	return resultado
}

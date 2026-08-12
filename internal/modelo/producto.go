package modelo

import (
	"errors"
	"strings"
)

// Producto representa un artículo comercializado por la tienda.
// Los atributos privados solo se consultan o modifican mediante métodos.
type Producto struct {
	id        string
	nombre    string
	categoria string
	precio    float64
	stock     int
	proveedor Proveedor
}

// NuevoProducto funciona como constructor y evita crear productos inválidos.
func NuevoProducto(
	id, nombre, categoria string,
	precio float64,
	stock int,
	proveedor Proveedor,
) (*Producto, error) {
	id = strings.ToUpper(strings.TrimSpace(id))
	nombre = strings.TrimSpace(nombre)
	categoria = strings.TrimSpace(categoria)

	if id == "" || nombre == "" || categoria == "" {
		return nil, errors.New("el ID, nombre y categoría son obligatorios")
	}
	if precio <= 0 {
		return nil, errors.New("el precio debe ser mayor que cero")
	}
	if stock < 0 {
		return nil, errors.New("el stock no puede ser negativo")
	}
	if proveedor.Nombre() == "" || proveedor.Email() == "" {
		return nil, errors.New("el proveedor del producto no es válido")
	}

	return &Producto{
		id:        id,
		nombre:    nombre,
		categoria: categoria,
		precio:    precio,
		stock:     stock,
		proveedor: proveedor,
	}, nil
}

// Los siguientes métodos usan receptor por valor porque solo consultan datos.
func (p Producto) ID() string                   { return p.id }
func (p Producto) Nombre() string               { return p.nombre }
func (p Producto) Categoria() string            { return p.categoria }
func (p Producto) Precio() float64              { return p.precio }
func (p Producto) Stock() int                   { return p.stock }
func (p Producto) Proveedor() Proveedor         { return p.proveedor }
func (p Producto) Disponible(cantidad int) bool { return cantidad > 0 && cantidad <= p.stock }

// CalcularSubtotal es un método por valor: realiza un cálculo sin cambiar el objeto.
func (p Producto) CalcularSubtotal(cantidad int) (float64, error) {
	if cantidad <= 0 {
		return 0, errors.New("la cantidad debe ser mayor que cero")
	}
	return p.precio * float64(cantidad), nil
}

// DescontarStock usa receptor por puntero porque modifica el producto original.
func (p *Producto) DescontarStock(cantidad int) error {
	if p == nil {
		return errors.New("no se puede actualizar un producto nulo")
	}
	if !p.Disponible(cantidad) {
		return errors.New("stock insuficiente para completar la operación")
	}
	p.stock -= cantidad
	return nil
}

// ReponerStock usa receptor por puntero para incrementar el inventario.
func (p *Producto) ReponerStock(cantidad int) error {
	if p == nil {
		return errors.New("no se puede actualizar un producto nulo")
	}
	if cantidad <= 0 {
		return errors.New("la cantidad a reponer debe ser mayor que cero")
	}
	p.stock += cantidad
	return nil
}

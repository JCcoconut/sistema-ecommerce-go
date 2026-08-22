package modelo

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

var (
	ErrProductoInvalido  = errors.New("producto inválido")
	ErrStockInsuficiente = errors.New("stock insuficiente")
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

	if id == "" || nombre == "" || categoria == "" || len(id) > 32 || len(nombre) > 120 || len(categoria) > 60 {
		return nil, fmt.Errorf("%w: el ID, nombre y categoría son obligatorios", ErrProductoInvalido)
	}
	if precio <= 0 || math.IsNaN(precio) || math.IsInf(precio, 0) {
		return nil, fmt.Errorf("%w: el precio debe ser mayor que cero", ErrProductoInvalido)
	}
	if stock < 0 {
		return nil, fmt.Errorf("%w: el stock no puede ser negativo", ErrProductoInvalido)
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

// Actualizar aplica cambios controlados sin exponer los campos privados.
// El ID y el stock se gestionan por operaciones específicas para conservar
// la identidad del producto y evitar ajustes de inventario accidentales.
func (p *Producto) Actualizar(nombre, categoria string, precio float64, proveedor Proveedor) error {
	if p == nil {
		return fmt.Errorf("%w: no se puede actualizar un producto nulo", ErrProductoInvalido)
	}
	nombre = strings.TrimSpace(nombre)
	categoria = strings.TrimSpace(categoria)
	if nombre == "" || categoria == "" || len(nombre) > 120 || len(categoria) > 60 {
		return fmt.Errorf("%w: el nombre y la categoría son obligatorios", ErrProductoInvalido)
	}
	if precio <= 0 || math.IsNaN(precio) || math.IsInf(precio, 0) {
		return fmt.Errorf("%w: el precio debe ser mayor que cero", ErrProductoInvalido)
	}
	if proveedor.Nombre() == "" || proveedor.Email() == "" {
		return fmt.Errorf("%w: el proveedor no es válido", ErrProductoInvalido)
	}
	p.nombre = nombre
	p.categoria = categoria
	p.precio = precio
	p.proveedor = proveedor
	return nil
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
		return fmt.Errorf("%w: disponibles %d, solicitadas %d", ErrStockInsuficiente, p.stock, cantidad)
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

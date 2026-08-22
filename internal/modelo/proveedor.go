// Package modelo contiene los objetos principales del dominio del e-commerce.
package modelo

import (
	"errors"
	"net/mail"
	"strings"
)

func validarCorreo(email string) bool {
	if len(email) > 254 || strings.ContainsAny(email, "\r\n") {
		return false
	}
	direccion, err := mail.ParseAddress(email)
	return err == nil && direccion.Address == email
}

// Proveedor representa la empresa que abastece un producto.
// Sus campos son privados para aplicar encapsulación.
type Proveedor struct {
	nombre string
	email  string
}

// NuevoProveedor valida los datos y construye un proveedor válido.
func NuevoProveedor(nombre, email string) (Proveedor, error) {
	nombre = strings.TrimSpace(nombre)
	email = strings.TrimSpace(email)

	if nombre == "" || len(nombre) > 100 {
		return Proveedor{}, errors.New("el nombre del proveedor es obligatorio")
	}
	if !validarCorreo(email) {
		return Proveedor{}, errors.New("el correo del proveedor no es válido")
	}

	return Proveedor{nombre: nombre, email: email}, nil
}

// Nombre devuelve el nombre sin permitir modificar directamente el campo.
func (p Proveedor) Nombre() string {
	return p.nombre
}

// Email devuelve el correo del proveedor.
func (p Proveedor) Email() string {
	return p.email
}

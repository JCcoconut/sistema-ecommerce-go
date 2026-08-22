package modelo

import (
	"errors"
	"strings"
)

// DireccionEntrega es un struct que luego se anida dentro de Cliente.
type DireccionEntrega struct {
	ciudad     string
	sector     string
	referencia string
}

// NuevaDireccionEntrega construye una dirección válida.
func NuevaDireccionEntrega(ciudad, sector, referencia string) (DireccionEntrega, error) {
	ciudad = strings.TrimSpace(ciudad)
	sector = strings.TrimSpace(sector)
	referencia = strings.TrimSpace(referencia)
	if ciudad == "" || sector == "" || len(ciudad) > 80 || len(sector) > 80 || len(referencia) > 160 {
		return DireccionEntrega{}, errors.New("la ciudad y el sector son obligatorios")
	}
	return DireccionEntrega{ciudad: ciudad, sector: sector, referencia: referencia}, nil
}

func (d DireccionEntrega) Ciudad() string     { return d.ciudad }
func (d DireccionEntrega) Sector() string     { return d.sector }
func (d DireccionEntrega) Referencia() string { return d.referencia }

// Cliente demuestra el uso de un struct anidado.
type Cliente struct {
	nombre    string
	email     string
	direccion DireccionEntrega
}

// NuevoCliente valida y crea un cliente con su dirección anidada.
func NuevoCliente(nombre, email string, direccion DireccionEntrega) (Cliente, error) {
	nombre = strings.TrimSpace(nombre)
	email = strings.TrimSpace(email)
	if nombre == "" || len(nombre) > 100 {
		return Cliente{}, errors.New("el nombre del cliente es obligatorio")
	}
	if !validarCorreo(email) {
		return Cliente{}, errors.New("el correo del cliente no es válido")
	}
	if direccion.Ciudad() == "" || direccion.Sector() == "" {
		return Cliente{}, errors.New("la dirección del cliente no es válida")
	}
	return Cliente{nombre: nombre, email: email, direccion: direccion}, nil
}

func (c Cliente) Nombre() string              { return c.nombre }
func (c Cliente) Email() string               { return c.email }
func (c Cliente) Direccion() DireccionEntrega { return c.direccion }

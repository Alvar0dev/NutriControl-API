package models

import "time"

type Paciente struct {
	ID              string     `json:"id,omitempty"`
	Nombre          string     `json:"nombre"`
	Edad            int        `json:"edad"`
	Peso            float64    `json:"peso"`
	Altura          float64    `json:"altura,omitempty"`
	PorcentajeGrasa float64    `json:"porcentaje_grasa"`
	Notas           string     `json:"notas"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// ControlClinico representa un registro de evolución clínica periódica de un paciente.
type ControlClinico struct {
	ID              string    `json:"id,omitempty"`
	PacienteID      string    `json:"paciente_id"`
	Fecha           time.Time `json:"fecha"`
	Peso            float64   `json:"peso"`
	PorcentajeGrasa float64   `json:"porcentaje_grasa"`
	IMC             float64   `json:"imc,omitempty"`
	Notas           string    `json:"notas,omitempty"`
}

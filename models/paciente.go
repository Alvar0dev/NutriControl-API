package models

type Paciente struct {
	ID              string  `json:"id,omitempty"`
	Nombre          string  `json:"nombre"`
	Edad            int     `json:"edad"`
	Peso            float64 `json:"peso"`
	PorcentajeGrasa float64 `json:"porcentaje_grasa"`
	Notas           string  `json:"notas"`
}

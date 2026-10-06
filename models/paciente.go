package models

type Paciente struct {
	Nombre          string  `json:"nombre"`
	Edad            int     `json:"edad"`
	Peso            float64 `json:"peso"`
	PorcentajeGrasa float64 `json:"porcentaje_grasa"`
	Notas           string  `json:"notas"`
}

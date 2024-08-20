package repository

import "github.com/LucasWiman90/GoWebApp/internal/models"

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) error
}

package main

import (
	"time"

	"github.com/gofrs/uuid"
)

type Model interface {
}

type Application struct {
	Id               int      `db:"id"`
	Name             string   `db:"name"`
	Namespace        string   `db:"namespace"`
	DatabaseURL      string   `db:"database_url"`
	RestrictedTables []string `db:"restricted_tables"`
	HelmVersion      string   `db:"helm_version"`
	FluxVersion      string   `db:"flux_version"`
}

type Deployment struct {
	Id          uuid.UUID `db:"id"`
	AppId       int       `db:"app_id"`
	Status      string    `db:"status"`
	ImageURL    string    `db:"image_url"`
	Author      string    `db:"author"`
	Description string    `db:"description"`
	UpdatedAt   time.Time `db:"updated_at"`
	CreatedAt   time.Time `db:"created_at"`
}

type TTYAccessToken struct {
	Id          uuid.UUID `db:"id"`
	Token       string    `db:"token"`
	DatabaseURL string    `db:"database_url"`
	RequestId   uuid.UUID `db:"request_id"`
	TTL         int       `db:"ttl"`
	CreatedAt   time.Time `db:"created_at"`
}

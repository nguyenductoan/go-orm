package main

import (
	"reflect"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
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

type ApplicationRepo struct {
	Repository
}

func NewApplicationRepo(db *pgxpool.Pool) *ApplicationRepo {
	repository := Repository{
		Db:         db,
		TableName:  "applications",
		Model:      reflect.TypeOf(&Application{}),
		PrimaryKey: "id",
	}
	return &ApplicationRepo{repository}
}

type DeploymentRepo struct {
	Repository
}

func NewDeploymentRepo(db *pgxpool.Pool) *DeploymentRepo {
	repository := Repository{
		Db:         db,
		TableName:  "deployments",
		Model:      reflect.TypeOf(&Deployment{}),
		PrimaryKey: "id",
	}
	return &DeploymentRepo{repository}
}

type TTYAccessTokenRepo struct {
	Repository
}

func NewTTYAccessTokenRepo(db *pgxpool.Pool) *TTYAccessTokenRepo {
	repository := Repository{
		Db:         db,
		TableName:  "tty_access_tokens",
		Model:      reflect.TypeOf(&TTYAccessToken{}),
		PrimaryKey: "id",
	}
	return &TTYAccessTokenRepo{repository}
}

package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	dbURL := "postgres://admin:admin@127.0.0.1:5432/hero2"
	connString := fmt.Sprintf("%s?sslmode=%s&pool_max_conns=%d", dbURL, "disable", 4)
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		panic(err)
	}
	db, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		panic(err)
	}

	applicationRepo := NewApplicationRepo(db)
	deploymentRepo := NewDeploymentRepo(db)

	fmt.Printf("** Test List method\n")
	records, err := applicationRepo.List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, r := range records {
		a := r.(*Application)
		fmt.Printf("Name: %s; Namespace: %s; FluxVersion: %s; RestrictedTables: %v\n", a.Name, a.Namespace, a.FluxVersion, a.RestrictedTables)
	}

	fmt.Printf("\nTest Find method\n")
	record, err := applicationRepo.Find(context.Background(), "2")
	if err != nil {
		panic(err)
	}
	recordFind := record.(*Application)
	fmt.Printf("Id: %d, Name: %s; Namespace: %s; FluxVersion: %s\n", recordFind.Id, recordFind.Name, recordFind.Namespace, recordFind.FluxVersion)

	fmt.Printf("\n** Test Add method\n")
	newApp := &Application{
		Name:             "new-app",
		Namespace:        "staging",
		DatabaseURL:      "",
		RestrictedTables: []string{"a", "b"},
	}
	record, err = applicationRepo.Add(context.Background(), newApp)
	if err != nil {
		fmt.Printf("Add new record error: %v\n", err)
	} else {
		recordAdd := record.(*Application)
		fmt.Printf("Name: %s; Namespace: %s; FluxVersion: %s\n", recordAdd.Name, recordAdd.Namespace, recordAdd.FluxVersion)
	}

	fmt.Printf("\n** Test Update method: recordId: %d\n", recordFind.Id)
	if recordFind.HelmVersion == "v2" {
		recordFind.HelmVersion = "v3"
	} else {
		recordFind.HelmVersion = "v2"
	}
	record, err = applicationRepo.Update(context.Background(), recordFind)
	if err != nil {
		fmt.Printf("Update record error: %v\n", err)
	} else {
		recordUpdate := record.(*Application)
		fmt.Printf("Id: %d, Name: %s; Namespace: %s; HelmVersion: %s\n", recordUpdate.Id, recordUpdate.Name, recordUpdate.Namespace, recordUpdate.HelmVersion)
	}

	fmt.Printf("\n**Test List method For TTYAccessToken\n")
	deploymentRecords, err := deploymentRepo.List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, r := range deploymentRecords {
		d := r.(*Deployment)
		fmt.Printf("Id: %s; AppId: %d; Status: %s; UpdatedAt: %v\n", d.Id, d.AppId, d.Status, d.UpdatedAt)
	}

	fmt.Printf("\n**Test Find and Update method For Deployment\n")
	recordFindToken, err := deploymentRepo.Find(context.Background(), "f66b66a2-18c7-43ce-8b55-a2a9991e436e")
	if err != nil {
		panic(err)
	}
	record, err = deploymentRepo.Update(context.Background(), recordFindToken)
	if err != nil {
		panic(err)
	}
	d := record.(*Deployment)
	fmt.Printf("Id: %s; AppId: %d; Status: %s; UpdatedAt: %v\n", d.Id, d.AppId, d.Status, d.UpdatedAt)
}

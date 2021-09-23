package main

import (
	"context"
	"fmt"
	"reflect"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
)

func main() {
	dbURL := "postgres://admin:admin@127.0.0.1:5432/hero2"
	connString := fmt.Sprintf("%s?sslmode=%s&pool_max_conns=%d", dbURL, "disable", 3)
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		panic(err)
	}
	db, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		panic(err)
	}

	applicationRepo := NewApplicationRepo(db)

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
	deploymentRepo := NewDeploymentRepo(db)
	deploymentRecords, err := deploymentRepo.List(context.Background())
	if err != nil {
		panic(err)
	}
	for _, r := range deploymentRecords {
		d := r.(*Deployment)
		fmt.Printf("Id: %s; AppId: %d; Status: %s; UpdatedAt: %v\n", d.Id, d.AppId, d.Status, d.UpdatedAt)
	}
}

func parseArr(arrValue interface{}) (result []reflect.Value) {
	val := reflect.ValueOf(arrValue)
	switch val.Kind() {
	case reflect.Struct:
		switch arrValue.(type) {
		case pgtype.TextArray:
			arr := arrValue.(pgtype.TextArray)
			elems := arr.Elements
			for _, elem := range elems {
				v, err := elem.Value()
				if v != nil && err == nil {
					result = append(result, reflect.ValueOf(v))
				}
			}
		default:
			fmt.Printf("WARN: type %T is not supported\n", arrValue)
		}
	case reflect.Ptr:
		switch arrValue.(type) {
		case *pgtype.TextArray:
			arr := arrValue.(*pgtype.TextArray)
			elems := arr.Elements
			for _, elem := range elems {
				v, err := elem.Value()
				if v != nil && err == nil {
					result = append(result, reflect.ValueOf(v))
				}
			}
		default:
			fmt.Printf("WARN: type %T is not supported\n", arrValue)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			result = append(result, val.Index(i))
		}
	default:
		fmt.Printf("value: %v\n", arrValue)
		fmt.Printf("WARN: kind %T is not supported\n", arrValue)
	}
	return
}

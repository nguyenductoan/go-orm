package main

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	DATABASE_TAG string = "db"
)

type RowInfo map[string]interface{}

type Repository struct {
	Db         *pgxpool.Pool
	TableName  string
	Model      reflect.Type
	PrimaryKey string
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

func (repo Repository) PrimaryKeyValue(record Model) string {
	v := reflect.ValueOf(record).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		if fieldInfo.Tag.Get(DATABASE_TAG) == repo.PrimaryKey {
			valueField := v.Field(i)
			val := reflect.ValueOf(valueField.Interface())
			return reflectValue(val).(string)
		}
	}
	return ""
}

func (repo Repository) List(ctx context.Context) (records []interface{}, err error) {
	stm := fmt.Sprintf("SELECT * FROM %s", repo.TableName)
	rows, err := repo.Db.Query(ctx, stm)

	if err != nil {
		return records, fmt.Errorf("exec query error: %v", err)
	}

	for rows.Next() {
		rowInfo, extractErr := extractRowInfo(rows)
		if extractErr != nil {
			err = fmt.Errorf("extract row info error: %v", extractErr)
		}
		record, initErr := repo.initRecord(rowInfo)
		if initErr != nil {
			err = fmt.Errorf("init record error: %v", initErr)
			return
		}
		records = append(records, record)
	}
	return
}

func (repo Repository) Find(ctx context.Context, primaryKey string) (record Model, err error) {
	if primaryKey == "" {
		return nil, errors.New("primaryKey is empty")
	}
	stm := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", repo.TableName, repo.PrimaryKey)
	rows, err := repo.Db.Query(ctx, stm, primaryKey)
	if !rows.Next() {
		if rows.Err() == nil {
			err = pgx.ErrNoRows
		} else {
			err = rows.Err()
		}
		return
	}

	rowInfo, extractErr := extractRowInfo(rows)
	if extractErr != nil {
		err = fmt.Errorf("extract row info error: %v", extractErr)
	}
	record, initErr := repo.initRecord(rowInfo)
	if initErr != nil {
		err = fmt.Errorf("init record error: %v", initErr)
		return
	}
	return
}

func (repo Repository) Add(ctx context.Context, record Model) (result Model, err error) {
	v := reflect.ValueOf(record).Elem()
	names := []string{}
	fieldValues := []interface{}{}
	fieldIndex := []string{}
	counter := 1
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		valueField := v.Field(i)
		tag := fieldInfo.Tag
		if tag.Get(DATABASE_TAG) == repo.PrimaryKey {
			continue
		}
		names = append(names, tag.Get(DATABASE_TAG))
		val := reflect.ValueOf(valueField.Interface())
		fieldValues = append(fieldValues, reflectValue(val))

		fieldIndex = append(fieldIndex, fmt.Sprintf("$%d", counter))
		counter = counter + 1
	}
	fieldNameString := strings.Join(names, ", ")
	fieldIndexString := strings.Join(fieldIndex, ", ")

	stm := fmt.Sprintf("INSERT INTO %s (%s) values(%s) returning %s", repo.TableName, fieldNameString, fieldIndexString, fieldNameString)
	rows, err := repo.Db.Query(ctx, stm, fieldValues...)
	if !rows.Next() {
		if rows.Err() == nil {
			err = pgx.ErrNoRows
		} else {
			err = fmt.Errorf("exec query error: %v\nstm: %v\n", rows.Err(), stm)
		}
		return
	}
	rowInfo, extractErr := extractRowInfo(rows)
	if extractErr != nil {
		err = fmt.Errorf("extract row info error: %v", extractErr)
	}
	record, initErr := repo.initRecord(rowInfo)
	if initErr != nil {
		err = fmt.Errorf("init record error: %v", initErr)
		return
	}
	return
}

func (repo Repository) Update(ctx context.Context, updateRecord Model) (record Model, err error) {
	v := reflect.ValueOf(updateRecord).Elem()
	fieldNames := []string{}
	fieldValues := []interface{}{}
	assignment := []string{}
	counter := 1
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag
		valueField := v.Field(i)
		val := reflect.ValueOf(valueField.Interface())
		if tag.Get(DATABASE_TAG) == repo.PrimaryKey {
			continue
		}

		fieldNames = append(fieldNames, tag.Get(DATABASE_TAG))
		fieldValues = append(fieldValues, reflectValue(val))
		assignment = append(assignment, fmt.Sprintf("%s=$%d", tag.Get(DATABASE_TAG), counter))
		counter = counter + 1
	}
	fieldNameString := repo.PrimaryKey + ", " + strings.Join(fieldNames, ", ")
	assignmentString := strings.Join(assignment, ", ")

	primaryKeyVal := repo.PrimaryKeyValue(updateRecord)
	stm := fmt.Sprintf("UPDATE %s SET %s WHERE %s=%s returning %s", repo.TableName, assignmentString, repo.PrimaryKey, primaryKeyVal, fieldNameString)
	fmt.Printf("stm: %v\n", stm)
	rows, err := repo.Db.Query(ctx, stm, fieldValues...)
	if !rows.Next() {
		if rows.Err() == nil {
			err = pgx.ErrNoRows
		} else {
			err = fmt.Errorf("exec query error: %v\nstm: %v\n", rows.Err(), stm)
		}
		return
	}
	rowInfo, extractErr := extractRowInfo(rows)
	if extractErr != nil {
		err = fmt.Errorf("extract row info error: %v", extractErr)
	}
	record, initErr := repo.initRecord(rowInfo)
	if initErr != nil {
		err = fmt.Errorf("init record error: %v", initErr)
		return
	}
	return
}

func (repo Repository) Delete(ctx context.Context, primaryKey string) (err error) {
	stm := fmt.Sprintf("DELETE FROM %s where %s=$1", repo.TableName, repo.PrimaryKey)
	_, err = repo.Db.Exec(ctx, stm, primaryKey)
	if err != nil {
		err = fmt.Errorf("delete record error: %v", err)
	}
	return
}

func populate(v reflect.Value, value interface{}) error {
	switch v.Kind() {
	case reflect.String:
		if _, ok := value.(string); ok {
			v.SetString(value.(string))
		} else {
			v.SetString("invalid")
		}
	case reflect.Slice:
		arr := parseArr(value)
		for _, elem := range arr {
			v.Set(reflect.Append(v, elem))
		}
	case reflect.Array:
		v.Set(reflect.ValueOf(value))
	case reflect.Int:
		i, err := strconv.ParseInt(fmt.Sprintf("%d", value), 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Bool:
		v.SetBool(value.(bool))
	case reflect.Struct:
		v.Set(reflect.ValueOf(value))
	case reflect.Ptr:
	default:
		return fmt.Errorf("unsupported kind %s", v.Type())
	}
	return nil
}

func extractRowInfo(rows pgx.Rows) (info RowInfo, err error) {
	info = make(RowInfo)
	fieldValues, err := rows.Values()
	if err != nil {
		err = fmt.Errorf("extract row value error: %v", err)
		return
	}
	for i, desc := range rows.FieldDescriptions() {
		value := fieldValues[i]
		if value == nil {
			continue
		}
		info[string(desc.Name)] = value
	}
	return
}

func (repo Repository) initRecord(rowInfo RowInfo) (record interface{}, err error) {
	record = reflect.New(repo.Model.Elem()).Interface()
	v := reflect.ValueOf(record).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		tag := fieldInfo.Tag
		name := tag.Get(DATABASE_TAG) // a reflect.StructTag
		f := v.Field(i)
		value, ok := rowInfo[name]
		if !ok {
			continue
		}
		err = populate(f, value)
		if err != nil {
			err = fmt.Errorf("populate value to field error: %v", err)
			return nil, err
		}
	}
	return record, nil
}


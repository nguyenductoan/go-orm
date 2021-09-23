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

type RecordInfo struct {
	FieldTagNames []string
	FieldValues   []reflect.Value
}

type Repository struct {
	Db         *pgxpool.Pool
	TableName  string
	Model      reflect.Type
	PrimaryKey string
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
	stm, fieldValues := repo.InsertStatement(record)
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
	stm, fieldValues := repo.UpdateStatement(updateRecord)
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

func extractRecordInfo(record Model) (info RecordInfo) {
	v := reflect.ValueOf(record).Elem()
	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		valueField := v.Field(i)
		val := reflect.ValueOf(valueField.Interface())

		info.FieldTagNames = append(info.FieldTagNames, fieldInfo.Tag.Get(DATABASE_TAG))
		info.FieldValues = append(info.FieldValues, val)
	}
	return
}

func (repo Repository) UpdateStatement(record Model) (stm string, values []interface{}) {
	recordInfo := extractRecordInfo(record)
	primaryKeyVal := repo.PrimaryKeyValue(record)
	assignment := []string{}

	for i, name := range recordInfo.FieldTagNames {
		assignment = append(assignment, fmt.Sprintf("%s=$%d", name, i+1))
	}
	assignmentStm := strings.Join(assignment, ", ")
	fieldNameStm := strings.Join(recordInfo.FieldTagNames, ", ")

	for _, val := range recordInfo.FieldValues {
		values = append(values, reflectValue(val))
	}
	stm = fmt.Sprintf("UPDATE %s SET %s WHERE %s=%s returning %s", repo.TableName, assignmentStm, repo.PrimaryKey, primaryKeyVal, fieldNameStm)
	return
}

func (repo Repository) InsertStatement(record Model) (stm string, values []interface{}) {
	recordInfo := extractRecordInfo(record)

	fieldNames := []string{}
	fieldIndex := []string{}
	for i, name := range recordInfo.FieldTagNames {
		fieldNames = append(fieldNames, name)
		fieldIndex = append(fieldIndex, fmt.Sprintf("$%d", i+1))
	}
	fieldNameStm := strings.Join(recordInfo.FieldTagNames, ", ")
	fieldIndexStm := strings.Join(fieldIndex, ", ")
	for _, val := range recordInfo.FieldValues {
		values = append(values, reflectValue(val))
	}

	stm = fmt.Sprintf("INSERT INTO %s (%s) values(%s) returning %s", repo.TableName, fieldNameStm, fieldIndexStm, fieldNameStm)
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

func reflectValue(val reflect.Value) (result interface{}) {
	switch val.Kind() {
	case reflect.Slice, reflect.Array:
		return []string{}
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Struct:
	default:
		panic(errors.New("convert reflectValue error"))
	}
	return
}

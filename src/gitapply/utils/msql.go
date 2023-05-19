package utils

import (
	"database/sql"
	"errors"
	"log"
	"strings"
)

// [0]map string,[1]map string
func SelectMagicFromDb(tx *sql.DB, selectForDb string, args ...interface{}) ([][]string, []map[string]string, error) {
	// Get Data
	data_text := make([][]string, 0)
	data_text_map := make([]map[string]string, 0)
	rows, err := tx.Query(selectForDb, args...)
	if err != nil {
		log.Printf("SelectMagicFromDb: %v, %v failed", selectForDb, args)
		return data_text, data_text_map, err
	}
	defer rows.Close()

	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		return data_text, data_text_map, err
	}
	if len(columns) == 0 {
		return data_text, data_text_map, errors.New("No columns in table " + selectForDb + ".")
	}

	// Read data

	for rows.Next() {
		// Init temp data storage

		data := make([]*sql.NullString, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range data {
			ptrs[i] = &data[i]
		}

		// Read data
		if err := rows.Scan(ptrs...); err != nil {
			return data_text, data_text_map, err
		}
		dataStrings := make([]string, len(columns))
		dataMap := make(map[string]string, len(columns))
		for key, value := range data {
			if value != nil && value.Valid {
				rune := `\'`
				dataStrings[key] = strings.Replace(value.String, "'", rune, -1)
				dataMap[columns[key]] = strings.Replace(value.String, "'", rune, -1)
			} else {
				dataStrings[key] = "null"
			}
		}

		data_text = append(data_text, dataStrings)
		data_text_map = append(data_text_map, dataMap)
		// fmt.Println("data_text_map db :", data_text_map)

	}

	log.Printf("SelectMagicFromDb: %v, %v success", selectForDb, args)
	// return strings.Join(data_text, ","), rows.Err()
	return data_text, data_text_map, rows.Err()
}

func DoExecDb(tx *sql.DB, doSqlStr string, args ...interface{}) error {

	_, err := tx.Exec(doSqlStr, args...)
	if err != nil {
		log.Printf("doExecDb: %s,args: %s error:%s \n", doSqlStr, args, err.Error())
		return err
	}
	log.Printf("doExecDb: %v, %v success", doSqlStr, args)
	return nil
}

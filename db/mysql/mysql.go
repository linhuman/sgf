package mysql

import (
	"database/sql"
	"fmt"
	"reflect"
	"sgf/config"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var instance map[string]*sql.DB
var once sync.Once

const INSERT_INTO = 1
const INSERT_REPLACE = 2
const INSERT_IGNORE = 3
const INSERT_ORUPDATE = 4

type mysql struct {
	db     *sql.DB
	report bool
}

func initialize() {
	instance = make(map[string]*sql.DB)
	for k, _ := range config.Entity.Db {
		dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", config.Entity.Db[k].User, config.Entity.Db[k].Password, config.Entity.Db[k].Host, config.Entity.Db[k].Port, config.Entity.Db[k].Database)
		instance[k], _ = sql.Open(config.Entity.Db[k].Driver, dsn)
		instance[k].SetMaxOpenConns(config.Entity.Db[k].MaxOpenConns)
		instance[k].SetMaxIdleConns(config.Entity.Db[k].MaxIdleConns)
		instance[k].SetConnMaxLifetime(config.Entity.Db[k].ConnMaxLiftetime)
		err := instance[k].Ping()
		if nil != err {
			panic(err)
		}
	}
}
func (m *mysql) Report(bl bool) {
	m.report = bl
}

func GetInstance(field ...interface{}) *mysql {
	once.Do(initialize)
	obj := new(mysql)
	pool_name := "default"
	if 0 < len(field) {
		pool_name = field[0].(string)
	}
	if nil == instance[pool_name] {
		panic("数据库连接[" + pool_name + "]不存在")
	}
	obj.db = instance[pool_name]

	return obj
}

//插入一条数据
func (m *mysql) AutoUpdate(table string, args ...map[string]interface{}) (int64, error, string) {

	var vals []interface{}
	where_str := ""
	set_str := ""
	if len(args) > 0 && len(args[0]) > 0 {
		set_map := args[0]
		var build strings.Builder
		if len(set_map) > 0 {
			for k, v := range set_map {
				if reflect.TypeOf(v).Kind() == reflect.Slice {
					l := v.([]interface{})

					if "+" == l[0] || "-" == l[0] {
						build.WriteString(set_str)
						build.WriteString("`")
						build.WriteString(k)
						build.WriteString("`=")
						build.WriteString(k)
						build.WriteString(l[0].(string))
						build.WriteString("?")
						/*
							switch l[1].(type) {
							case string:
								build.WriteString(l[1].(string))
							case int:
								build.WriteString(strconv.Itoa(l[1].(int)))
							default:
								build.WriteString(l[1].(string))
							}
						*/
						build.WriteString(", ")
						vals = append(vals, l[1])
					}

				} else {
					build.WriteString(set_str)
					build.WriteString("`")
					build.WriteString(k)
					build.WriteString("`=?, ")
					vals = append(vals, v)
				}

			}
		}
		set_str = build.String()
		set_str = strings.Trim(set_str, ", ")
	}
	if len(args) > 1 && len(args[1]) > 0 {
		where_map := args[1]
		var build strings.Builder
		if len(where_map) > 0 {
			for k, v := range where_map {
				build.WriteString(where_str)
				build.WriteString("`")
				build.WriteString(k)
				build.WriteString("`=? AND ")
				vals = append(vals, v)
			}
			where_str = strings.Trim(build.String(), "AND ")
		}
	}
	sql := table
	var build strings.Builder

	if "" != set_str && "" != where_str {
		build.WriteString("UPDATE ")
		build.WriteString(sql)
		build.WriteString(" SET ")
		build.WriteString(set_str)
		build.WriteString(" WHERE ")
		build.WriteString(where_str)
		sql = build.String()
	} else if "" != where_str {
		build.WriteString(sql)
		build.WriteString(" WHERE ")
		build.WriteString(where_str)
		sql = build.String()

	}
	result, err, report := m.Exec(sql, vals...)
	if nil != err {
		return 0, err, report
	}
	rows_affected, err := result.RowsAffected()
	return rows_affected, err, report

}

func (m *mysql) AutoInsert(table string, args map[string]interface{}, cmd int, update ...map[string]interface{}) (int64, int64, error, string) {
	statement := ""
	if len(args) == 0 {
		return 0, 0, fmt.Errorf("%s", "args参数不存在"), ""
	}
	keys_str := ""
	values_str := ""
	var vals []interface{}
	for k, v := range args {
		keys_str = keys_str + "`" + k + "`, "
		values_str = values_str + "?, "
		vals = append(vals, v)
	}
	keys_str = strings.Trim(keys_str, ", ")
	values_str = strings.Trim(values_str, ", ")
	if INSERT_INTO == cmd {
		statement = "INSERT INTO " + table + " (" + keys_str + ")" + " VALUES (" + values_str + ")"
	} else if INSERT_REPLACE == cmd {
		statement = "REPLACE INTO " + table + " (" + keys_str + ")" + " VALUES (" + values_str + ")"
	} else if INSERT_IGNORE == cmd {
		statement = "INSERT IGNORE INTO " + table + " (" + keys_str + ")" + " VALUES (" + values_str + ")"
	} else if INSERT_ORUPDATE == cmd {
		update_string := ""
		if 0 == len(update) {
			return 0, 0, fmt.Errorf("%s", "update参数不存在"), ""
		}
		for k, v := range update[0] {
			if reflect.TypeOf(v).Kind() == reflect.Slice {
				l := v.([]interface{})
				if "+" == l[0] || "-" == l[0] {
					update_string = update_string + "`" + k + "`=k" + l[0].(string) + "?, "
					vals = append(vals, l[1])
				}
			} else {
				update_string = update_string + "`" + k + "`=?, "
				vals = append(vals, v)
			}
		}
		update_string = " ON DUPLICATE KEY UPDATE " + strings.Trim(update_string, ", ")
		statement = "INSERT INTO " + table + " (" + keys_str + ")" + " VALUES (" + values_str + ")" + update_string
	}
	result, err, report := m.Exec(statement, vals...)
	if nil != err {
		return 0, 0, err, report
	}
	last_insert_id, err := result.LastInsertId()
	rows_affected, err := result.RowsAffected()

	return last_insert_id, rows_affected, err, report
}
func (m *mysql) query(query_sql string, args ...interface{}) ([]map[string]string, error, string) {
	report := ""
	start_time := time.Now().UnixNano() / 1e6
	rows, err := m.db.Query(query_sql, args...)
	if true == m.report {
		sql_str := genSql(query_sql, args...)
		end_time := time.Now().UnixNano() / 1e6
		elapsed_time := end_time - start_time
		var build strings.Builder
		build.WriteString("[SQL]")
		build.WriteString(sql_str)
		build.WriteString(" [time]")
		build.WriteString(strconv.FormatInt(elapsed_time, 10))
		build.WriteString("ms")
		report = build.String()
	}
	result := make([]map[string]string, 0)
	if nil != err {
		return result, err, report
	} else {
		result, err = getQueryResult(rows)
	}
	return result, err, report
}
func (m *mysql) Exec(exec_sql string, args ...interface{}) (sql.Result, error, string) {
	report := ""
	start_time := time.Now().UnixNano() / 1e6
	result, err := m.db.Exec(exec_sql, args...)
	if true == m.report {
		sql_str := genSql(exec_sql, args...)
		end_time := time.Now().UnixNano() / 1e6
		elapsed_time := end_time - start_time
		var build strings.Builder
		build.WriteString("[SQL]")
		build.WriteString(sql_str)
		build.WriteString(" [time]")
		build.WriteString(strconv.FormatInt(elapsed_time, 10))
		build.WriteString("ms")
		report = build.String()
	}
	return result, err, report
}

//查找一行数据
func (m *mysql) GetRow(query_str string, args ...[]interface{}) (map[string]string, error, string) {
	var result []map[string]string
	var err error
	var report string
	if 0 == len(args) {
		result, err, report = m.query(query_str)
	} else {
		result, err, report = m.query(query_str, args[0]...)
	}
	if 0 == len(result) {
		return map[string]string{}, err, report
	}
	return result[0], err, report

}

//查询所有数据
func (m *mysql) GetAll(query_str string, args ...[]interface{}) ([]map[string]string, error, string) {
	if 0 == len(args) {
		return m.query(query_str)
	} else {
		return m.query(query_str, args[0]...)
	}
}

//查询一列数据
func (m *mysql) GetCol(query_str string, args ...[]interface{}) (string, error, string) {
	var query_rs []map[string]string
	var val string
	var err error
	var report string
	if 0 == len(args) {
		query_rs, err, report = m.query(query_str)
	} else {
		query_rs, err, report = m.query(query_str, args[0]...)
	}
	if len(query_rs) > 0 {
		for _, v := range query_rs[0] {
			val = v
			break
		}
	}

	return val, err, report
}

//查询列所有数据
func (m *mysql) GetCols(query_str string, args ...[]interface{}) ([]string, error, string) {
	var query_rs []map[string]string
	var err error
	var report string
	if 0 == len(args) {
		query_rs, err, report = m.query(query_str)
	} else {
		query_rs, err, report = m.query(query_str, args[0]...)
	}
	var rs_arr = make([]string, 0)
	if len(query_rs) > 0 {
		for _, v := range query_rs {
			for _, v1 := range v {
				rs_arr = append(rs_arr, v1)
				continue
			}
		}
	}
	return rs_arr, err, report

}
func getQueryResult(rows *sql.Rows) ([]map[string]string, error) {
	defer rows.Close()
	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	scan_args := make([]interface{}, len(values))
	for i := range values {
		scan_args[i] = &values[i]
	}
	var value string
	var err error
	var result []map[string]string
	for rows.Next() {
		one := make(map[string]string)
		err = rows.Scan(scan_args...)
		if nil != err {
			return result, err
		}
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)

			}
			one[columns[i]] = value

		}
		result = append(result, one)

	}

	return result, err
}
func genSql(sql_fmt string, args ...interface{}) string {
	var build strings.Builder
	pos := 0
	if 0 < len(args) {
		for _, v := range args {
			relative_pos := strings.Index(sql_fmt[pos:], "?")
			if -1 == relative_pos {
				return sql_fmt
			}
			build.WriteString(sql_fmt[pos:(pos + relative_pos)])
			switch v.(type) {
			case int:
				build.WriteString(strconv.Itoa(v.(int)))
				relative_pos++
			case int64:
				build.WriteString(strconv.FormatInt(v.(int64), 10))
				relative_pos++
			case int32:
				build.WriteString(strconv.FormatInt(int64(v.(int32)), 10))
				relative_pos++
			case int16:
				build.WriteString(strconv.FormatInt(int64(v.(int16)), 10))
				relative_pos++
			case int8:
				build.WriteString(strconv.FormatInt(int64(v.(int8)), 10))
				relative_pos++
			case uint:
				build.WriteString(strconv.FormatUint(uint64(v.(uint)), 10))
				relative_pos++
			case uint8:
				build.WriteString(strconv.FormatUint(uint64(v.(uint8)), 10))
				relative_pos++
			case uint16:
				build.WriteString(strconv.FormatUint(uint64(v.(uint16)), 10))
				relative_pos++
			case uint32:
				build.WriteString(strconv.FormatUint(uint64(v.(uint32)), 10))
				relative_pos++
			case uint64:
				build.WriteString(strconv.FormatUint(v.(uint64), 10))
				relative_pos++
			case float64:
				build.WriteString(strconv.FormatFloat(v.(float64), 'f', -1, 64))
				relative_pos++
			case float32:
				build.WriteString(strconv.FormatFloat(float64(v.(float32)), 'f', -1, 32))
				relative_pos++
			case string:
				build.WriteString("\"" + v.(string) + "\"")
				relative_pos++
			default:
				build.WriteString("\"" + v.(string) + "\"")
				relative_pos++
			}
			pos = pos + relative_pos

		}

		build.WriteString(sql_fmt[pos:])
	} else {
		build.WriteString(sql_fmt)
	}

	return build.String()
}

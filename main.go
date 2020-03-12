package zyorm

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"reflect"
	"strings"
	"sync"
)

type Engine struct {
	Db *sql.DB

	RWMuTables *sync.RWMutex
	Tables map[string]TableInfo

}

type DnsConf struct {
	Username string
	Password string
	Ip string
	Port string
	TableName string
	ParamsStr string
}

func NewEngine(dnsConf DnsConf) (*Engine, error) {

	dsn := dnsConf.Username + ":" + dnsConf.Password + "@tcp(" + dnsConf.Ip + ":" + dnsConf.Port + ")/" + dnsConf.TableName + "?" + dnsConf.ParamsStr

	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}


	//设置数据库空闲连接
	db.SetMaxIdleConns(20)
	//设置最大打开数量
	db.SetMaxOpenConns(20)

	//直接判断是不是能连接成功
	err = db.Ping()
	if err != nil {
		log.Printf("ping error: %s", err)

		return nil, err
	}

	engine := &Engine{Db: db, RWMuTables:new(sync.RWMutex), Tables: make(map[string]TableInfo)}

	return engine, nil

}

func (engine *Engine) NewSession() (*Session) {
	session := engine.createSession()

	return session
}

func (engine *Engine) Table(tableName string) *Session {
	session := engine.createSession()
	return session.Table(tableName)
}

func (engine *Engine) Prepare(sqlstr string) *Session {
	session := engine.createSession()
	return session.Prepare(sqlstr)
}


func (engine *Engine) createSession() *Session {
	return &Session{Engine: engine}
}

func (engine *Engine) Find(p interface{}) (bool, error) {

	session := engine.createSession()

	return session.Find(p)
}

func (engine *Engine) Select(p interface{}) error {
	session := engine.createSession()
	return session.Select(p)
}

func (engine *Engine) Where(wheres map[string]interface{}) *Session {

	session := engine.createSession()

	return session.Where(wheres)

}

func (engine *Engine) OrWhere(wheres map[string]interface{}) *Session {

	session := engine.createSession()

	return session.OrWhere(wheres)

}

func (engine *Engine) Limit(limit string) *Session {

	session := engine.createSession()

	return session.Limit(limit)
}

func (engine *Engine) Order(order string) *Session {

	session := engine.createSession()

	return session.Order(order)
}

func (engine *Engine) Group(group string) *Session {

	session := engine.createSession()

	return session.Group(group)
}

func (engine *Engine) Join(join string) *Session {

	session := engine.createSession()

	return session.Join(join)
}

func (engine *Engine) registerTable(t reflect.Type) error {

	engine.RWMuTables.Lock()
	defer engine.RWMuTables.Unlock()

	structName := t.Name()

	var tableName string
	if i := strings.Index(structName, "_JOIN"); i > 0 {
		tableName = structName[:i]
	} else {
		tableName = structName
	}

	fmt.Println("structName: ", structName)
	fmt.Println("tableName: ", tableName)

	tableInfo := TableInfo{
		Name: strings.ToLower(tableName),
		RWRuField: new(sync.RWMutex),
		Fields: make(map[string]FieldInfo),
	}

	for i := 0; i < t.NumField(); i ++ {

		attributeName := t.Field(i).Name
		fieldName := t.Field(i).Tag.Get("zyfield")
		asName := t.Field(i).Tag.Get("zyas")

		//获取 zytable tag 中的 表名, 如果没有, 就使用 tableName
		zytableName := t.Field(i).Tag.Get("zytable")
		if len(zytableName) < 1 {
			zytableName = strings.ToLower(tableName)
		}



		if fieldName == "-" {
			continue
		}

		if fieldName == "" {
			fieldName = strings.ToLower(attributeName)
		}

		//
		if asName == "" {
			asName = fieldName
		}


		tableInfo.RWRuField.Lock()

		tableInfo.Fields[asName] = FieldInfo{
								AttrName: attributeName,
								FieldName: fieldName,
								AsName: asName,
								TableName: zytableName,
							}

		tableInfo.RWRuField.Unlock()
	}

	engine.Tables[structName] = tableInfo

	return nil
}



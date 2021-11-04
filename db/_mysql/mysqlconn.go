package _mysql

import (
	"fmt"
	"github.com/fly-way/gofly/logs"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	opeQuery     = iota // 查询
	opeQueryMore = iota // 查询多条
	opeExec             // 插入
)

type mysqlWork struct {
	ope   int
	query string
	args  []interface{}
	data  interface{}
	done  *chan bool
}

type mysqlConn struct {
	obj       *sqlx.DB // 数据库对象
	works     chan mysqlWork
	closeChan chan bool
	idleConn  int
	openConn  int
}

func (_this *mysqlConn) SetConnPoolInfo(idleConn, openConn int) {
	_this.idleConn = idleConn
	_this.openConn = openConn
	if _this.obj != nil {
		_this.obj.SetMaxIdleConns(idleConn)
		_this.obj.SetMaxOpenConns(openConn)
	}

	logs.System("Mysql SetConnPoolInfo, idleConn=", _this.idleConn, "openConn=", _this.openConn)
}

func (_this *mysqlConn) StartConn(ip string, port int, user string, pwd string, dbName string, charset string, workSize int) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", user, pwd, ip, port, dbName, charset)
	obj, err := sqlx.Open("mysql", dsn)
	if err != nil {
		logs.Panic("Mysql StartConn open err:", err.Error())
	}

	if err := obj.Ping(); err != nil {
		logs.Panic("Mysql StartConn ping err:", err.Error())
	}

	_this.obj = obj
	_this.works = make(chan mysqlWork, workSize)
	_this.closeChan = make(chan bool, 1)
	if _this.idleConn != 0 && _this.openConn != 0 {
		_this.obj.SetMaxIdleConns(_this.idleConn)
		_this.obj.SetMaxOpenConns(_this.openConn)
	}

	go _this.startWork()

	logs.System("Mysql StartConn:", dsn, "idleConn=", _this.idleConn, "openConn=", _this.openConn, "workSize=", workSize)
}

func (_this *mysqlConn) CloseConn() {
	_this.closeChan <- true
	_this.obj.Close()
}

func (_this *mysqlConn) ExecChan(query string, args ...interface{}) {
	_this.works <- mysqlWork{
		ope:   opeExec,
		query: query,
		args:  args,
	}
}

func (_this *mysqlConn) QueryChan(query string, dest interface{}) chan bool {
	done := make(chan bool, 1)
	_this.works <- mysqlWork{
		ope:   opeQuery,
		query: query,
		data:  dest,
		done:  &done,
	}
	return done
}

func (_this *mysqlConn) QueryMoreChan(query string, dest interface{}) chan bool {
	done := make(chan bool, 1)
	_this.works <- mysqlWork{
		ope:   opeQueryMore,
		query: query,
		data:  dest,
		done:  &done,
	}
	return done
}

func (_this *mysqlConn) startWork() {
	var (
		result mysqlWork
		ok bool
	)

	for {
		select {
		case result, ok = <-_this.works:
		case <-_this.closeChan:
			return
		}

		if ok {
			var err error
			switch result.ope {
			case opeQuery:
				err = _this.obj.Get(result.data, result.query)
				close(*result.done)
			case opeQueryMore:
				err = _this.obj.Select(result.data, result.query)
				close(*result.done)
			case opeExec:
				_, err = _this.obj.Exec(result.query, result.args...)
			}

			if err != nil {
				logs.Error(fmt.Sprintf("Mysql ope failed, error:[%v], query:[%s]", err.Error(), result.query))
			}
		}
	}
}

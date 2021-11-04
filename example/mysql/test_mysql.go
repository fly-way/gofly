package main

import (
	"github.com/fly-way/gofly/db/_mysql"
	"github.com/fly-way/gofly/logs"
	"time"
)

func main() {
	// fly_game 数据库连接对象
	gameDB := _mysql.NewMysqlConn()
	gameDB.SetConnPoolInfo(4, 32)
	gameDB.StartConn("127.0.0.1", 3306, "root", "654321", "fly_game",
		"utf8mb4", 16)

	// 建表语句不建议在代码中执行, 这里是为了更直观的查看案例的表结构
	gameDB.ExecChan("DROP TABLE if EXISTS tb_player_info;")
	strSql := "CREATE TABLE `tb_player_info` (" +
		" `pid` bigint(20) NOT NULL DEFAULT 0 COMMENT '玩家ID'," +
		" `name` varchar(32) NOT NULL DEFAULT '' COMMENT '昵称'," +
		" `head` varchar(256) NOT NULL DEFAULT '' COMMENT '头像'," +
		" `sex` tinyint(1) NOT NULL DEFAULT 0 COMMENT '性别'," +
		" PRIMARY KEY (`pid`)" +
		") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT '玩家信息表';"
	gameDB.ExecChan(strSql)

	gameDB.ExecChan("INSERT INTO tb_player_info(`pid`,`name`) VALUES (?,?)", 10086, "移静员工")
	gameDB.ExecChan("INSERT INTO tb_player_info(`pid`,`name`,`sex`) VALUES (?,?,?)", 99999, "管理员", 1)

	var count int64
	<-gameDB.QueryChan("SELECT count(*) FROM tb_player_info", &count)
	logs.Logic(count)

	type SPlayerInfo struct {
		Pid  int64  `db:"pid"`
		Name string `db:"name"`
		Head string `db:"head"`
		Sex  int    `db:"sex"`
	}

	var player = SPlayerInfo{}
	<-gameDB.QueryChan("SELECT * FROM tb_player_info WHERE pid = 10086", &player)
	logs.Logic(player)

	var players = make([]SPlayerInfo, 0)
	<-gameDB.QueryMoreChan("SELECT * FROM tb_player_info", &players)
	logs.Logic(players)


	// 可以再建立一条数据库连接对象, 指向 fly_log (有些场景下可能需要同时连接多个数据库, 一般一个库对应一个 MysqlConn 即可)
	logDB := _mysql.NewMysqlConn()
	logDB.SetConnPoolInfo(2, 4)
	logDB.StartConn("127.0.0.1", 3306, "root", "654321", "fly_log",
		"utf8mb4", 16)

	// 也可以通过 logDB 对象操作 fly_log 数据库, 这里就不示范了
	// logDB.ExecChan(...)

	for {
		time.Sleep(1 * time.Second)
	}
}

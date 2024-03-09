package config

import (
	"net/url"
	"reflect"
	"strings"

	"github.com/go-ini/ini"
	lg "github.com/zwcway/castserver-go/common/log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func parseDatabase(cfg reflect.Value, k *ini.Key, ck *CfgKey) {
	var dbstr string
	if k == nil {
		dbstr = "sqlite:///file:memdb1?mode=memory&cache=shared"
		dbstr = "sqlite://castserver.db"
	} else {
		dbstr = k.String()
	}
	db := openDatabase(dbstr)
	if db == nil {
		return
	}
	cfg.Set(reflect.ValueOf(db))
}

func openDatabase(dbstr string) *gorm.DB {
	dbUrl, err := url.Parse(dbstr)
	if err != nil {
		log.Panic("database invalid", lg.String("url", dbstr), lg.Error(err))
		return nil
	}
	scheme := strings.ToLower(dbUrl.Scheme)
	dbUrl.Scheme = ""

	var dialector gorm.Dialector

	switch scheme {
	case "sqlite":
		dialector = sqlite.Open(dbUrl.String()[2:])
	// case "mysql":
	// 	dialector = mysql.Open(dbUrl.String()[2:])
	case "postgres", "pg":
		dialector = postgres.Open(postgresDSN(dbUrl, false))
	case "pgs":
		dialector = postgres.Open(postgresDSN(dbUrl, true))
	case "ms", "sqlserver", "sqls":
		dbUrl.Scheme = "sqlserver"
		dialector = sqlserver.Open(dbUrl.String())
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: lg.NewDBLog(lg.DebugLevel),
	})
	if err != nil {
		log.Panic("database connect failed", lg.String("url", dbstr), lg.Error(err))
		return nil
	}
	return db
}

func postgresDSN(Url *url.URL, ssl bool) (dsn string) {
	if ssl {
		dsn += "sslmode=enable "
	} else {
		dsn += "sslmodel=disable "
	}
	if len(Url.Host) > 0 {
		dsn += "host=" + Url.Host + " "
	}
	if len(Url.User.Username()) > 0 {
		dsn += "user=" + Url.User.Username() + " "
	}
	if pass, ok := Url.User.Password(); ok {
		dsn += "password=" + pass + " "
	}
	if len(Url.Port()) > 0 {
		dsn += "port=" + Url.Port() + " "
	}
	if len(Url.Path) > 0 {
		dsn += "dbname=" + Url.Host + " "
	}
	for k, v := range Url.Query() {
		if len(v) > 0 {
			dsn += k + "=" + v[0] + " "
		}
	}

	dsn = dsn[:len(dsn)-1]
	return
}

func getDSN(db *gorm.DB) string {
	d := db.Dialector
	if sd, ok := d.(*sqlite.Dialector); ok {
		return "sqlite://" + sd.DSN
		// } else if sd, ok := d.(*mysql.Dialector); ok {
		// 	return "mysql://" + sd.Config.DSN
	} else if sd, ok := d.(*postgres.Dialector); ok {
		return "postgres://" + sd.Config.DSN
	} else if sd, ok := d.(*sqlserver.Dialector); ok {
		return "sqlserver://" + sd.Config.DSN
	}
	return ""
}

func Deinit() {
	d, err := DB.DB()
	if err == nil {
		d.Close()
	}
}

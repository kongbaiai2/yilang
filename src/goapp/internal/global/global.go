package global

import (
	"database/sql"
	"runtime"

	"github.com/kongbaiai2/yilang/goapp/internal/callothers/cacti_proxy"
	"github.com/kongbaiai2/yilang/goapp/internal/config"

	"github.com/gomodule/redigo/redis"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	CONFIG config.Server

	DB    *gorm.DB
	SqlDB *sql.DB

	LOG   *zap.SugaredLogger
	DBLOG *zap.SugaredLogger
	//VIPER  *viper.Viper
	REDIS *redis.Pool

	// ProcessExit 全局变量 进程是否退出
	ProcessExit = false
	Cacti       *cacti_proxy.CactiOptions
)

func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	LOG.Infof("==> %s", string(buf[:n]))
}

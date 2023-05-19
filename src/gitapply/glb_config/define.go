package glb_config

var (
	Cfg config
)

const (
	// "select h.itemid,h.time,h.value,i.name from (select itemid,from_unixtime(clock) as time,value from history_uint where itemid IN (select itemid  from items where hostid=10439 and (name like '%/7%' or name like '%/8%' or name like '%/9%' or name like '%/10%' or name like '%/11%' or name like '%/12%') and key_ like '%ifHC%' ) and clock >= unix_timestamp('2023/04/01 00:00:00') and clock <= unix_timestamp('2023/05/1 00:00:00')) as h INNER join items i where h.itemid=i.itemid limit 5"
	SelectItemidForHostInt    = `select itemid from items where hostid=(select hostid from hosts where host='%s') and name like '%s' and key_ like '%s'`
	SelectValueForHistoryUnit = `select clock,value from history_uint where itemid=?`
	ConditionTimeToTime       = ` and clock >= unix_timestamp(?) and clock <= unix_timestamp(?)`
)

type config struct {
	Dbstring string `mapstructure:"dbstring"`
}

type Respond struct {
	Host        string
	IntName     string
	InOrOut     string
	StartTime   string
	EndTime     string
	Itemids     DataForHistory
	TurnItemids DataForHistory
}

type DataForHistory struct {
	Itemid string
	Datas  []DataMapForHistory
}
type DataMapForHistory struct {
	Clock int64
	Value int64
}

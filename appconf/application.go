package appconf

// Version 程序版本号
const Version string = "V1.0.1"

// Application 应用程序配置对象
// 该对象用于持有配置文件出现的所有配置参数
type Application struct {
	Database        Database `yaml:"database"`        // 数据库连接配置，不同的数据库驱动连接配置不一样，见数据库驱动
	Port            int      `yaml:"port"`            // 端口
	LogKeepMaxDays  int      `yaml:"logKeepMaxDays"`  // 操作日志最大保存天数，注意若该值小于等于0则表示不删除。
	NoteKeepMaxDays int      `yaml:"noteKeepMaxDays"` // 操作日志最大保存天数，注意若该值小于等于0则表示不删除
	SSOBaseUrl      string   `yaml:"SSOBaseUrl"`      // 单点登录基础路径
	Debug           bool     `yaml:"debug"`           // 调试模式
}

// Database 数据库配置
type Database struct {
	Type string // 数据库类型：mysql、dm
	DSN  string // 连接地址
}

// 无法找到配置文件时候的缺省配置
var defaultConfig = Application{
	Database: Database{
		DSN:  "root:123qwe@tcp(127.0.0.1:3306)/note?charset=utf8mb4&parseTime=true",
		Type: "mysql",
	},
	LogKeepMaxDays:  3 * 30, // 3月
	NoteKeepMaxDays: 30,     // 1月
	Port:            8011,
	SSOBaseUrl:      "http://nantemen.hzauth.com",
	Debug:           true,
}

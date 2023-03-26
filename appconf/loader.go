package appconf

import (
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

// Load 从配置文件中加在配置信息
// 默认从可执行程序的相对目录中获取名为 application.yml 文件
func Load() *Application {
	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	// 兼容两种格式的Yaml命名风格
	p := filepath.Join(base, "application.yml")
	bin, _ := os.ReadFile(p)
	if len(bin) == 0 {
		p = filepath.Join(base, "application.yaml")
		bin, _ = os.ReadFile(p)
	}
	if len(bin) == 0 {
		// 没有加载到配置文件的情况使用默认配置
		res := &defaultConfig
		// 生成默认配置文件
		b, _ := yaml.Marshal(res)
		_ = os.WriteFile(p, b, os.FileMode(0666))
		return res
	}
	// 复制一份默认配置，防止出现空文件的导致的配置参数为空
	var res = Application{}
	err := yaml.Unmarshal(bin, &res)
	if res.Port <= 0 {
		res.Port = 80
	}

	if err != nil {
		// 没有加载到配置文件的情况使用默认配置
		return &defaultConfig
	}
	return &res
}

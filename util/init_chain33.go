package util

import (
	"bytes"
	"fmt"

	_ "github.com/33cn/chain33/system" // chain system init
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin" // chain plugin init
	tml "github.com/BurntSushi/toml"
)

// types/config L100 初始化， 需要手动调用， 不能自动
// 1. chain33 更新配置的实现后， 插件需要手动调用生效
// 2. 用于不同的链，插件不同加载的配置不同
// 3. 提供一个全的配置， 用对应链的配置来覆盖默认的配置， 再用这个配置来初始化
// 4. meger config 不可用， 只能用于合并 chain33配置和plugin的配置(合并的项不能重复)， 和这里覆盖不一样
// 所以要在这里实现一个

// InitChain33 初始化chain33
func InitChain33(title, symbol, configFile string) {
	config := Chain33DefaultConfig
	configString := configV165(title, symbol, Chain33V165)
	if configFile != "" {
		configString = types.ReadFile(configFile)
	}
	config = mergeConfig(configString, config)
	types.NewChain33Config(config)
}

func configV165(title, symbol, defaultConfig string) string {
	titleStr := fmt.Sprintf(Chain33TitleFmt, title)
	symbolStr := fmt.Sprintf(Chain33CoinSymbol, symbol)
	configString := titleStr + symbolStr + defaultConfig
	return configString
}

// 覆盖链上的配置
func overwriteConfig(conf map[string]interface{}, def map[string]interface{}) {
	for key1, value1 := range conf {
		//log.Debug("conf", "key", key1)
		if conf1, ok := value1.(map[string]interface{}); ok {
			if vdef, ok := def[key1]; ok {
				def1, _ := vdef.(map[string]interface{})
				//log.Debug("conf->", "key", key1)
				overwriteConfig(conf1, def1)
				//log.Debug("conf<-", "key", key1)
				def[key1] = def1
			} else {
				def[key1] = conf1
			}
		} else {
			//log.Debug("set2", "key", key1, "from", def[key1], "to", value1)
			def[key1] = value1
		}
	}
	//log.Error("config", "def", def)
}

// mergeConfig
func mergeConfig(userConfig, defaultConfig string) string {
	//1. defconfig
	def := make(map[string]interface{})
	_, err := tml.Decode(defaultConfig, &def)
	if err != nil {
		panic(err)
	}
	//2. userconfig
	conf := make(map[string]interface{})
	_, err = tml.Decode(userConfig, &conf)
	if err != nil {
		panic(err)
	}

	overwriteConfig(conf, def)
	buf := new(bytes.Buffer)
	tml.NewEncoder(buf).Encode(def)
	return buf.String()
}

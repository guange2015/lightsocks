package cmd

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	ListenAddr string `json:"listen"`
	RemoteAddr string `json:"remote"`
	Password   string `json:"password"`
}

// 保存配置到配置文件
func SaveConfig(configPath string, config *Config) error {
	configJson, _ := json.MarshalIndent(config, "", "	")
	err := ioutil.WriteFile(configPath, configJson, 0644)
	if err != nil {
		log.Printf("保存配置到文件 %s 出错: %s\n", configPath, err)
		return err
	}
	log.Printf("保存配置到文件 %s 成功\n", configPath)
	return nil
}

func ReadConfig(configPath string) (error, *Config) {
	config := &Config{}

	// 如果配置文件存在，就读取配置文件中的配置 assign 到 config
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		log.Printf("从文件 %s 中读取配置\n", configPath)
		file, err := os.Open(configPath)
		if err != nil {
			log.Printf("打开配置文件 %s 出错:%s", configPath, err)
			return err, nil
		}
		defer file.Close()

		err = json.NewDecoder(file).Decode(config)
		if err != nil {
			log.Printf("格式不合法的 JSON 配置文件:\n%s", file)
			return err, nil
		}
	} else {
		log.Printf("配置文件不存在:%s\n", configPath)
		return errors.New("配置文件不存在"), nil
	}

	return nil, config
}

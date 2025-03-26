package viperx

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/viper"
)

var paths = []string{
	".",
	"./etc",
	"./config",
}

// ConfigFileEncryptorHook 配置文件加密器
type ConfigFileEncryptorHook interface {
	IsCipherText([]byte) bool
	Encrypt([]byte) []byte
	Decrypt([]byte) []byte
}

var configFileEH ConfigFileEncryptorHook

// RegisterConfigEncryptor 注册配置文件加密器
func RegisterConfigEncryptor(h ConfigFileEncryptorHook) {
	configFileEH = h
}

// GetViper 获取viper配置
// name: 配置文件名,不带后缀
// filetype: 配置文件类型，如json、yaml、ini等
func GetViper(name, filetype string) *viper.Viper {
	buf, err := ReadFile(name, filetype)
	if err != nil {
		panic(fmt.Sprintf("Error while reading config file %s: %v", name+filetype, err))
	}

	if configFileEH != nil {
		if configFileEH.IsCipherText(buf) {
			buf = configFileEH.Decrypt(buf)
		}
	}

	vp := viper.New()
	vp.SetConfigName(name)
	vp.SetConfigType(filetype)
	if err = vp.ReadConfig(bytes.NewReader(buf)); err != nil {
		panic(fmt.Sprintf("Error while reading config file %s: %v", name+filetype, err))
	}
	return vp
}

func ReadFile(name, filetype string) ([]byte, error) {
	for _, v := range paths {
		file := fmt.Sprintf("%s/%s.%s", v, name, filetype)
		if _, err := os.Stat(file); err == nil {
			return os.ReadFile(file)
		}
	}
	return nil, nil
}

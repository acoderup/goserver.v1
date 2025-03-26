package core

import (
	"io"
	"strings"

	"github.com/acoderup/goserver/core/logger"
	"github.com/acoderup/goserver/core/viperx"
)

var packages = make(map[string]Package)
var packagesLoaded = make(map[string]bool)

// Package 功能包
// 只做初始化，不要依赖其它功能包
type Package interface {
	Name() string
	Init() error
	io.Closer
}

// RegistePackage 注册功能包
func RegistePackage(p Package) {
	packages[p.Name()] = p
}

// IsPackageRegistered 判断功能包是否已经注册
func IsPackageRegistered(name string) bool {
	if _, exist := packages[name]; exist {
		return true
	}
	return false
}

// IsPackageLoaded 判断功能包是否已经加载
func IsPackageLoaded(name string) bool {
	if _, exist := packagesLoaded[name]; exist {
		return true
	}
	return false
}

// RegisterConfigEncryptor 注册配置文件加密器
func RegisterConfigEncryptor(h viperx.ConfigFileEncryptorHook) {
	viperx.RegisterConfigEncryptor(h)
}

// LoadPackages 加载功能包
func LoadPackages(configFile string) {
	val := strings.Split(configFile, ".")
	if len(val) != 2 {
		panic("config file name error")
	}

	vp := viperx.GetViper(val[0], val[1])

	var err error
	var notFoundConfig []string
	var notFoundPackage []string
	for k := range vp.AllSettings() {
		if _, ok := packages[k]; !ok {
			notFoundPackage = append(notFoundPackage, k)
			continue
		}

		name := k
		pkg := packages[k]
		if err = vp.UnmarshalKey(k, pkg); err != nil {
			logger.Logger.Errorf("Package %s: Error while unmarshalling from config file %s: %v", name, configFile, err)
			continue
		}

		if err = pkg.Init(); err != nil {
			logger.Logger.Errorf("Package %s: Error while initializing from config file %s: %v", name, configFile, err)
			continue
		}

		packagesLoaded[pkg.Name()] = true
		logger.Logger.Infof("package [%16s] load success", pkg.Name())
	}

	for k := range packages {
		if !IsPackageLoaded(k) {
			notFoundConfig = append(notFoundConfig, k)
		}
	}

	if len(notFoundConfig) > 0 {
		logger.Logger.Warnf("package load success, not found config: %v", notFoundConfig)
	}

	if len(notFoundPackage) > 0 {
		logger.Logger.Warnf("package load success, not found package: %v", notFoundPackage)
	}
}

// ClosePackages 关闭功能包
func ClosePackages() {
	for _, pkg := range packages {
		err := pkg.Close()
		if err != nil {
			logger.Logger.Errorf("Error while closing package %s: %s", pkg.Name(), err)
		}
	}
}

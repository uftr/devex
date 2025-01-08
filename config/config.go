package config

import (
	"reflect"
	"strconv"
	"strings"
)

const (
	WORKSPACE_KEY = "environment"
	WORKSPACE_DEF = "default"
)

type Config struct {
	Modfile   string `default:"main.tf"`
	ConfPath  string `default:"src"`
	ConfFile  string `default:"config.txt"`
	CachePath string `default:".cache"`
	LogFile   string `default:"log.txt"`
	Tabsize   int    `default:"4"`
}

// Returns new Config object
func NewConfig() Config {
	p := Config{}
	setDefaults(&p)
	return p
}

func (cfg *Config) GetConfFile(myenv string) string {
	if myenv == WORKSPACE_DEF {
		return cfg.ConfFile
	}
	return myenv + "-" + cfg.ConfFile
}

func GetEnvFromConfFile(myconfFile string) string {
	idx := strings.Index(myconfFile, "-")
	if idx > 0 && idx < len(myconfFile) {
		return myconfFile[idx+1:]
	}
	return WORKSPACE_DEF
}

// Sets the filepath
func (cfg *Config) SetFilePath(p string) {
	cfg.Modfile = p
}

// Sets the conf filepath
func (cfg *Config) SetConfPath(p string) {
	cfg.ConfPath = p
}

// Sets the tab size
func (cfg *Config) SetTabSize(s int) {
	cfg.Tabsize = s
}

// Sets Default values of the config object
func setDefaults(p *Config) {
	// Iterate over the fields of the Person struct using reflection
	// and set the default value for each field if the field is not provided
	// by the caller of the constructor function.
	for i := 0; i < reflect.TypeOf(*p).NumField(); i++ {
		field := reflect.TypeOf(*p).Field(i)
		if value, ok := field.Tag.Lookup("default"); ok {
			switch field.Type.Kind() {
			case reflect.String:
				if field.Name == "Modfile" && p.Modfile == "" {
					p.Modfile = value
				}
				if field.Name == "ConfPath" && p.ConfPath == "" {
					p.ConfPath = value
				}
				if field.Name == "ConfFile" && p.ConfFile == "" {
					p.ConfFile = value
				}
				if field.Name == "CachePath" && p.CachePath == "" {
					p.CachePath = value
				}
				if field.Name == "LogFile" && p.LogFile == "" {
					p.LogFile = value
				}
			case reflect.Int:
				if p.Tabsize == 0 {
					if intValue, err := strconv.Atoi(value); err == nil {
						p.Tabsize = intValue
					}
				}
			}
		}
	}
}

package mailwatcher

import (
	"errors"
	"log"
	"os"
	"path"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Extractor struct {
	Reg     regexp.Regexp
	Capture interface{}
}

type Configuration struct {
	DatabasePath string
	Subjects     []string
	Extractors   []Extractor
}

func DefaultConfigFile() string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln("Failed to get current working directory")
	}

	return path.Join(cwd, "config.yaml")
}

func LoadConfig(configFile string) (Configuration, error) {
	conf := Configuration{}
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return conf, err
	}

	bytes, err := os.ReadFile(configFile)
	if err != nil {
		return conf, err
	}

	var config = struct {
		Db   string   `yaml:"database,omitempty"`
		Subs []string `yaml:"subjects"`
		Extr []struct {
			Reg string      `yaml:"regex"`
			Cap interface{} `yaml:"capture"`
		} `yaml:"extractors"`
	}{}
	err = yaml.Unmarshal(bytes, &config)

	if err != nil {
		return conf, err
	}

	regs := []Extractor{}
	for _, reg := range config.Extr {
		pReg, err := regexp.Compile(reg.Reg)
		if err != nil {
			log.Fatalf("Invalid extraction regex: %s\n", reg.Reg)
		}
		_, isString := reg.Cap.(string)
		_, isInt := reg.Cap.(int)
		if !isString && !isInt {
			log.Fatalln("Extraction capture must be either an index (positive int) or a name (string)")
		}

		regs = append(regs, Extractor{
			Reg:     *pReg,
			Capture: reg.Cap,
		})
	}
	conf.Extractors = regs
	conf.Subjects = config.Subs
	conf.DatabasePath = config.Db

	return conf, nil
}

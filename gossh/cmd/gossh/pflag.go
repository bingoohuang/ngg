package main

import (
	"bytes"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/bingoohuang/ngg/gossh/pkg/cnf"
	"github.com/bingoohuang/ngg/ss"
	"github.com/spf13/viper"
	"log"
	"os"
)

// LoadByPflag load values to cfgValue from pflag cnf specified path.
func LoadByPflag(cfgValues ...any) {
	f := ss.ExpandHome(viper.GetString("cnf"))
	Load(f, cfgValues...)
}

// Load loads the cnfFile content and viper bindings to value.
func Load(cnfFile string, values ...any) {
	if cnfFile != "" {
		if err := LoadE(cnfFile, values...); err != nil {
			log.Printf("E! Load Cnf %s error %v", cnfFile, err)
		}
	}
	cnf.ViperToStruct(values...)
}

// LoadE similar to Load.
func LoadE(cnfFile string, values ...any) error {
	f, err := cnf.FindFile(cnfFile)
	if err != nil {
		return fmt.Errorf("FindFile error %w", err)
	}

	bs, err := os.ReadFile(f)
	if err != nil {
		return err
	}

	for _, value := range values {
		_, err := toml.NewDecoder(bytes.NewReader(bs)).Decode(value)
		if err != nil {
			return fmt.Errorf("DecodeFile error %w", err)
		}
	}

	return nil
}

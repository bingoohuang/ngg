package conf

import (
	_ "embed"
	"log"
	"os"

	"github.com/bingoohuang/ngg/ss"
)

func init() {
	registerOptions(`%demo.env`, `%demo.env;`,
		func(name string, options *replOptions) {
			if ok, _ := ss.Exists(".env"); ok {
				log.Printf(".env file already exists, please remove/rename it first")
				return
			}
			if err := os.WriteFile(".env", demoEnv, 0o644); err != nil {
				log.Printf("write .env file failed: %v", err)
				return
			}

			log.Printf("demo .env file created!")
		}, func(name string, options *replOptions, args []string, pureArg string) error {
			return nil
		})
}

//go:embed demo.env
var demoEnv []byte

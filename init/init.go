package init

import (
	"log"
	"os"
	cfg "vdex/config"
	parcer "vdex/parser"
)

/*
 * Prompts the user for the configuration data and saves it in the target location
 * Returns
 * string: file location where the config is saved
 * error: if any failure
 */
func VdexInit(config *cfg.Config) (string, error) {

	log.Printf("\nIn VdexInit")
	file, err := os.Open(config.Modfile)
	if err != nil {
		log.Println("Failed to open file:", config.Modfile)
		return "", err
	}
	defer file.Close()

	//log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	parcedBlocks, err := parcer.ParseTF(config.Modfile, nil, nil)
	if err != nil {
		log.Println("Failed to parse file:", config.Modfile)
		return "", err
	}

	//tfbs.Walk(0, config.Tabsize, outlog)
	return PromptConfig(parcedBlocks, config.ConfPath, config.ConfFile)

}

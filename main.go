package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	cfg "vdex/config"
	vinit "vdex/init"
	vplan "vdex/plan"
)

func main() {

	print_help := false
	total_args := len(os.Args[1:])

	if total_args != 1 {
		print_help = true
	} else if os.Args[1] == "-help" || os.Args[1] == "--help" || os.Args[1] == "?" || os.Args[1] == "help" {
		print_help = true
	}

	if print_help {
		pgname := path.Base(os.Args[0])
		idx := strings.LastIndex(pgname, string(os.PathSeparator))
		if idx >= 0 {
			if idx < len(pgname) {
				idx++
				pgname = pgname[idx:]
			}
		}
		fmt.Println("Usage:")
		fmt.Println(pgname, "init | plan | apply")
		fmt.Println("    init    - takes user input for REPLACE-ME values and stores the config in sys/<SYSTEM-NAME>/")
		fmt.Println("              <SYSTEM-NAME> is one of the user input")
		fmt.Println("    plan    - generates the main.tf file with the user values in sys/<SYSTEM-NAME>/")
		fmt.Println("    apply   - executes terraform plan on the generated file")
		fmt.Println("    help    - this usage text")
		return
	}

	config := cfg.NewConfig()

	// open log file
	logFileLocation := filepath.Join(config.ConfPath, config.LogFile)
	logFile, err := os.OpenFile(logFileLocation, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()

	// Set log out put and enjoy :)
	log.SetOutput(logFile)

	user_cmd := os.Args[1]

	switch user_cmd {
	case "init":

		file, err := os.Open(config.Modfile)
		if err != nil {
			log.Println("Failed to open file:", config.Modfile)
			log.Println("Error:", err)
			fmt.Println("Failed to open file:", config.Modfile)
			return
		}
		defer file.Close()

		var saveConfFile string
		saveConfFile, err = vinit.VdexInit(&config)
		if err != nil {
			fmt.Printf("\ninit failed, see logs %s\n", logFileLocation)
		} else {
			fmt.Printf("\ninit Success - config is saved in %s\n", saveConfFile)
		}
	case "plan":
		fileList, err := vplan.VdexPlanGen(&config)
		if err != nil {
			fmt.Printf("\nplan generation failed, see logs %s\n", logFileLocation)
		} else {
			fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
			vplan.VdexTerraformExecute(&config, fileList, "plan")
		}
	case "apply":
		fileList, err := vplan.VdexPlanGen(&config)
		if err != nil {
			fmt.Printf("\nplan generation failed, see logs %s\n", logFileLocation)
		} else {
			fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
			vplan.VdexTerraformExecute(&config, fileList, "apply")
		}
	default:
	}
}

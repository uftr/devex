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

func printHelp(pgname string) {
	idx := strings.LastIndex(pgname, string(os.PathSeparator))
	if idx >= 0 {
		if idx < len(pgname) {
			idx++
			pgname = pgname[idx:]
		}
	}
	fmt.Println("Usage:")
	fmt.Println(pgname, "init | plan [-s] | apply [-s]")
	fmt.Println("    init       - takes user input for REPLACE-ME values and stores the config in sys/<SYSTEM-NAME>/")
	fmt.Println("                 <SYSTEM-NAME> is one of the user input")
	fmt.Println("    plan  [-s] - generates the main.tf file with the user values in sys/<SYSTEM-NAME>/.cache and executes terraform init & plan")
	fmt.Println("                 if -s option is passed, terraform init will be skipped")
	fmt.Println("    apply [-s] - similar plan but terraform apply is executed instead of terraform plan")
	fmt.Println("                 if -s option is passed, terraform init will be skipped")
	fmt.Println("    help       - this usage text")
}

func main() {

	print_help := false
	apply_tf_init := true
	cmd_start_idx := 1
	user_cmd := ""
	//tf_env := "default"
	total_args := len(os.Args[1:])

	pgname := path.Base(os.Args[0])

	if total_args == 0 || total_args > 3 {
		print_help = true
	} else if os.Args[1] == "-help" || os.Args[1] == "--help" || os.Args[1] == "?" || os.Args[1] == "help" {
		print_help = true
	}

	if total_args == 2 || total_args == 3 {
		if strings.TrimSpace(os.Args[2]) == "-s" {
			apply_tf_init = false
			cmd_start_idx = 1
		} else if strings.TrimSpace(os.Args[1]) == "-s" {
			apply_tf_init = false
			cmd_start_idx = 2
		} else {
			print_help = true
		}
	}

	if print_help {
		printHelp(pgname)
		return
	}

	user_cmd = os.Args[cmd_start_idx]

	config := cfg.NewConfig()

	// open log file
	if _, err := os.Stat(config.ConfPath); os.IsNotExist(err) { // Create Path if not present
		err = os.Mkdir(config.ConfPath, 0755) //create a directory
		if err != nil {
			log.Println("Failed to create directory", config.ConfPath) //print the error on the console
			return
		}
	}
	logFileLocation := filepath.Join(config.ConfPath, config.LogFile)
	logFile, err := os.OpenFile(logFileLocation, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Println(err)
	}
	defer logFile.Close()

	// Set log out put and enjoy :)
	log.SetOutput(logFile)

	switch user_cmd {
	case "init":

		file, err := os.Open(config.Modfile)
		if err != nil {
			log.Println("Failed to open file:", config.Modfile)
			log.Println("Error:", err)
			fmt.Println("Failed to access the terraform file:", config.Modfile)
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
			if len(fileList) > 0 {
				fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
				vplan.VdexTerraformExecute(&config, fileList, "plan", apply_tf_init)
			} else {
				fmt.Printf("\nplan generation skipped - no config file is found, try init \n")
			}
		}
	case "apply":
		fileList, err := vplan.VdexPlanGen(&config)
		if err != nil {
			fmt.Printf("\nplan generation failed, see logs %s\n", logFileLocation)
		} else {
			if len(fileList) > 0 {
				fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
				vplan.VdexTerraformExecute(&config, fileList, "apply", apply_tf_init)
			} else {
				fmt.Printf("\nplan generation skipped - no config file is found, try init \n")
			}
		}
	default:
		printHelp(pgname)
		return
	}
}

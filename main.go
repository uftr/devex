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
	vlist "vdex/list"
	vplan "vdex/plan"
)

// Prints the Help text
func printHelp(pgname string) {
	idx := strings.LastIndex(pgname, string(os.PathSeparator))
	if idx >= 0 {
		if idx < len(pgname) {
			idx++
			pgname = pgname[idx:]
		}
	}
	fmt.Println("Usage:")
	fmt.Println(pgname, "init | plan [-s] | apply [-s] | list")
	fmt.Println("    init [envName] - Takes user input for REPLACE-ME values found in main.tf and stores the config in")
	fmt.Println("                     sys/<SYSTEM-NAME>/, <SYSTEM-NAME> is one of the user input")
	fmt.Println("                   - envName is optional argument and if passed, it is treated as the environment which creates")
	fmt.Println("                     a distrinct config file for the environment. It generated the file <envName>-config.txt")
	fmt.Println("")
	fmt.Println("    plan [-s] [envName] - Generates the main.tf (in sys/<SYSTEM-NAME>/.cache) by replacing the")
	fmt.Println("                     variable values with the user provided values and executes terraform init & plan")
	fmt.Println("                     If -s option is passed, terraform init is skipped but plan is executed")
	fmt.Println("                   - envName is optional argument and if passed, it is treated as the target environment causing it to")
	fmt.Println("                     process the config file named <envName>-config.txt. The Workspace named <envName> gets created")
	fmt.Println("                     during terraform init and terraform plan.")
	fmt.Println("")
	fmt.Println("    apply [-s] [envName]- similar plan but terraform apply is executed instead of terraform plan")
	fmt.Println("                     otherwise, rest of the behaviour is same as plan.")
	fmt.Println("")
	fmt.Println("    list [envName] - Lists out the user configured system-names and the environments")
	fmt.Println("                   - envName is optional argument and if passed, filter gets applied on the environments")
	fmt.Println("")
	fmt.Println("    help           - this usage text")
}

func main() {

	// default values
	print_help := false
	apply_tf_init := true
	cmd_start_idx := 1
	user_cmd := ""
	user_env := cfg.WORKSPACE_DEF
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
			if total_args > 2 {
				user_env = os.Args[3]
			}
		} else if strings.TrimSpace(os.Args[1]) == "-s" {
			apply_tf_init = false
			cmd_start_idx = 2
			if total_args > 2 {
				user_env = os.Args[3]
			}
		} else {
			cmd_start_idx = 1
			user_env = os.Args[2]
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
	case "init": // handle init command

		file, err := os.Open(config.Modfile)
		if err != nil {
			log.Println("Failed to open file:", config.Modfile)
			log.Println("Error:", err)
			fmt.Println("Failed to access the terraform file:", config.Modfile)
			return
		}
		defer file.Close()

		var saveConfFile string
		saveConfFile, err = vinit.VdexInit(&config, user_env)
		if err != nil {
			fmt.Printf("\ninit failed, see logs %s\n", logFileLocation)
		} else {
			fmt.Printf("\ninit Success - config is saved in %s\n", saveConfFile)
		}
	case "plan": // handle plan command
		fileList, err := vplan.VdexPlanGen(&config, user_env)
		if err != nil {
			fmt.Printf("\nplan generation failed, see logs %s\n", logFileLocation)
		} else {
			if len(fileList) > 0 {
				fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
				vplan.VdexTerraformExecute(&config, fileList, "plan", apply_tf_init, user_env)
			} else {
				fmt.Printf("\nplan generation skipped - no config file is found, try init \n")
			}
		}
	case "apply": // handle apply command
		fileList, err := vplan.VdexPlanGen(&config, user_env)
		if err != nil {
			fmt.Printf("\nplan generation failed, see logs %s\n", logFileLocation)
		} else {
			if len(fileList) > 0 {
				fmt.Printf("\nplan generation Success - generated files %v\n", fileList)
				vplan.VdexTerraformExecute(&config, fileList, "apply", apply_tf_init, user_env)
			} else {
				fmt.Printf("\nplan generation skipped - no config file is found, try init \n")
			}
		}
	case "list":
		vlist.ListSystems(&config, user_env)
	default:
		printHelp(pgname)
		return
	}
}

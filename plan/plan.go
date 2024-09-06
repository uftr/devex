package plan

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	cfg "vdex/config"
	parcer "vdex/parser"
)

func ReadConfigFile(config *cfg.Config, teamCfgPath string, teamCfgFile string) (string, error) {
	var parcedBlocks parcer.TFBlocks
	parcedBlocks.Init()

	log.Printf("\nIn ReadConfigFile %s", teamCfgFile)

	// Read the user configuration file into userConfig
	var userConfig map[string]string = make(map[string]string)

	file, err := os.Open(teamCfgFile)
	if err != nil {
		log.Println("Failed to open file:", teamCfgFile)
		return "", err
	}
	defer file.Close()

	cfgScanner := bufio.NewScanner(file)

	for cfgScanner.Scan() {
		text := strings.TrimSpace(cfgScanner.Text())
		if text == "" || strings.HasPrefix(text, parcer.COMMENT1) || strings.HasPrefix(text, parcer.COMMENT2) || strings.HasPrefix(text, parcer.COMMENT3) {
			continue
		}
		idx := strings.Index(text, "=")
		if idx >= 0 {
			k := strings.TrimSpace(text[:idx])
			v := ""
			if idx+1 < len(text) {
				v = strings.TrimSpace(text[idx+1:])
			}
			userConfig[k] = v
		}
	}

	// Set the userConfig to parced params object
	for k, v := range userConfig {
		var newParam parcer.ParamValue
		newParam.P_value = v
		parcedBlocks.Param[k] = newParam
		log.Println(k, "=>", v)
	}

	// create the main.tf
	mainPath := path.Join(teamCfgPath, config.CachePath)
	if _, err := os.Stat(mainPath); os.IsNotExist(err) { // Create Path if not present
		err = os.Mkdir(mainPath, 0755) //create a directory
		if err != nil {
			log.Println("Failed to create directory", mainPath) //print the error on the console
			return "", err
		}
	}

	mainFile := path.Join(mainPath, config.Modfile)
	oFile, err := os.OpenFile(mainFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Failed to open file:", mainFile)
		return mainFile, err
	}
	defer oFile.Close()

	// Parce the template main.tf
	parcedBlocks.Skip = true
	_, err = parcer.ParseTF(config.Modfile, &parcedBlocks, oFile)
	if err != nil {
		log.Println("Failed to parse file:", config.Modfile)
		return "", err
	}

	//parcedBlocks.Walk(0, config.Tabsize, oFile)

	return mainFile, nil
}

func ProcessConfigFiles(config *cfg.Config) ([]string, error) {
	var fileList []string
	confPath := config.ConfPath
	entries, err := os.ReadDir(confPath)

	if err != nil {
		log.Println(err)
		return fileList, err
	}

	for _, v := range entries {
		if !v.IsDir() {
			continue
		}
		//fmt.Println(v.Name())

		teamCfgFile := path.Join(confPath, v.Name(), config.ConfFile)
		if _, err := os.Stat(teamCfgFile); err == nil {
			log.Printf("File %s exists\n", teamCfgFile)
			teamCfgPath := path.Join(confPath, v.Name())
			genfile, err := ReadConfigFile(config, teamCfgPath, teamCfgFile)
			if err == nil {
				fileList = append(fileList, genfile)
			}
		}
	}

	return fileList, nil
}

/*
 * Runs terraform plan on the generated files
 * Returns
 * error: if any failure
 */
func VdexTerraformExecute(config *cfg.Config, fileList []string, tfparam string, tfinit bool) error {
	log.Printf("\nIn VdexPlanExecute")
	app := "terraform"

	if runtime.GOOS == "windows" {
		app = "terraform.exe"
	}
	_, err := exec.LookPath(app)
	if err != nil {
		log.Printf("\n terraform binary not found %s", app)
		fmt.Printf("\n terraform binary not found %s", app)
	}

	curPath, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get the current directory\n %s", err)
	} else {
		log.Printf("current working path %s", curPath)
	}

	for _, tfFile := range fileList {

		// get the service-team path and cd to it
		idx := strings.LastIndex(tfFile, config.Modfile)

		if idx < 0 {
			continue
		}
		tfPath := "."
		if idx > 0 {
			tfPath = tfFile[:idx]
		}
		err = os.Chdir(tfPath)
		if err != nil {
			log.Printf("\nFailed to cd the directory to service-team %s", tfPath)
			continue
		}
		log.Printf("\ncd the directory to service-team %s", tfPath)

		if tfinit {
			// execute terraform init command
			fmt.Println("terraform init...")
			cmdoutput, err := exec.Command(app, "init").Output()

			if err != nil {
				log.Println(err.Error())
				log.Println("Failed to execute terraform", "init", "in", tfPath)
				fmt.Println(err.Error())
				fmt.Println("Failed to execute terraform", "init", "in", tfPath)
			} else {
				log.Println(string(cmdoutput))
				log.Println("Successfully executed terraform", "init", "in", tfPath)
				//fmt.Println(string(cmdoutput))
				fmt.Println("Successfully executed terraform", "init", "in", tfPath)
			}
		}

		// execute terraform plan or apply command
		cmdoutput, err := exec.Command(app, tfparam).Output()

		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("Failed to execute terraform", tfparam, "in", tfPath)
			fmt.Println("Please verify validity of the terraform")
			log.Println(err.Error())
			log.Println("Failed to execute terraform", tfparam, "in", tfPath)
		} else {
			//fmt.Println(string(cmdoutput))
			fmt.Println("Successfully executed terraform", tfparam, "in", tfPath)
			log.Println(string(cmdoutput))
			log.Println("Successfully executed terraform", tfparam, "in", tfPath)
		}

		// Return to the working directory
		err = os.Chdir(curPath)
		if err != nil {
			log.Printf("\nFailed to return to the working path %s", curPath)
		}
	}
	return nil
}

/*
 * Reads the config files and generates main.tf
 * Returns
 * list of generated files
 * error: if any failure
 */
func VdexPlanGen(config *cfg.Config) ([]string, error) {
	log.Printf("\nIn VdexPlan")
	fileList, err := ProcessConfigFiles(config)
	return fileList, err
}

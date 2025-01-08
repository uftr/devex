package list

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	cfg "vdex/config"
	plan "vdex/plan"
)

// Prints list of workspaces present
func ListWorkSpaces() {
	app := "terraform"
	cmdoutput, err := exec.Command(app, "workspace", "list").Output()
	if err == nil {
		fmt.Println(cmdoutput)
	}
}

// Prints list of systems present
func ListSystems(config *cfg.Config, myenv string) {
	//var fileList []string
	confPath := config.ConfPath
	entries, err := os.ReadDir(confPath)

	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("%-15s %-20s %-15s\n", "system-name", "conf-file", "environment")
	fmt.Printf("--------------- -------------------- ---------------\n")
	// loop over all system-names
	for _, v := range entries {
		if !v.IsDir() {
			continue
		}

		fmt.Printf("%-15s", v.Name())
		teamCfgPath := path.Join(confPath, v.Name())

		cfgentries, err := os.ReadDir(teamCfgPath)
		if err != nil {
			continue
		}

		// loop over all config files
		firstLine := true
		for _, cv := range cfgentries {
			if !strings.HasSuffix(cv.Name(), config.ConfFile) {
				continue
			} else if myenv != cfg.WORKSPACE_DEF && cv.Name() != config.GetConfFile(myenv) {
				continue
			}

			teamCfgFile := path.Join(teamCfgPath, cv.Name())

			if _, err := os.Stat(teamCfgFile); err == nil {
				log.Printf("File %s exists\n", teamCfgFile)

				reqWorkspace := plan.GetConfigWorkspace(teamCfgFile)
				if firstLine {
					fmt.Printf(" %-20s %-15s\n", cv.Name(), reqWorkspace)
				} else {
					fmt.Printf("%-15s %-20s %-15s\n", "", cv.Name(), reqWorkspace)
				}
				firstLine = false
			}
		}

	}
}

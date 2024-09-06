package init

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"vdex/parser"
)

/*
 * Writes the configuration data to the target location confPath + confFileName
 * Returns
 * error: if any failure
 */
func SaveConfig(parcedBlocks *parser.TFBlocks, confPath string, confFileName string) error {

	if _, err := os.Stat(confPath); os.IsNotExist(err) { // Create Path if not present
		err = os.Mkdir(confPath, 0755) //create a directory
		if err != nil {
			log.Println("Failed to create directory", confPath) //print the error on the console
			return err
		}
	}

	confFile := filepath.Join(confPath, confFileName)

	file, err := os.OpenFile(confFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Println("Failed to open file:", confFile)
		return err
	}
	defer file.Close()

	file.WriteString("# This is config file that contains the input values for the REPLACE-ME indicated variables in main.tf")
	file.WriteString("\n# Right hand side values can be edited. Please do not edit left hand side names")
	for k, v := range parcedBlocks.Param {
		//fmt.Printf("\n%s=>%s", k, v.P_value)
		file.WriteString("\n" + k + " = " + v.P_value)
	}
	return nil
}

/*
 * Prompts the user for the configuration data and saves
 * Returns
 * string: file location where the config is saved
 * error: if any failure
 */
func PromptConfig(parcedBlocks *parser.TFBlocks, confPath string, confFile string) (string, error) {
	var mvalue, sysName string
	n := len(parcedBlocks.Param)
	reader := bufio.NewReader(os.Stdin)

	if n == 0 {
		fmt.Println("\nNothing to be replaced in the file")
		return "", nil
	} else {
		fmt.Printf("\nThe terraform file needs %d user input values.", n)
		fmt.Printf("\n!!Please enter the value of each variable when prompted and press ENTER!!")
		fmt.Printf("\n!!To leave the default value unchanged, just Hit ENTER!!\n")
	}

	var userConfig map[string]string = make(map[string]string)

	for k, v := range parcedBlocks.Param {
		var err error
		n = 0
		maxAttempt := 3
		attempt := 0

		fmt.Printf("\n%s[default=%s]:", k, v.P_value)
		for attempt < maxAttempt {
			if v.P_type == parser.V_MAP_OR_SET || v.P_type == parser.V_LIST || v.P_type == parser.V_SCALAR {
				mvalue, err = reader.ReadString('\n')
				if err == nil && strings.TrimSpace(mvalue) != "" {
					n = 1
				}
			} else {
				n, err = fmt.Scanln(&mvalue)
				if err != nil {
				}
			}
			attempt++
			if n <= 0 && attempt < maxAttempt && v.P_value == parser.REPLACE2 {
				fmt.Printf("\n this param has no default value, input again attempt %d of %d:", attempt+1, maxAttempt)
			} else {
				break
			}
		}
		if n > 0 {
			userConfig[k] = strings.TrimSpace(mvalue)
			//v.P_value = strings.TrimSpace(mvalue)
			//parcedBlocks.Param[k] = v
		}
		if strings.HasSuffix(k, "tags.\"System-Name\"") {
			if n > 0 {
				sysName = strings.TrimSpace(mvalue)
			} else {
				sysName = v.P_value
			}
		}
	}

	for k, v := range userConfig {
		newParam := parcedBlocks.Param[k]
		newParam.P_value = v
		parcedBlocks.Param[k] = newParam
	}

	path := filepath.Join(confPath, strings.ReplaceAll(sysName, "\"", ""))
	confFFile := filepath.Join(path, confFile)

	return confFFile, SaveConfig(parcedBlocks, path, confFile)

}

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

type valueType int

// type of the terraform block
var blockTypes = []string{
	"terraform",
	"provider",
	"terraform",
	"data",
	"resource",
	"variable",
	"output",
	"locals",
	"terraform_backend_config",
	"#",
	"//",
	"/*",
	"\n"}

// Type of the Variable or Parameter Value
const (
	V_SCALAR     = 1
	V_NUMERIC    = 2
	V_BOOLEAN    = 3
	V_STRING     = 4
	V_LIST       = 5
	V_REFERANCE  = 6
	V_MAP_OR_SET = 7
	V_NULL       = 8
)

// Special Symbols in the terraform file
const (
	BLKBEGIN  byte   = '{'
	BLKEND    byte   = '}'
	BLKASSIGN byte   = '='
	STRDELIM  byte   = '"'
	ESCAPESEQ byte   = '\\'
	COMMENTST byte   = '/'
	LISTBEGIN byte   = '['
	LISTEND   byte   = ']'
	COMMENT   string = "//"
	REPLACE   string = "REPLACE-ME"
	REPLACE2  string = "\"REPLACE-ME\""
)

// TF Parameter Value structure
type ParamValue struct {
	// Value of the TF parameter
	p_value string
	// type of the value one of (BOOLEAN, NUMERIC, BOOLEAN, STRING, LIST, MAP etc)
	p_type valueType
	// Boolean indicating whether the Parameter is to be replaced
	p_replace bool
}

// structure to hold contents of a flat block like module
type ModuleBlock struct {
	// name of the block - "module <name>"
	blockName string
	// holds key value pairs of the block
	params map[string]ParamValue
}

// structure to hold contents of multi level block like provider
type TFBlock struct {
	// represents text which is present before the block - like comments
	prefix string
	// represents name of the block
	blockName string
	// represents full name of the block
	blockfName string
	// represents parent block
	parent *TFBlock
	// holds children that are sub blocks
	child []*TFBlock
	// holds key value pairs of the block
	params map[string]ParamValue
}

// structure to hold the hierarchical Parsed TFBlocks
type TFBlocks struct {
	// List of ModuleBlocks
	MList []ModuleBlock
	// List to hold TFBlock in a hierarchical/stack representation
	TFList []TFBlock
	// Map to store input param and the user input/config
	//param map[string]string
}

// structure holds the input data stream for Parsing
type TFParser struct {
	// scanner
	scan *bufio.Scanner
	// logger
	logger *log.Logger
	// pending text
	text string
}

// Return new ModuleBlock object
func CreateModuleBlock() ModuleBlock {
	p := ModuleBlock{}
	p.Init("")
	return p
}

// Initiate the maps of a ModuleBlock object
func (mb *ModuleBlock) Init(name string) {
	mb.blockName = name
	mb.params = make(map[string]ParamValue)
}

// Return new TFBlock object
func CreateTFBlock() TFBlock {
	p := TFBlock{}
	p.Init("")
	return p
}

// Initiate the maps of a TFBlock object
func (tfb *TFBlock) Init(name string) {
	tfb.blockName = name
	tfb.params = make(map[string]ParamValue)
}

// Clear the maps of a TFBlock object
func (tfb *TFBlock) Clear() {
	tfb.blockName = ""
	clear(tfb.params)
	tfb.child = tfb.child[:0]
}

/*
 * Takes string input along with current position and the lengh
 * Returns if the text contains double quoted value as boolean
 * and also returns the next byte position after the quoted value
 */
func (tfp *TFParser) ParseStringValue(text string, i int, n int) (int, bool) {
	i++
	for ; i < n && (text[i] != STRDELIM); i++ {
		// Handle value with escape sequence cases "Val \"50\""
		if text[i] == ESCAPESEQ {
			i++
		}
	}
	if i < n && text[i] == STRDELIM {
		return i + 1, true
	}
	return i, false
}

/*
 * Takes string input along with current position and the lengh - parse the list object
 * Returns
 * int index where the list ends in the multi/single last line
 * boolean if the text contains valid list object
 */
func (tfp *TFParser) ParseListValue(text string, i int, n int) (int, bool) {

	for i = i + 1; i < n && text[i] != LISTEND; {
		if text[i] == STRDELIM {
			i, _ = tfp.ParseStringValue(text, i, n)
		} else {
			i++
		}
	}

	if i < n && text[i] == LISTEND {
		return i + 1, true
	}
	return i, false
}

/*
 * Takes string input along with current position and the lengh - parse the multi line list object
 * Returns
 * string that holds the entire list
 * string that holds the last line that is part of a multi/single line list
 * int index where the list ends in the multi/single last line
 * boolean if the text contains valid list object
 */
func (tfp *TFParser) ParseMListValue(text string, i int, n int) (string, string, int, bool) {
	var listEnd bool

	ln := 0
	listtext := text

	ln, listEnd = tfp.ParseListValue(text, i, n)

	if !listEnd {
		ln = n
		for !listEnd && tfp.scan.Scan() {
			text = tfp.scan.Text()
			n = len(text)
			listtext += "\n" + text
			i, listEnd = tfp.ParseListValue(text, -1, n)
			ln += 1 + i
		}

	}
	return listtext, text, ln, listEnd
}

/*
 * Parses the input text (which is RHS of a variable) and returns ParamValue object that contains
 * -string the value of it after skimming undesired left and right such as comments, spaces
 * -type of the value one of (SCALER, BOOLEAN, STRING, LIST, MAP)
 * -bool indicating whether this value is eligible for ** REPLACE-IT **
 */
func (tfp *TFParser) ParseValue(itext string) ParamValue {
	var paramVal ParamValue

	paramVal.p_type = V_SCALAR

	var replaceEligible bool
	text := strings.TrimSpace(itext)
	listtext := text

	i := 0
	n := len(text)
	wstart := i
	wend := n

	switch text[i] {
	case STRDELIM: // Value in quotation eg: "5"
		wstart = i
		var isValid bool
		wend, isValid = tfp.ParseStringValue(text, i, n)
		if !isValid {
			log.Println("Invalid string format ", text)
		}
		paramVal.p_type = V_STRING
	case LISTBEGIN: // Value is a list/array
		var isValid bool
		//wend, isValid = tfp.ParseListValue(text, i, n)
		listtext, _, wend, isValid = tfp.ParseMListValue(text, i, n)
		text = listtext
		n = len(text)
		fmt.Printf("LISTBE:%s:%s:", text, string(text[wstart:wend]))
		if !isValid {
			log.Println("Invalid List format ", text)
		}
		paramVal.p_type = V_LIST
	case BLKBEGIN: // Value is either map or set
		for i = i + 1; i < n && text[i] != BLKEND; i++ {
			// Handle cases ["Val \[2\]", "Val2" ]
			if text[i] == ESCAPESEQ {
				i++
			}
		}
		paramVal.p_type = V_MAP_OR_SET
	default:
		// Numeric Case
		if text[i] >= '0' && text[i] <= '9' {
			j := strings.Index(text, "//")
			if j < 0 {
				j = strings.Index(text, " ")
			}
			if j > 0 {
				wend = j
			}
			paramVal.p_type = V_NUMERIC
		} else if text == "true" || text == "false" {
			paramVal.p_type = V_BOOLEAN
		}
	}
	// Value is suffixed with REPLACE-ME comment eg: foo = 5 // REPLACE-ME
	if wend < n {
		if strings.Contains(string(text[wend:n]), COMMENT) {
			replaceEligible = strings.HasSuffix(string(text[i:n]), REPLACE)
		}
	}
	// Value is "REPLACE-ME"
	if !replaceEligible && (string(text[wstart:wend]) == REPLACE2) {
		replaceEligible = true
	}
	paramVal.p_value = string(text[wstart:wend])
	paramVal.p_replace = replaceEligible
	return paramVal
}

// Return new Parser object
func CreateTFParser() TFParser {
	p := TFParser{}
	return p
}

/*
 * initializes the scanner object
 */
func (tfp *TFParser) SetScanner(scanner *bufio.Scanner) {
	tfp.scan = scanner
}

/*
 * initializes the logger object
 */
func (tfp *TFParser) SetLoger(logger *log.Logger) {
	tfp.logger = logger
}

/*
 * determines type of the terraform block
 * checks if the text matches any of the blockTypes and
 * returns index in the blockType array, -1 if not found
 * returns lenth of the block
 */
func (tfp *TFParser) ParseBlockType(itext string) (int, int) {
	i := 0
	blen := len(blockTypes)
	n := len(itext)

	text := strings.TrimSpace(itext)

	// check if the text indicates comment
	if strings.HasPrefix(text, "#") || strings.HasPrefix(text, "//") {
		if tfp.text != "" {
			tfp.text += "\n"
		}
		tfp.text += itext
		return 10, n
	} else if strings.HasPrefix(text, "/*") {

		if tfp.text != "" {
			tfp.text += "\n"
		}
		tfp.text += itext

		for !strings.HasSuffix(strings.TrimSpace(itext), "*/") && tfp.scan.Scan() {
			itext = tfp.scan.Text()
			tfp.text += "\n" + itext
		}
		return 11, n
	} else if text == "" { // only white spaces
		//tfp.text += "\n"
		return 12, n
	}

	j := strings.Index(text, " ")
	if j < 0 {
		j = strings.Index(text, "{")
		if j < 0 {
			j = n
		}
	}

	if j < n {
		text = text[0:j]
	}

	for ; i < blen && blockTypes[i] != text; i++ {
	}
	if i == blen { // text not found in the blockType
		i = -1
		j = 0
	}

	return i, j
}

/*
 * Parses the file pointed to by the scanner and populates TFBlock structure
 * Return:
 */
func (tfp *TFParser) ProcessStream(parsedData *TFBlocks) int {
	var i, s, nextpos, depth int
	var tfbp *TFBlock

	mb := CreateModuleBlock()

	// Parse the text for tokens and words
	for tfp.scan.Scan() {
		nextpos = 0
		text := tfp.scan.Text()
		n := len(text)

		// Process any comments or blank lines and store them
		bidx, _ := tfp.ParseBlockType(text)
		if bidx >= 10 {
			nextpos = n
		}

		// Process terraform blocks
		for i = nextpos; i < n; i++ {
			switch text[i] {
			case BLKBEGIN:
				var parentName string
				tfb := CreateTFBlock()

				if depth > 0 { // this is a sub block, add this as child to the current block
					tfbp.child = append(tfbp.child, &tfb)
					tfb.parent = tfbp
					parentName = tfbp.blockfName + "."
				}

				// point the current block to the newly crated block
				tfbp = &tfb

				tfbp.blockName = strings.TrimSpace(text[s:i])
				tfbp.blockfName = parentName + tfbp.blockName
				tfbp.prefix = tfp.text
				tfp.text = ""

				mb.blockName = text[s:i]
				depth++
				log.Printf("\nBLKBEGIN: %s::%d", tfbp.blockName, depth)

			case BLKEND:
				parsedData.MList = append(parsedData.MList, mb)
				if depth <= 1 {

				}
				log.Printf("\nBLKEND: %s::%d", tfbp.blockName, depth)

				depth--

				if depth > 0 {
					tfbp = tfbp.parent
				} else {
					parsedData.TFList = append(parsedData.TFList, *tfbp)
				}

			case BLKASSIGN:
				param := strings.TrimSpace(text[s:i])
				rhs := text[i+1 : n]
				value := tfp.ParseValue(rhs)

				if value.p_value != "{" {
					mb.params[param] = value
					tfbp.params[param] = value
					log.Printf("\n\tkey: %s:%s=>%v", tfbp.blockName, param, value)
					i = n
				}

			default:
			}
		}
	}
	return 0
}

func (parcedBlock *TFBlock) Walk(level int, ts int, logger *log.Logger) int {

	// print any free text like comments, blanks
	if parcedBlock.prefix != "" {
		fmt.Printf("\n%s", parcedBlock.prefix)
	}

	// the block name
	fmt.Printf("\n%s", strings.Repeat(" ", ts*level))
	fmt.Printf("%s %s", parcedBlock.blockfName, "{")

	// all parameters of the block
	for k, v := range parcedBlock.params {
		fmt.Printf("\n%s", strings.Repeat(" ", ts*(level+1)))
		fmt.Printf("%s = %s[%v][%d]", k, v.p_value, v.p_replace, v.p_type)
	}

	// all sub blocks
	childcount := len(parcedBlock.child)
	for j := 0; j < childcount; j++ {
		(parcedBlock.child[j]).Walk(level+1, ts, logger)
	}
	fmt.Printf("\n%s%s", strings.Repeat(" ", ts*level), "}")

	return 0
}

func (parcedBlocks *TFBlocks) Walk(level int, ts int, logger *log.Logger) int {
	var i, n int

	n = len(parcedBlocks.TFList)

	for i = 0; i < n; i++ {
		if i > 0 {
			logger.Printf("\n")
		}
		parcedBlocks.TFList[i].Walk(0, ts, logger)
	}

	return 0
}

/*
 * Takes file name and parses the stream and
 * constructs the Parsed Block structure along with the questions
 * Input: filename of the input tf module
 * Returns: constructed TFBlock structure
 */
func ParseTF(modfile string) (TFBlocks, error) {
	var tfb TFBlocks

	file, err := os.Open(modfile)
	if err != nil {
		fmt.Println("Failed to open file:", modfile)
		fmt.Println("Error:", err)
		return tfb, err
	}
	defer file.Close()

	tfp := CreateTFParser()
	scanner := bufio.NewScanner(file)

	tfp.SetLoger(log.Default())
	tfp.SetScanner(scanner)

	tfp.ProcessStream(&tfb)

	if err := scanner.Err(); err != nil {
		//log.Fatal(err)
		return tfb, err
	}
	return tfb, nil
}

package parser

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
	COMMENT1  string = "//"
	COMMENT2  string = "#"
	COMMENT3  string = "/*"
	REPLACE   string = "REPLACE-ME"
	REPLACE2  string = "\"REPLACE-ME\""
)

// TF Parameter Value structure
type ParamValue struct {
	// Value of the TF parameter
	P_value string
	// type of the value one of (BOOLEAN, NUMERIC, BOOLEAN, STRING, LIST, MAP etc)
	P_type valueType
	// Boolean indicating whether the Parameter is to be replaced
	P_replace bool
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
	Prefix string
	// represents name of the block
	BlockName string
	// represents full name of the block
	BlockfName string
	// type of the block to track MAP or SET
	BlockType int
	// represents parent block
	Parent *TFBlock
	// holds children that are sub blocks
	Child []*TFBlock
	// holds key value pairs of the block
	Params map[string]ParamValue
}

// structure to hold the hierarchical Parsed TFBlocks
type TFBlocks struct {
	// List of ModuleBlocks
	MList []ModuleBlock
	// List to hold TFBlock in a hierarchical/stack representation
	TFList []TFBlock
	// Map to store input param and the user input/config
	Param map[string]ParamValue
	// flag to indicate to fill Param
	Skip bool
}

// structure holds the input data stream for Parsing
type TFParser struct {
	// scanner
	scan *bufio.Scanner
	// logger
	file *os.File
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
	tfb.BlockName = name
	tfb.Params = make(map[string]ParamValue)
}

// Clear the maps of a TFBlock object
func (tfb *TFBlock) Clear() {
	tfb.BlockName = ""
	clear(tfb.Params)
	tfb.Child = tfb.Child[:0]
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

	paramVal.P_type = V_SCALAR

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
		paramVal.P_type = V_STRING
	case LISTBEGIN: // Value is a list/array
		var isValid bool
		//wend, isValid = tfp.ParseListValue(text, i, n)
		listtext, _, wend, isValid = tfp.ParseMListValue(text, i, n)
		text = listtext
		n = len(text)
		//fmt.Printf("LISTBE:%s:%s:", text, string(text[wstart:wend]))
		if !isValid {
			log.Println("Invalid List format ", text)
		}
		paramVal.P_type = V_LIST
	case BLKBEGIN: // Value is either map or set
		for i = i + 1; i < n && text[i] != BLKEND; i++ {
			// Handle cases ["Val \[2\]", "Val2" ]
			if text[i] == ESCAPESEQ {
				i++
			}
		}
		paramVal.P_type = V_MAP_OR_SET
	default:
		// Numeric Case
		if text[i] >= '0' && text[i] <= '9' {
			j := strings.Index(text, COMMENT1)
			if j < 0 {
				j = strings.Index(text, COMMENT2)
			}
			if j < 0 {
				j = strings.Index(text, " ")
			}
			if j > 0 {
				wend = j
			}
			paramVal.P_type = V_NUMERIC
		} else if text == "true" || text == "false" || strings.HasPrefix(text, "true") || strings.HasPrefix(text, "false") {
			j := strings.Index(text, COMMENT1)
			if j < 0 {
				j = strings.Index(text, COMMENT2)
			}
			if j < 0 {
				j = strings.Index(text, " ")
			}
			if j > 0 {
				wend = j
			} else {
				j = n
			}
			if strings.TrimSpace(string(text[:j])) == "true" || strings.TrimSpace(string(text[:j])) == "false" {
				paramVal.P_type = V_BOOLEAN
			}
		}
	}
	// Value is suffixed with REPLACE-ME comment eg: foo = 5 // REPLACE-ME
	if wend < n {
		if strings.Contains(string(text[wend:n]), COMMENT1) || strings.Contains(string(text[wend:n]), COMMENT2) {
			replaceEligible = strings.HasSuffix(string(text[i:n]), REPLACE)
		}
	}
	// Value is "REPLACE-ME"
	if !replaceEligible && (string(text[wstart:wend]) == REPLACE2) {
		replaceEligible = true
	}
	paramVal.P_value = strings.TrimSpace(string(text[wstart:wend]))
	paramVal.P_replace = replaceEligible
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
func (tfp *TFParser) SetLoger(outfile *os.File) {
	tfp.file = outfile
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
		tfp.text += itext
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
	var i, s, nextpos, depth, iteration int
	var tfbp *TFBlock

	log.Printf("\nIn ProcessStream")

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
			if tfp.file != nil {
				if iteration > 0 {
					fmt.Fprintf(tfp.file, "\n")
				}
				fmt.Fprintf(tfp.file, "%s", tfp.text)
				tfp.text = ""
			}
		} else {
			if tfp.file != nil && !strings.Contains(text, "=") {
				if iteration > 0 {
					fmt.Fprintf(tfp.file, "\n")
				}
				fmt.Fprintf(tfp.file, "%s", text)
			}
		}

		// Process terraform blocks
		for i = nextpos; i < n; i++ {
			switch text[i] {
			case BLKBEGIN:
				var parentName string
				tfb := CreateTFBlock()

				if depth > 0 { // this is a sub block, add this as child to the current block
					tfbp.Child = append(tfbp.Child, &tfb)
					tfb.Parent = tfbp
					parentName = tfbp.BlockfName + "."
				}

				// point the current block to the newly crated block
				tfbp = &tfb

				bname := strings.TrimSpace(text[s:i])
				//bname = strings.Replace(bname, " ", "", -1)
				if strings.HasSuffix(bname, "=") {
					tfbp.BlockType = V_MAP_OR_SET
					bname = strings.TrimSuffix(bname, "=")
					bname = strings.TrimSpace(bname)
				}

				tfbp.BlockName = bname
				tfbp.BlockfName = parentName + tfbp.BlockName
				tfbp.Prefix = tfp.text
				tfp.text = ""

				mb.blockName = text[s:i]
				depth++
				//log.Printf("\nBLKBEGIN: %s::%d", tfbp.BlockName, depth)

			case BLKEND:

				if depth < 1 { // invalid format
					log.Printf("\nInvalid block end %s", text)
					continue
				}
				parsedData.MList = append(parsedData.MList, mb)
				//log.Printf("\nBLKEND: %s::%d", tfbp.BlockName, depth)

				depth--

				if depth > 0 {
					tfbp = tfbp.Parent
				} else if tfbp != nil {
					parsedData.TFList = append(parsedData.TFList, *tfbp)
				}

			case BLKASSIGN:

				if depth < 1 { // invalid format
					log.Printf("\nInvalid = %s", text)
					continue
				}

				param := strings.TrimSpace(text[s:i])
				rhs := text[i+1 : n]
				value := tfp.ParseValue(rhs)

				if value.P_value != "{" {
					mb.params[param] = value
					tfbp.Params[param] = value
					if value.P_replace {
						if !parsedData.Skip {
							parsedData.Param[tfbp.BlockfName+"."+param] = value
						} else {
							value.P_value = parsedData.Param[tfbp.BlockfName+"."+param].P_value
						}
						log.Printf("\n\tkey: %s:%s=>%v", tfbp.BlockName, param, value)
					}
					if tfp.file != nil {
						fmt.Fprintf(tfp.file, "\n%s = %s", text[s:i], value.P_value)
					}

					i = n
				} else {
					if tfp.file != nil {
						fmt.Fprintf(tfp.file, "\n%s = %s", text[s:i], value.P_value)
					}
				}

			default:
			}
		}
		iteration++
	}
	return 0
}

func (parcedBlock *TFBlock) Walk(level int, i int, ts int, file *os.File, tfbs *TFBlocks) int {

	if i > 0 || level > 0 {
		fmt.Fprintf(file, "\n")
	}
	// print any free text like comments, blanks
	if parcedBlock.Prefix != "" {
		fmt.Fprintf(file, "%s\n", parcedBlock.Prefix)
	}

	// the block name
	fmt.Fprintf(file, "%s", strings.Repeat(" ", ts*level))
	sep := " "
	if parcedBlock.BlockType == V_MAP_OR_SET {
		sep = " = "
	}
	fmt.Fprintf(file, "%s%s%s", parcedBlock.BlockName, sep, "{")
	//fmt.Fprintf(file, "%s%s%s", parcedBlock.BlockfName, sep, "{")

	// all parameters of the block
	for k, v := range parcedBlock.Params {
		fmt.Fprintf(file, "\n%s", strings.Repeat(" ", ts*(level+1)))
		if tfbs.Skip && v.P_replace {
			user_key := parcedBlock.BlockfName + "." + k
			//fmt.Printf("\n%s :: %s", tfbs.Param[user_key].P_value, user_key)
			user_value := tfbs.Param[user_key].P_value
			if user_value != "" {
				v.P_value = user_value
			}
		}
		fmt.Fprintf(file, "%s = %s", k, v.P_value)
		//fmt.Fprintf(file, "%s = %s[%v][%d]", k, v.P_value, v.P_replace, v.P_type)
	}

	// all sub blocks
	childcount := len(parcedBlock.Child)
	for j := 0; j < childcount; j++ {
		(parcedBlock.Child[j]).Walk(level+1, i, ts, file, tfbs)
	}
	fmt.Fprintf(file, "\n%s%s", strings.Repeat(" ", ts*level), "}")

	return 0
}

func (parcedBlocks *TFBlocks) Walk(level int, ts int, file *os.File) int {
	var i, n int

	log.Printf("\nparcedBlocks In Walk")
	n = len(parcedBlocks.TFList)

	for i = 0; i < n; i++ {
		if i > 0 {
			fmt.Fprintf(file, "\n")
		}
		parcedBlocks.TFList[i].Walk(0, i, ts, file, parcedBlocks)
	}

	return 0
}

// Create and return TFBlocks object
func CreateTFBlocks() {
	tfbs := TFBlocks{}
	tfbs.Init()
}

// Initiate the maps of a TFBlocks object
func (tfbs *TFBlocks) Init() {
	tfbs.Param = make(map[string]ParamValue)
}

/*
 * Takes file name and parses the stream and
 * constructs the Parsed Block structure along with the questions
 * Input: filename of the input tf module
 * Returns: constructed TFBlock structure
 */
func ParseTF(modfile string, tfbp *TFBlocks, wfile *os.File) (*TFBlocks, error) {
	var tfbptr *TFBlocks

	log.Println("In ParseTF")
	file, err := os.Open(modfile)
	if err != nil {
		log.Println("Failed to open file:", modfile)
		log.Println("Error:", err)
		return nil, err
	}
	defer file.Close()

	if tfbp == nil {
		var tfb TFBlocks
		tfb.Init()
		tfbptr = &tfb
	} else {
		tfbptr = tfbp
	}

	tfp := CreateTFParser()
	scanner := bufio.NewScanner(file)

	tfp.SetLoger(wfile)
	tfp.SetScanner(scanner)

	tfp.ProcessStream(tfbptr)

	if err := scanner.Err(); err != nil {
		return tfbptr, err
	}
	return tfbptr, nil
}

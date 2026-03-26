package int

import (
	// "bufio"
	"errors"
	"fmt"
	"path/filepath"

	// "go/scanner"
	"os"
	"strings"
)

func UserInput() ([]string, bool){
	retVal := []string{}
	argslen := len(os.Args)
	if  argslen < 2 || argslen > 3 {
		fmt.Println("Missmatch in the input.")
		return nil, false
	}
	inputArgs := os.Args[1]	
	inputArgs = strings.TrimSpace(inputArgs)
	if !checker(inputArgs){
		fmt.Printf("file %s does not exist\n", inputArgs)
		return nil, false
	}
	if !(strings.HasSuffix(inputArgs, ".txt")) {
		fmt.Printf("%s Is not a .txt file.\n", inputArgs)
		return nil, false
	}
	retVal = append(retVal, inputArgs)
	usevisualizer := false
	if argslen == 3 {
		visualizer := os.Args[2]
		visualizer = strings.TrimSpace(visualizer)
		if visualizer == "visualizer" {
			usevisualizer = true
		}else {
			fmt.Printf("it is visualizer, not %s\n", visualizer)
		}
	}
	return retVal, usevisualizer
}

func checker(filename string) bool{
	filename  = strings.TrimSpace(filename)

	path := "./lem-in/tests"
	fullpath := filepath.Join(path,filename)

	_, err := os.Stat(fullpath)
	return !errors.Is(err, os.ErrNotExist)
}
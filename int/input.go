package int

import (
	"bufio"
	"fmt"
	"go/scanner"
	"os"
	"strings"
)

func UserInput() []string{
	retVal := []string{}
	if len(os.Args) != 2 {
		fmt.Println("More areguments than expected")
		return nil
	}
	inputArgs := os.Args[1]	
	inputArgs = strings.TrimSpace(inputArgs)
	if !(strings.HasSuffix(inputArgs, ".txt")) {
		fmt.Printf("%s Is not a .txt file.", inputArgs)
	}
	for _, c := range inputArgs{
		retVal = append(retVal, string(c))
	}
	return retVal
}
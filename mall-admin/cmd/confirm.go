package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// mustConfirm prints the prompt and returns true only if the user types "yes".
func mustConfirm(prompt string) bool {
	fmt.Println(prompt)
	fmt.Print("Type 'yes' to confirm: ")
	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(answer) == "yes"
}

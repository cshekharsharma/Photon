package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var passwordReader = term.ReadPassword

// Prompt displays a message and returns the user input.
func Prompt(message string) string {
	fmt.Print(message + " ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// PromptWithDefault returns the user input or a default if Enter is pressed.
func PromptWithDefault(message, defaultValue string) string {
	fmt.Printf("%s [%s]: ", message, defaultValue)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

// Confirm asks the user for yes/no confirmation.
func Confirm(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// PromptPassword securely asks for a hidden input (e.g. password).
func PromptPassword(message string) string {
	fmt.Print(message + " ")
	bytePassword, _ := passwordReader(int(os.Stdin.Fd()))
	fmt.Println() // print newline after input
	return strings.TrimSpace(string(bytePassword))
}

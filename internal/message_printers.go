/*
Defines a few functions that can be used to make the
printed messages displayed nicely.

Currently only supports unix-like systems.
*/

package internal

import (
	"fmt"
	"strings"

	"golang.org/x/term"
)

const MAX_MESSAGE_WIDTH = 55
const MAX_TOTAL_WIDTH = 58

func PrintRightJustifiedMessage(message string) {
	cols, _, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	border_string := rightJustifyText(cols, strings.Repeat("-", MAX_TOTAL_WIDTH-1))
	fmt.Println(border_string)
	message_parts := strings.Split(message, "\n")
	for _, message_part := range message_parts {
		message_chunks := batchMessage(message_part, MAX_MESSAGE_WIDTH)
		for _, chunk := range message_chunks {
			fmt.Println(rightJustifyText(cols, chunk))
		}
	}
	fmt.Println(border_string + "\n")
}

// Chops up the given message into chunks of size MAX_MESSAGE_WIDTH
// and prints them left justified with a border.
func PrintLeftJustifiedMessage(message string) {
	cols, _, err := term.GetSize(0)
	if err != nil {
		panic(err)
	}
	border_string := strings.Repeat("-", MAX_TOTAL_WIDTH) + "|"
	fmt.Println(border_string)
	for _, message_part := range strings.Split(message, "\n") {
		for _, chunk := range batchMessage(message_part, MAX_MESSAGE_WIDTH) {
			fmt.Println(leftJustifyText(cols, chunk))
		}
	}
	fmt.Println(border_string)
}

// Batches a message into chunks such that each chunk has length
// less than or equal to the max_width.
func batchMessage(message string, max_width int) []string {
	var batches []string
	for position := 0; position < len(message); position += max_width {
		end := position + max_width
		if end >= len(message) {
			end = len(message)
		}
		batches = append(batches, message[position:end])
	}
	return batches
}

// Returns the padded text to right justify when printing to a terminal
func rightJustifyText(cols int, text string) string {
	if cols < MAX_MESSAGE_WIDTH {
		return text
	}
	left_padding := strings.Repeat(" ", cols-MAX_TOTAL_WIDTH)
	right_padding := strings.Repeat(" ", (MAX_TOTAL_WIDTH-len(text))-1)
	return fmt.Sprintf("%s|%s%s", left_padding, right_padding, text)
}

// returns the original string with a '  |' appended to the end
// so that the pipe is aligned with the MAX_TOTAL_WIDTH column.
// Resulting text is '{original_text}{whitespace}|'
// where whitespace is length
func leftJustifyText(cols int, text string) string {
	if cols < MAX_MESSAGE_WIDTH {
		return text
	}
	padding := strings.Repeat(" ", MAX_TOTAL_WIDTH-len(text))
	return fmt.Sprintf("%s%s|", text, padding)
}

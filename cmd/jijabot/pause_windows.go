//go:build windows

package main

import (
	"bufio"
	"fmt"
	"os"
)

func wait() {
	fmt.Fprint(os.Stderr, "\nPress enter to continue...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

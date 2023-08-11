package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func inputHandler() {

	reader := bufio.NewReader(os.Stdin)

	for {

		text, _ := reader.ReadString('\n')

		text = strings.TrimSpace(text)

		if strings.Compare("status", text) == 0 {

			displayNodeStatus()

		} else if strings.Compare("version", text) == 0 {

			versionText := semverInfo()

			fmt.Println(versionText)

		} else {

			fmt.Println("")

		}
	}
}

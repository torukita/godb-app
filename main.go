package main

import (
	"fmt"
	"github.com/torukita/godb-app/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

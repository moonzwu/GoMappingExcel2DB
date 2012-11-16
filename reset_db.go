package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	path := os.Getenv("PATH")
	os.Setenv("PATH", path+":/usr/local/mysql/bin")
	args := [2][]string{{"-uroot", "-e", "drop database bmw"}, {"mysql", "-uroot", "-e", "create database bmw character set utf8 collate utf8_general_ci"}}

	for i := 0; i < len(args); i++ {
		command := exec.Command("mysql", args[i]...)
		err := command.Run()
		if err != nil {
			fmt.Println(err)
		}
	}

	fmt.Println("done")
}

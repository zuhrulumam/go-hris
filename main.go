package main

import (
	"github.com/joho/godotenv"
	"github.com/zuhrulumam/go-hris/cmd"
)

func main() {
	_ = godotenv.Load()
	cmd.Execute()
}

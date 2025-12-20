package main

import (
	"fmt"

	"github.com/gorax/gorax/internal/buildinfo"
)

func main() {
	info := buildinfo.GetInfo()
	fmt.Println(info.String())
}

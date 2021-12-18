package main

import (
	"archive/zip"
	"fmt"
	"log"
	"os"

	"github.com/juergenhoetzel/log4j2go/internal/log4j"
)

func main() {
	fmt.Println("{")
	for _, s := range os.Args[1:] {
		inputJar := s
		r, err := zip.OpenReader(inputJar)
		if err != nil {
			log.Fatal(err)
		}
		defer r.Close()

		key, ok := log4j.Log4jHash(r.File)
		if ok {
			fmt.Printf("%#v: %q,\n", key, inputJar)
		}
	}
	fmt.Println("}")
}

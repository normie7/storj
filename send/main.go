package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/normie7/storj/filesender"
)

func main() {

	address, filePath := parseArgs()

	// checking if file exists
	if err := checkFile(filePath); err != nil {
		log.Fatal(err)
	}

	// localhost:27001
	s, err := filesender.RegisterSender(address)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = s.Quit()
		if err != nil {
			log.Fatal(err)
		}
	}()

	fmt.Println(s.SecretKey())

	err = s.WaitForProxy()
	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	_, filename := path.Split(filePath)
	fi, err := f.Stat()
	if err != nil {
		log.Println(err)
		return
	}

	if fi.IsDir() {
		log.Println(f.Name(), "is directory")
		return
	}

	_, err = s.SendFile(filename, fi.Size(), f)
	if err != nil {
		log.Println(f.Name(), "is directory")
		return
	}
	return
}

func parseArgs() (address, filePath string) {
	if len(os.Args) != 3 {
		log.Fatal("expected 'address' and 'filePath' arguments")
	}

	if len(strings.Split(os.Args[1], ":")) != 2 {
		log.Fatal("wrong address format. try relayHost:relayPort")
	}

	return os.Args[1], os.Args[2]
}

func checkFile(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("%s is a directory", filePath)
	}

	return nil
}

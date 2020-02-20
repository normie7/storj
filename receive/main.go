package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/normie7/storj/filesender"
)

func main() {

	address, secretCode, dir := parseArgs()

	if err := checkDir(dir); err != nil {
		log.Fatal(err)
	}

	r, err := filesender.RegisterReceiver(address, secretCode)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = r.Quit()
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = r.AskForFile()
	if err != nil {
		log.Println(err)
		return
	}

	fileName, fileSize, reader, err := r.ReceiveFile()
	if err != nil {
		log.Println(err)
		return
	}

	// don't want to accidentally erase existing file
	f, err := os.OpenFile(path.Join(dir, fileName), os.O_RDWR|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	_, err = io.CopyN(f, reader, fileSize)
	if err != nil {
		log.Println(err)
		return
	}

	return
}

func parseArgs() (address, secretCode, dir string) {
	if len(os.Args) != 4 {
		log.Fatal("expected 'address' and 'secretCode' and 'dir' arguments")
	}

	if len(strings.Split(os.Args[1], ":")) != 2 {
		log.Fatal("wrong address format. try relayHost:relayPort")
	}

	return os.Args[1], os.Args[2], os.Args[3]
}

func checkDir(dir string) error {
	fi, err := os.Stat(dir)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	return nil
}

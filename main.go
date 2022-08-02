package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/alessio/shellescape"
)

func main() {
	if len(os.Args) < 1 {
		log.Fatal("Please provide playlist file path\n")
	}

	playlistFilePath := os.Args[1]

	file, err := os.Open(playlistFilePath)
	if err != nil {
		log.Fatalf("Error reading file: %+v\n", err)
	}

	defer file.Close()

	artistsMap := map[string]string{}

	r := csv.NewReader(file)

	for {

		record, err := r.Read()

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error while reading: %+v", err)
		}

		if record[1] == "Artist Name" {
			continue
		}

		artistsMap[record[1]] = shellescape.Quote(record[1])
	}

	commandFormat := "tidal-dl -s %s"
	//commandFormat := "echo %s"

	failed := []string{}
	for _, artist := range artistsMap {
		log.Printf("Artist: %s", artist)

		cmd := exec.Command("bash", "-c", fmt.Sprintf(commandFormat, artist))

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			failed = append(failed, artist)
			log.Printf("CMD: %s", cmd.String())
			log.Printf("Error in STD out pipe: %s", err)
			return
		}

		cmd.Start()
		go print(stdout)
		cmd.Wait()

		// Sleep for two seconds to not spam the api
		time.Sleep(2 * time.Second)
	}

	log.Print("Failed artists:")
	log.Print(failed)
}

// to print the processed information when stdout gets a new line
func print(stdout io.ReadCloser) {
	r := bufio.NewReader(stdout)
	line, _, err := r.ReadLine()
	fmt.Printf("\t%s\n", string(line))
	if err != nil {
		log.Printf("\tERROR: %s", err.Error())
	}
}

package main

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
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

	failed := []string{}
	artistsAmount := len(artistsMap)
	counter := 0

	for _, artist := range artistsMap {
		counter++

		log.Printf("Artist: %s", artist)
		log.Printf("%d out of %d", counter, artistsAmount)

		cmd := exec.Command("bash", "-c", fmt.Sprintf(commandFormat, artist))

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			failed = append(failed, artist)
			log.Printf("CMD: %s", cmd.String())
			log.Printf("Error in STD out pipe: %s", err)
			err = writeLog("/mnt/3TB/failedLog.txt", artist)
			if err != nil {
				log.Fatalf("Error writting log %s", err.Error())
			}
		}

		cmd.Start()
		err = print(stdout)
		if err != nil {
			failed = append(failed, artist)
			log.Printf("CMD: %s", cmd.String())
			log.Printf("Error in STD out pipe: %s", err)
			err = writeLog("/mnt/3TB/failedLog.txt", artist)
			if err != nil {
				log.Fatalf("Error writting log %s", err.Error())
			}
		}
		cmd.Wait()

		// Sleep for 5 seconds to not spam the api
		time.Sleep(5 * time.Second)
	}

	log.Print("Failed artists:")
	log.Print(failed)
}

func writeLog(path string, failedArtists string) error {
	failedLogFile, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		return err
	}

	defer failedLogFile.Close()

	failedLogFile.WriteString(failedArtists)
	return nil
}

func print(stdout io.ReadCloser) error {
	r := bufio.NewReader(stdout)

	for {
		line, _, err := r.ReadLine()
		if err != nil {
			log.Printf("\tERROR: %s", err.Error())
			return nil
		}

		if line == nil {
			return nil
		}
		fmt.Printf("\t%s\n", string(line))

		if strings.Contains(string(line), "ERR") {
			return errors.New("Err with this artist")
		}
	}
	return nil
}

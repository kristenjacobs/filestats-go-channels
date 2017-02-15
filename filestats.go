package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"unicode"
)

type StatFunc func(c chan *string, waitGroup *sync.WaitGroup)
type StatChan chan *string
type Stat struct {
	f StatFunc
	c StatChan
}

func LineCount(c chan *string, waitGroup *sync.WaitGroup) {
	var lineCount = 0
	for {
		line, more := <-c
		if more {
			_ = line
			lineCount++
		} else {
			fmt.Printf("The line count is: %d\n", lineCount)
			waitGroup.Done()
			return
		}
	}
}

func WordCount(c chan *string, waitGroup *sync.WaitGroup) {
	var wordCount = 0
	for {
		line, more := <-c
		if more {
			wordCount += len(strings.Fields(*line))
		} else {
			fmt.Printf("The word count is: %d\n", wordCount)
			waitGroup.Done()
			return
		}
	}
}

func AverageLettersPerWord(c chan *string, waitGroup *sync.WaitGroup) {
	var numLetters = 0
	var numWords = 0
	for {
		line, more := <-c
		if more {
			for _, char := range *line {
				if unicode.IsLetter(char) {
					numLetters++
				}
			}
			numWords += len(strings.Fields(*line))
		} else {
			var alpw = 0.0
			if numWords > 0 {
				alpw = float64(numLetters) / float64(numWords)
			}
			fmt.Printf("The average number of letters per word is: %.2f\n", alpw)
			waitGroup.Done()
			return
		}
	}
}

func MostCommonLetter(c chan *string, waitGroup *sync.WaitGroup) {
	var letterFrequencyMap = make(map[rune]int)
	for {
		line, more := <-c
		if more {
			for _, char := range *line {
				if unicode.IsLetter(char) {
					letterFrequencyMap[unicode.ToLower(char)]++
				}
			}
		} else {
			var maxVal = -1
			var maxKey = 'a'
			for key, val := range letterFrequencyMap {
				if val > maxVal {
					maxKey = key
					maxVal = val
				}
			}
			if maxVal != -1 {
				fmt.Printf("Most common letter is: %c\n", maxKey)
			} else {
				fmt.Printf("Could not calculate the most common letter\n")
			}
			waitGroup.Done()
			return
		}
	}
}

func openFile(fileName string) *os.File {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func startStats(stats []Stat) *sync.WaitGroup {
	waitGroup := &sync.WaitGroup{}
	waitGroup.Add(len(stats))
	for _, stat := range stats {
		go stat.f(stat.c, waitGroup)
	}
	return waitGroup
}

func processFile(fileName string, stats []Stat) {
	file := openFile(fileName)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var line = scanner.Text()
		for _, stat := range stats {
			stat.c <- &line
		}
	}
	file.Close()
}

func stopStats(stats []Stat, waitGroup *sync.WaitGroup) {
	for _, stat := range stats {
		close(stat.c)
	}
	waitGroup.Wait()
}

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Printf("Error: Invalid arguments.\n")
		fmt.Printf("  Usage:\n")
		fmt.Printf("    filestats-go <file>\n")
		os.Exit(1)
	}
	fileName := args[0]

	stats := []Stat{
		Stat{LineCount, make(chan *string)},
		Stat{WordCount, make(chan *string)},
		Stat{MostCommonLetter, make(chan *string)},
		Stat{AverageLettersPerWord, make(chan *string)}}

	waitGroup := startStats(stats)
	processFile(fileName, stats)
	stopStats(stats, waitGroup)
}

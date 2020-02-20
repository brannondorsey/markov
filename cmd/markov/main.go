package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	wr "github.com/mroth/weightedrand"
	flag "github.com/spf13/pflag"
)

func main() {

	args := parseArgs()
	seed := args.Prompt

	rand.Seed(time.Now().UTC().UnixNano()) // always seed random!
	hist, err := LoadOrCreateHistogram(args.InputFilename, args.N)
	PanicOnError(err)
	sample := GetSamplerFromStringHistogram(hist)
	// PrintStringHistogram(hist)
	// PrintMemUsage()
	// runtime.GC()
	for i := 0; i < args.MaxCharacters; i++ {
		next, err := sample(seed[len(seed)-args.N:])
		if err != nil {
			break
		}
		seed += next
	}
	// PrintMemUsage()
	fmt.Println(seed)
}

type StringHistogram = map[string]map[string]uint32

func PanicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func BuildStringHistogram(r io.Reader, n int) *StringHistogram {
	frequency := make(StringHistogram)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanRunes)
	buf := make([]string, n)
	for scanner.Scan() {
		text := scanner.Text()
		buf = append(buf, text)
		if len(buf) > n*2+1 {
			buf = buf[1:]
			lower := strings.ToLower(strings.Join(buf, ""))
			gram := lower[0:n]
			nextGram := lower[n : len(lower)-1]
			// fmt.Printf("lower: %v, gram: %v, nextGram: %v\n", lower, gram, nextGram)
			if _, ok := frequency[gram]; !ok {
				frequency[gram] = make(map[string]uint32)
			}
			frequency[gram][nextGram]++
		}
	}
	return &frequency
}

func GetSamplerFromStringHistogram(hist *StringHistogram) func(string) (string, error) {
	samplers := make(map[string]*wr.Chooser)
	for gram := range *hist {
		nextGrams := (*hist)[gram]
		choices := make([]wr.Choice, len(nextGrams))
		i := 0
		for key := range nextGrams {
			// fmt.Println(i, key, (*hist)[key])
			choices[i] = wr.Choice{
				Item:   key,
				Weight: uint(nextGrams[key]),
			}
			i++
		}
		chooser := wr.NewChooser(choices...)
		samplers[gram] = &chooser
	}
	return func(search string) (string, error) {
		if _, ok := samplers[search]; !ok {
			return "", fmt.Errorf("sample error: %v was not present in the histogram", search)
		}
		return samplers[search].Pick().(string), nil
	}
}

func PrintStringHistogram(h *StringHistogram) {
	for key := range *h {
		fmt.Printf("%v: %v\n", key, (*h)[key])
	}
}

func LoadOrCreateHistogram(filename string, n int) (*StringHistogram, error) {
	cacheFilename := fmt.Sprintf("%v.cache.n%d.json", filename, n)
	cacheFile, err := os.Open(cacheFilename)
	// Load from cache
	if err == nil {
		defer cacheFile.Close()
		cacheBytes, err := ioutil.ReadAll(cacheFile)
		if err != nil {
			return nil, err
		}
		hist := make(StringHistogram)
		err = json.Unmarshal(cacheBytes, &hist)
		if err != nil {
			return nil, err
		}
		return &hist, nil
	} else if os.IsNotExist(err) {
		// Build histogram and save cache
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		hist := BuildStringHistogram(file, n)
		err = CacheHistogram(hist, cacheFilename)
		if err != nil {
			return nil, err
		}
		return hist, nil
	}
	return nil, err
}

func CacheHistogram(histogram *StringHistogram, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	serialized, err := json.Marshal(histogram)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(serialized))
	return err
}

type arguments struct {
	InputFilename string
	Prompt        string
	N             int
	MaxCharacters int
}

func parseArgs() arguments {
	corpusFilename := flag.StringP("corpus", "i", "", "The input corpus to build the n-gram histogram with.")
	prompt := flag.StringP("prompt", "p", "hello", "The prompt to (optional).")
	n := flag.IntP("n-gram-length", "n", 1, "The number of characters to use for each n-gram.")
	maxCharacters := flag.IntP("max-characters", "c", 1000, "The maximum number of characters to generate. Fewer characters may be generated if the sequence encounters an n-gram that has no next n-grams in the dataset.")
	help := flag.BoolP("help", "h", false, "Show this screen.")
	flag.Parse()
	if flag.NArg() != 0 || *help {
		flag.Usage()
		os.Exit(1)
	}
	if *n < 1 || *n > 6 {
		fmt.Printf("[ERROR] The value of --n-gram-length must be between 1 and 6. Received %d.\n", *n)
		os.Exit(1)
	}
	if *corpusFilename == "" {
		fmt.Printf("[ERROR] The --corpus flag is required.\n")
		flag.Usage()
		os.Exit(1)
	}
	if _, err := os.Stat(*corpusFilename); os.IsNotExist(err) {
		fmt.Printf("[ERROR] Corpus file \"%s\" does not exist.\n", *corpusFilename)
		os.Exit(1)
	}
	return arguments{
		InputFilename: *corpusFilename,
		Prompt:        *prompt,
		N:             *n,
		MaxCharacters: *maxCharacters,
	}
}

// PrintMemUsage outputs the current, total and OS memory being used. As well as the number
// of garage collection cycles completed.
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

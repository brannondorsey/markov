package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"time"

	wr "github.com/mroth/weightedrand"
	flag "github.com/spf13/pflag"
)

type StringHistogram = map[string]map[string]uint32

func main() {

	args := parseArgs()

	rand.Seed(time.Now().UTC().UnixNano()) // always seed random!
	hist, err := LoadOrCreateHistogram(args.InputFilename, args.N, args.Lowercase, args.Words)
	if err != nil {
		panic(err)
	}
	sample := GetSamplerFromStringHistogram(hist)
	generated := make([]string, 0, args.Max)
	generated = append(generated, GetSeed(args.Prompt, args.N, args.Lowercase, args.Words, hist)...)
	for i := 0; i < args.Max; i++ {
		next, err := sample(generated[len(generated)-1])
		if err != nil {
			break
		}
		generated = append(generated, next)
	}
	fmt.Println(strings.Join(generated, GetSeparator(args.Words)))
}

// GetSeparator returns " " if words is true, "" otherwise
func GetSeparator(words bool) string {
	if words {
		return " "
	} else {
		return ""
	}
}

// GetSeed splits prompt into n-grams if prompt is usable or returns a random n-gram if not
func GetSeed(prompt string, n int, lower bool, words bool, hist StringHistogram) []string {
	var seed []string
	separator := GetSeparator(words)
	if lower {
		prompt = strings.ToLower(prompt)
	}
	// If the prompt contains at least one n-gram's worth of text
	if promptSplit := strings.Split(prompt, separator); prompt != "" && len(promptSplit) >= n {
		// Split it on the last n-gram
		first := strings.Join(promptSplit[:len(promptSplit)-n], separator)
		last := strings.Join(promptSplit[len(promptSplit)-n:], separator)
		// fmt.Printf("first: %v, last: %v\n", first, last)
		// And if the ngram appears in the corpus histogram
		if _, ok := hist[last]; ok {
			// Return []string{first, last}
			if first != "" {
				seed = append(seed, first)
			}
			seed = append(seed, last)
			return seed
		}
	}
	// Use the first random ngram that contains at least one child
	for randNgram := range hist {
		if len(hist[randNgram]) < 1 {
			continue
		}
		seed = []string{randNgram}
		break
	}
	return seed
}

func BuildStringHistogram(r io.Reader, n int, lowercase bool, words bool) StringHistogram {
	frequency := make(StringHistogram)
	scanner := bufio.NewScanner(r)
	separator := GetSeparator(words)
	if words {
		scanner.Split(bufio.ScanWords)
	} else {
		scanner.Split(bufio.ScanRunes)
	}
	buf := make([]string, 0, n)
	for scanner.Scan() {
		text := scanner.Text()
		buf = append(buf, text)
		if len(buf) > n*2 {
			gram := strings.Join(buf[0:n], separator)
			nextGram := strings.Join(buf[n:len(buf)-1], separator)
			if lowercase {
				gram = strings.ToLower(gram)
				nextGram = strings.ToLower(nextGram)
			}
			// fmt.Printf("gram: %v, nextGram: %v\n", gram, nextGram)
			if _, ok := frequency[gram]; !ok {
				frequency[gram] = make(map[string]uint32)
			}
			frequency[gram][nextGram]++
			buf = buf[1:]
		}
	}
	return frequency
}

func GetSamplerFromStringHistogram(hist StringHistogram) func(string) (string, error) {
	samplers := make(map[string]*wr.Chooser)
	for gram := range hist {
		nextGrams := hist[gram]
		choices := make([]wr.Choice, len(nextGrams))
		i := 0
		for key := range nextGrams {
			// fmt.Println(i, key, hist[key])
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

// func PrintStringHistogram(hist StringHistogram) {
// 	for key := range hist {
// 		fmt.Printf("%v: %v\n", key, hist[key])
// 	}
// }

func LoadOrCreateHistogram(filename string, n int, lowercase bool, words bool) (StringHistogram, error) {
	lowercaseString := ""
	if lowercase {
		lowercaseString = "lower"
	}
	wordsString := ""
	if words {
		wordsString = "words"
	}
	cacheFilename := fmt.Sprintf("%v.cache.n%d%s%s.json", filename, n, lowercaseString, wordsString)
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
		return hist, nil
	} else if os.IsNotExist(err) {
		// Build histogram and save cache
		file, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		hist := BuildStringHistogram(file, n, lowercase, words)
		err = CacheHistogram(hist, cacheFilename)
		if err != nil {
			return nil, err
		}
		return hist, nil
	}
	return nil, err
}

func CacheHistogram(histogram StringHistogram, filename string) error {
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
	Max           int
	Lowercase     bool
	Words         bool
}

func parseArgs() arguments {
	prompt := flag.StringP("prompt", "p", "", "The prompt to use.")
	n := flag.IntP("n-gram-length", "n", 3, "The number of characters to use for each n-gram.")
	max := flag.IntP("max", "m", 1000, "The maximum number of n-gram tokens to generate. Fewer characters may begenerated if\nthe sequence encounters an n-gram that has no next n-grams in the dataset.")
	help := flag.BoolP("help", "h", false, "Show this screen.")
	lowercase := flag.BoolP("lowercase", "l", false, "Convert text to lowercase. Lowers the complexity of the sampling task, and may produce\nbetter results depending on the corpus.")
	words := flag.BoolP("words", "w", false, "Use word-level n-grams instead of character-level n-grams.")

	flag.Parse()
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] <input-file> ...\n", os.Args[0])
		fmt.Println("Note: <input-file> is a required positional argument.")
		flag.PrintDefaults()
	}
	if flag.NArg() != 1 || *help {
		flag.Usage()
		os.Exit(1)
	}
	if *n < 1 || *n > 6 {
		fmt.Printf("[ERROR] The value of --n-gram-length must be between 1 and 6. Received %d.\n", *n)
		os.Exit(1)
	}
	inputFilename := flag.Args()[0]
	if fileInfo, err := os.Stat(inputFilename); os.IsNotExist(err) || fileInfo.IsDir() {
		if fileInfo != nil && fileInfo.IsDir() {
			fmt.Printf("[ERROR] Input file \"%s\" must be a file, not a directory.\n", inputFilename)
		} else {
			fmt.Printf("[ERROR] Input file \"%s\" does not exist.\n", inputFilename)
		}
		os.Exit(1)
	}
	return arguments{
		InputFilename: inputFilename,
		Prompt:        *prompt,
		N:             *n,
		Max:           *max,
		Lowercase:     *lowercase,
		Words:         *words,
	}
}

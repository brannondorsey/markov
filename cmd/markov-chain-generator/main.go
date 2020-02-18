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
)

const NGRAMS = 3
const MAX_CHARACTERS_TO_GENERATE = 1000

func main() {
	rand.Seed(time.Now().UTC().UnixNano()) // always seed random!
	filename := "data/rockyou-train.txt"
	hist, err := LoadOrCreateHistogram(filename, NGRAMS)
	PanicOnError(err)
	sample := GetSamplerFromStringHistogram(hist)
	seed := "once upon a time in wonderland "
	for i := 0; i < MAX_CHARACTERS_TO_GENERATE; i++ {
		next, err := sample(seed[len(seed)-NGRAMS:])
		if err != nil {
			break
		}
		seed += next
	}
	fmt.Println(seed)
}

type StringHistogram = map[string]map[string]uint

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
				frequency[gram] = make(map[string]uint)
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
				Weight: nextGrams[key],
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
		hist := BuildStringHistogram(file, NGRAMS)
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

package main

import (
	"reflect"
	"strings"
	"testing"
)

// func Test_main(t *testing.T) {
// 	tests := []struct {
// 		name string
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			main()
// 		})
// 	}
// }

func TestGetSeparator(t *testing.T) {
	type args struct {
		words bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Word separator", args{words: true}, " "},
		{"Character separator", args{words: false}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSeparator(tt.args.words); got != tt.want {
				t.Errorf("GetSeparator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSeed(t *testing.T) {
	type args struct {
		prompt string
		n      int
		lower  bool
		words  bool
		hist   StringHistogram
	}

	text := "Hello world! This is a text string to be used during testing. Its short."
	CharHists := make(map[int]StringHistogram)
	CharLowerHists := make(map[int]StringHistogram)
	WordHists := make(map[int]StringHistogram)
	WordLowerHists := make(map[int]StringHistogram)

	for n := 1; n <= 6; n++ {
		CharHists[n] = BuildStringHistogram(strings.NewReader(text), n, false, false)
		CharLowerHists[n] = BuildStringHistogram(strings.NewReader(text), n, true, false)
		WordHists[n] = BuildStringHistogram(strings.NewReader(text), n, false, true)
		WordLowerHists[n] = BuildStringHistogram(strings.NewReader(text), n, true, true)
	}

	tests := []struct {
		name string
		args args
		want []string
	}{
		{"n=1, usable prompt", args{prompt: "H", n: 1, lower: false, words: false, hist: CharHists[1]}, []string{"H"}},
		{"n=1, usable prompt", args{prompt: "H", n: 1, lower: true, words: false, hist: CharLowerHists[1]}, []string{"h"}},
		{"n=1, usable prompt", args{prompt: "Hello", n: 1, lower: false, words: true, hist: WordHists[1]}, []string{"Hello"}},
		{"n=1, usable prompt", args{prompt: "Hello", n: 1, lower: true, words: true, hist: WordLowerHists[1]}, []string{"hello"}},

		{"n=2, usable prompt", args{prompt: "He", n: 2, lower: false, words: false, hist: CharHists[2]}, []string{"He"}},
		{"n=2, usable prompt", args{prompt: "He", n: 2, lower: true, words: false, hist: CharLowerHists[2]}, []string{"he"}},
		{"n=2, usable prompt", args{prompt: "Hello world!", n: 2, lower: false, words: true, hist: WordHists[2]}, []string{"Hello world!"}},
		{"n=2, usable prompt", args{prompt: "Blah blah Hello world!", n: 2, lower: true, words: true, hist: WordLowerHists[2]}, []string{"blah blah", "hello world!"}},

		{"n=3, usable prompt", args{prompt: "some characters before the promptHel", n: 3, lower: false, words: false, hist: CharHists[3]}, []string{"some characters before the prompt", "Hel"}},
		{"n=3, usable prompt", args{prompt: "Hel", n: 3, lower: true, words: false, hist: CharLowerHists[3]}, []string{"hel"}},
		{"n=3, usable prompt", args{prompt: "This text is everything before Hello world! This", n: 3, lower: false, words: true, hist: WordHists[3]}, []string{"This text is everything before", "Hello world! This"}},
		{"n=3, usable prompt", args{prompt: "This text is everything before Hello world! This", n: 3, lower: true, words: true, hist: WordLowerHists[3]}, []string{"this text is everything before", "hello world! this"}},

		{"n=4, usable prompt", args{prompt: "Some other totally unrelated characters: Hell", n: 4, lower: false, words: false, hist: CharHists[4]}, []string{"Some other totally unrelated characters: ", "Hell"}},
		{"n=4, usable prompt", args{prompt: "Hell", n: 4, lower: true, words: false, hist: CharLowerHists[4]}, []string{"hell"}},
		{"n=4, usable prompt", args{prompt: "This text is everything before Hello world! This is", n: 4, lower: false, words: true, hist: WordHists[4]}, []string{"This text is everything before", "Hello world! This is"}},
		{"n=4, usable prompt", args{prompt: "This text is everything before Hello world! This is", n: 4, lower: true, words: true, hist: WordLowerHists[4]}, []string{"this text is everything before", "hello world! this is"}},

		{"n=5, usable prompt", args{prompt: "Some other totally unrelated characters: Hello", n: 5, lower: false, words: false, hist: CharHists[5]}, []string{"Some other totally unrelated characters: ", "Hello"}},
		{"n=5, usable prompt", args{prompt: "Hello", n: 5, lower: true, words: false, hist: CharLowerHists[5]}, []string{"hello"}},
		{"n=5, usable prompt", args{prompt: "This text is everything before Hello world! This is a", n: 5, lower: false, words: true, hist: WordHists[5]}, []string{"This text is everything before", "Hello world! This is a"}},
		{"n=5, usable prompt", args{prompt: "This text is everything before Hello world! This is a", n: 5, lower: true, words: true, hist: WordLowerHists[5]}, []string{"this text is everything before", "hello world! this is a"}},

		{"n=6, usable prompt", args{prompt: "Some other totally unrelated characters: Hello world!", n: 6, lower: false, words: false, hist: CharHists[6]}, []string{"Some other totally unrelated characters: Hello ", "world!"}},
		{"n=6, usable prompt", args{prompt: "Hello ", n: 6, lower: true, words: false, hist: CharLowerHists[6]}, []string{"hello "}},
		{"n=6, usable prompt", args{prompt: "This text is everything before Hello world! This is a text", n: 6, lower: false, words: true, hist: WordHists[6]}, []string{"This text is everything before", "Hello world! This is a text"}},
		{"n=6, usable prompt", args{prompt: "This text is everything before Hello world! This is a text", n: 6, lower: true, words: true, hist: WordLowerHists[6]}, []string{"this text is everything before", "hello world! this is a text"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSeed(tt.args.prompt, tt.args.n, tt.args.lower, tt.args.words, tt.args.hist); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSeed() = %v, want %v", got, tt.want)
			}
		})
	}

	tests = []struct {
		name string
		args args
		want []string
	}{
		{"n=1, unusable prompt", args{prompt: "Z", n: 1, lower: false, words: false, hist: CharHists[1]}, []string{"H"}},
		{"n=1, unusable prompt", args{prompt: "z", n: 1, lower: true, words: false, hist: CharLowerHists[1]}, []string{"z"}},
		{"n=1, unusable prompt", args{prompt: "HelZo", n: 1, lower: false, words: true, hist: WordHists[1]}, []string{"HelZo"}},
		{"n=1, unusable prompt", args{prompt: "HelZo", n: 1, lower: true, words: true, hist: WordLowerHists[1]}, []string{"helZo"}},

		{"n=2, unusable prompt", args{prompt: "Ze", n: 2, lower: false, words: false, hist: CharHists[2]}, []string{"Ze"}},
		{"n=2, unusable prompt", args{prompt: "Ze", n: 2, lower: true, words: false, hist: CharLowerHists[2]}, []string{"ze"}},
		{"n=2, unusable prompt", args{prompt: "Hell0 world!", n: 2, lower: false, words: true, hist: WordHists[2]}, []string{"Hell world!"}},
		{"n=2, unusable prompt", args{prompt: "Blah blah Hell0 world!", n: 2, lower: true, words: true, hist: WordLowerHists[2]}, []string{"blah blah", "Hell0 world!"}},

		{"n=3, unusable prompt", args{prompt: "some characters before the promptHal", n: 3, lower: false, words: false, hist: CharHists[3]}, []string{"some characters before the prompt", "Hal"}},
		{"n=3, unusable prompt", args{prompt: "Hil", n: 3, lower: true, words: false, hist: CharLowerHists[3]}, []string{"hil"}},
		{"n=3, unusable prompt", args{prompt: "This text is everything before Hello world! I", n: 3, lower: false, words: true, hist: WordHists[3]}, []string{"This text is everything before", "Hello world! I"}},
		{"n=3, unusable prompt", args{prompt: "This text is everything before Hello world! I", n: 3, lower: true, words: true, hist: WordLowerHists[3]}, []string{"this text is everything before", "hello world! I"}},

		{"n=4, unusable prompt", args{prompt: "Some other totally unrelated characters: Hel1", n: 4, lower: false, words: false, hist: CharHists[4]}, []string{"Some other totally unrelated characters: ", "Hel1"}},
		{"n=4, unusable prompt", args{prompt: "Hel1", n: 4, lower: true, words: false, hist: CharLowerHists[4]}, []string{"hel1"}},
		{"n=4, unusable prompt", args{prompt: "This text is everything before Hello world! This was", n: 4, lower: false, words: true, hist: WordHists[4]}, []string{"This text is everything before", "Hello world! This was"}},
		{"n=4, unusable prompt", args{prompt: "This text is everything before Hello world! This was", n: 4, lower: true, words: true, hist: WordLowerHists[4]}, []string{"this text is everything before", "hello world! this was"}},

		{"n=5, unusable prompt", args{prompt: "Some other totally unrelated characters: Dudey", n: 5, lower: false, words: false, hist: CharHists[5]}, []string{"Some other totally unrelated characters: ", "Dudey"}},
		{"n=5, unusable prompt", args{prompt: "wowie", n: 5, lower: true, words: false, hist: CharLowerHists[5]}, []string{"wowie"}},
		{"n=5, unusable prompt", args{prompt: "This text is everything before Hello world! This is a dude", n: 5, lower: false, words: true, hist: WordHists[5]}, []string{"This text is everything before", "Hello world! This is a dude"}},
		{"n=5, unusable prompt", args{prompt: "This text is everything before Hello world! This is a dude", n: 5, lower: true, words: true, hist: WordLowerHists[5]}, []string{"this text is everything before", "hello world! this is a dude"}},

		{"n=6, unusable prompt", args{prompt: "Some other totally unrelated characters: Hello world.", n: 6, lower: false, words: false, hist: CharHists[6]}, []string{"Some other totally unrelated characters: Hello ", "world."}},
		{"n=6, unusable prompt", args{prompt: "Hello!", n: 6, lower: true, words: false, hist: CharLowerHists[6]}, []string{"hello!"}},
		{"n=6, unusable prompt", args{prompt: "This text is everything before Hello world! This is a tax", n: 6, lower: false, words: true, hist: WordHists[6]}, []string{"This text is everything before", "Hello world! This is a tax"}},
		{"n=6, unusable prompt", args{prompt: "This text is everything before Hello world! This is a tax", n: 6, lower: true, words: true, hist: WordLowerHists[6]}, []string{"this text is everything before", "hello world! this is a tax"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetSeed(tt.args.prompt, tt.args.n, tt.args.lower, tt.args.words, tt.args.hist); reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSeed() = %v, didn't want %v", got, tt.want)
			}
		})
	}
}

// func TestBuildStringHistogram(t *testing.T) {
// 	type args struct {
// 		r         io.Reader
// 		n         int
// 		lowercase bool
// 		words     bool
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want StringHistogram
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := BuildStringHistogram(tt.args.r, tt.args.n, tt.args.lowercase, tt.args.words); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("BuildStringHistogram() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestGetSamplerFromStringHistogram(t *testing.T) {
// 	type args struct {
// 		hist StringHistogram
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want func(string) (string, error)
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := GetSamplerFromStringHistogram(tt.args.hist); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("GetSamplerFromStringHistogram() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestPrintStringHistogram(t *testing.T) {
// 	type args struct {
// 		hist StringHistogram
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			PrintStringHistogram(tt.args.hist)
// 		})
// 	}
// }

// func TestLoadOrCreateHistogram(t *testing.T) {
// 	type args struct {
// 		filename  string
// 		n         int
// 		lowercase bool
// 		words     bool
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    StringHistogram
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := LoadOrCreateHistogram(tt.args.filename, tt.args.n, tt.args.lowercase, tt.args.words)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("LoadOrCreateHistogram() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("LoadOrCreateHistogram() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// func TestCacheHistogram(t *testing.T) {
// 	type args struct {
// 		histogram StringHistogram
// 		filename  string
// 	}
// 	tests := []struct {
// 		name    string
// 		args    args
// 		wantErr bool
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if err := CacheHistogram(tt.args.histogram, tt.args.filename); (err != nil) != tt.wantErr {
// 				t.Errorf("CacheHistogram() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

// func Test_parseArgs(t *testing.T) {
// 	tests := []struct {
// 		name string
// 		want arguments
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := parseArgs(); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("parseArgs() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

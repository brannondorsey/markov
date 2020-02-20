## Markov

A simple Markov chain generator. Originally created for the Language Modeling lecture in the [Introduction to Synthetic Media](https://github.com/runwayml/Intro-Synthetic-Media) class at ITP/NYU.

### Download

Binary releases can be downloaded from the [releases page](https://github.com/brannondorsey/markov/releases/). Be sure to unzip your file, and put the `markov` binary somewhere in your `$PATH` (like inside `/usr/local/bin/`).

### Usage

```bash
# Download the UCI News Aggregator Dataset (400K news headlines) as a sample corpus
wget https://github.com/brannondorsey/markov/releases/download/v0.1.0/uci-news-aggregator-dataset.txt

# Build and cache an n-gram frequency histogram, then use it to generate text
markov --corpus uci-news-aggregator-dataset.txt --n-gram-length 3 --prompt "For the first time in a decade"
```

Below is a summary of the full usage of the `markov` command.

```
Usage of markov:
  -i, --corpus string       The input corpus to build the n-gram histogram with.
  -h, --help                Show this screen.
  -n, --n-gram-length int   The number of characters to use for each n-gram. (default 1)
  -p, --prompt string       The prompt to (optional). (default "hello")
```

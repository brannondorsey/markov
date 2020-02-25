# Changelog

## v0.3.0

* Make corpus file a required positional argument instead of a named flag
* Fix panic if corpus file was a directory
* Add `--words` flag to create word-level markov chains
* Change `--max-characters` argument to `--max`
* Allow sampling without a prompt (remove default "hello" prompt)
* Rename `cmd/markov/main.go` to `cmd/markov/markov.go`
* Add tests
* Add LICENSE
* Add Changelog

## v0.2.0

* Make `--n-gram-length` default to 3 instead of 1
* Add `--lowercase` flag
* Add README

## v0.1.0

Initial release

package main

type Applier interface {
	Compile(path string) (string, error)
}

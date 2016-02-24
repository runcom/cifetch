package main

import (
	"fmt"
	"os"
	"strings"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindDocker
	KindAppc
)

type Image interface {
	Kind() Kind
	GetLayers() ([]string, error)
}

func parseImage(img string) (Image, error) {
	switch {
	case strings.HasPrefix(img, dockerPrefix):
		return parseDockerImage(strings.TrimPrefix(img, dockerPrefix))
		//case strings.HasPrefix(img, appcPrefix):
		//
	}
	return nil, fmt.Errorf("no valid prefix provided")
}

func main() {
	img, err := parseImage(os.Args[1])
	if err != nil {
		panic(err)
	}
	_, err = img.GetLayers()
	if err != nil {
		panic(err)
	}
}

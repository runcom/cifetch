package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindDocker
	KindAppc
)

type Image interface {
	Kind() Kind
	GetLayers() error
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
		logrus.Fatal(err)
	}
	if err := img.GetLayers(); err != nil {
		logrus.Fatal(err)
	}
}

package main

import (
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

type Node any

// type Node struct {
// 	Label    string
// 	Children []Node
// }

type BehaviorNode struct {
	Type       string `yaml:"type"`
	Definition any    `yaml:"definition"`
}

type Behavior struct {
	EntityType string         `yaml:"entity_type"`
	Tree       []BehaviorNode `yaml:"tree"`
}

func main() {
	var t Node

	data, err := ioutil.ReadFile("ideal.yaml")
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(t)
}

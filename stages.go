package main

import "fmt"

func (a app) stageTest() {
	for msg := range a.unmarshalChan() {
		fmt.Println(msg)
	}
}

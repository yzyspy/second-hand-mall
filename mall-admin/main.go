package main

import "fmt"

type myError struct {
	string
}

type myError2 struct {
	string
}

func (e *myError) Error() string {
	return e.string
}

func DoBus() *myError {
	return nil
}

func main() {
	if err := DoBus(); err != nil {
		fmt.Println("xxxx")
	} else {
		fmt.Println("yyyy")
	}

}

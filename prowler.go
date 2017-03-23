package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
)

type config struct {
	Username string   `json:"username"`
	Repos    []string `json:"repos,omitempty"` // blank == all
	Token    string   `json:"token"`
	Services []string `json:"services,omitempty"` // blank == all
	Successs []string `json:"successStates,omitempty"`
	Pendings []string `json:"pendingStates,omitempty"`
	Failures []string `json:"failureStates,omitempty"`
	Conficts bool     `json:"hideMergeConflicts"`
	All      bool     `json:"showAllPrs"`
}

func check(err error, doing string) {
	if err != nil {
		fmt.Printf("Error %s: %s\n", doing, err)
		os.Exit(1)
	}
}

func main() {
	fmt.Println("Let's do this in GO!!!")
	usr, err := user.Current()
	check(err, "identifying user")
	data, err := ioutil.ReadFile(usr.HomeDir + "/.prowler.conf")
	check(err, "reading ~/.prowler.conf")
	var cfg config
	check(json.Unmarshal(data, &cfg), "unmarshaling json")
	fmt.Printf("%#v\n", cfg)
}

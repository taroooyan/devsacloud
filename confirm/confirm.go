package confirm

import (
	"fmt"
	"log"
)

// when call this function, until wait for user input data.
// input data expect yes or no
// when yes, return true
// when no, return false
func AskConfirm() bool {
	var res string
	_, err := fmt.Scanln(&res)
	if err != nil {
		log.Fatal(err)
	}

	yesResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	noResponses := []string{"n", "N", "no", "No", "NO"}
	if compareRes(yesResponses, res) {
		return true
	} else if compareRes(noResponses, res) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return AskConfirm()
	}
}

// whether res is included in expect.
// when yes, return true
// when no, return false
func compareRes(expect []string, res string) bool {
	for _, e := range expect {
		if e == res {
			return true
		}
	}
	return false
}

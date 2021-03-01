package db

import (
	"fmt"
	"testing"
)

func TestQueryContent(t *testing.T) {
	InitWithDbPath("/home/lq/ifb/src/eshop/system.db")

	abc := ContentAll("Game")
	for i, _ := range abc {
		fmt.Printf("\t%d", i)
	}
	fmt.Printf("\ntotal search result %d", len(abc))
	total, bcm := QueryContent("Game", "star tr", false)
	for _, item := range bcm {
		fmt.Println(string(item))
	}
	fmt.Printf("Total %d,Found %d\n", total, len(bcm))
	total, bcm = RegexContent("Game", "star ", true)
	for _, item := range bcm {
		fmt.Println(string(item))
	}
	fmt.Printf("Total %d,Found %d\n", total, len(bcm))
}

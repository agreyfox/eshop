package db

import (
	"fmt"
	"testing"

	_ "github.com/agreyfox/eshop/content"
	"github.com/agreyfox/eshop/system/search"
)

func TestQueryContentByBleve(t *testing.T) {
	InitWithDbPath("/home/lq/ifb/src/eshop/system.db")
	search.SetSearchDir("/home/lq/ifb/src/eshop/")
	InitSearchIndex()
	search.PrintSearchType()

	abc := ContentAll("Order")
	for i, _ := range abc {
		fmt.Printf("\t%s\n", abc[i])
	}
	fmt.Printf("\ntotal search result %d", len(abc))
	/* 	total, bcm := QueryContent("Order", "star tr", false)
	   	for _, item := range bcm {
	   		fmt.Println(string(item))
	   	}
	   	fmt.Printf("Total %d,Found %d\n", total, len(bcm))
	   	total, bcm = RegexContent("Game", "star ", true)
	   	for _, item := raitemtring(item))
	   	}
		   fmt.Printf("Total %d,Found %d\n", total, len(bcm)) */

	abccc, err := search.TypeMatchAll("Order", 1000, 0)
	fmt.Printf("total query is %d", len(abccc))
	fmt.Println(err)
	for item := range abccc {
		fmt.Printf("\t%s", abccc[item])
	}
}

func TestQueryContent(t *testing.T) {
	InitWithDbPath("/home/lq/ifb/src/eshop/system.db")
	bc, cs := Query("Order", QueryOptions{100, 0, "desc"})
	fmt.Println(bc)
	for i := range cs {
		fmt.Println(string(cs[i]))
	}
}

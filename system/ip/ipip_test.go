package ip

import (
	"testing"

	"github.com/agreyfox/eshop/system/db"
)

func TestLookupStandard(T *testing.T) {
	db.Init()
	Init()
	client := NewClient("", true)
	c, err := client.LookupStandard("193.22.152.235")
	if err == nil {
		T.Log(c)
	} else {
		T.Error(err)
	}

}

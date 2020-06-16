package ip

import "testing"

func TestLookupStandard(T *testing.T) {

	client := NewClient("", false)
	c, err := client.LookupStandard("193.22.152.235")
	if err == nil {
		T.Log(c)
	} else {
		T.Error(err)
	}

}

package codec

import (
	"bytes"
	"testing"
)

type BenchTestStruct struct {
	Name        string  `json:"name" constraints:"required=true;nillable=true;min-length=5"`
	Age         int     `json:"age" constraints:"required=true;nillable=true;min=21"`
	Description string  `json:"description" constraints:"required=true;nillable=true;max-length=50"`
	Cost        float64 `json:"cost" constraints:"required=true;nillable=true;exclusiveMin=200"`
	ItemCount   int     `json:"itemCount" constraints:"required=true;nillable=true;multipleOf=5"`
}

func BenchmarkJsonCodec(b *testing.B) {
	msg := BenchTestStruct{
		Name:        "BenchTest",
		Age:         25,
		Description: "this is bench testing",
		Cost:        299.9,
		ItemCount:   2000,
	}
	c, _ := Get("application/json", nil)
	buf := new(bytes.Buffer)

	for i := 0; i < b.N; i++ {
		if err := c.Write(msg, buf); err != nil {
			b.Errorf("error in write: %d", err)
		}
	}
}

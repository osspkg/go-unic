package ref_test

import (
	"testing"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/unic/internal/node"
	"go.osspkg.com/unic/internal/ref"
)

type (
	testType1 struct {
		A  int        `unic:"a"`
		B1 testType2  `unic:"b1"`
		B2 *testType2 `unic:"b2"`
		C  *string    `unic:"c"`
	}
	testType2 struct {
		D  string     `unic:"d"`
		E1 *testType3 `unic:"e1"`
	}
	testType3 struct {
		E string            `unic:"e"`
		F []string          `unic:"f"`
		G map[string]string `unic:"f"`
	}
)

func TestUnit_New(t *testing.T) {
	_, err := ref.New(nil)
	casecheck.Error(t, err)

	_, err = ref.New(testType1{})
	casecheck.Error(t, err)

	_, err = ref.New(&testType1{})
	casecheck.NoError(t, err)

	_, err = ref.New(0)
	casecheck.Error(t, err)
}

func TestUnit_Resolve(t *testing.T) {
	data := testType1{
		A: 1111,
		B1: testType2{
			D:  "2222",
			E1: nil,
		},
		B2: nil,
		C:  nil,
	}

	res, err := ref.New(&data)
	casecheck.NoError(t, err)

	b := node.NewBlock()
	casecheck.NoError(t, res.Build(b))
}

package dagRun

import (
	"fmt"
	"testing"
)

func TestDot(t *testing.T) {
	fmt.Println(g().DOT(
		WithCommonGraphAttr([]string{`label="testDot"`, `rankdir=LR`}),
		WithCommonNodeAttr([]string{`color="blue"`, `fontcolor="red"`}),
		WithCommonEdgeAttr([]string{`color="blue"`, `fontcolor="red"`}),
	),
	)
}

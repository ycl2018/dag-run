package dagRun

import (
	"fmt"
	"testing"
)

func TestDot(t *testing.T) {
	fmt.Println(g().DOT(
		WithCommonGraphAttr(`label="testDot"`, `rankdir=LR`),
		WithCommonNodeAttr(`color="blue"`, `fontcolor="red"`),
		WithCommonEdgeAttr(`color="blue"`, `fontcolor="red"`),
	),
	)
}

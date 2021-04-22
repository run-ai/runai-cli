package completion

import (
	"gotest.tools/assert"
	"testing"
)

func TestCompletionZsh(t *testing.T) {
	completionCommand := NewCompletionCmd()
	_, err := genZshCompletion(completionCommand)
	if err != nil {
		assert.Assert(t, false, "Expecting no error from zsh completion, received " + err.Error())
	}
}


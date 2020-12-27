package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"testing"
)

func TestLogin(t *testing.T) {
	err := authentication.Authenticate(&types.AuthenticationParams{})
	if err != nil {
		fmt.Println(err)
	}
}

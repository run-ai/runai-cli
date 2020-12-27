package login

import (
	"fmt"
	"github.com/run-ai/runai-cli/pkg/authentication"
	"github.com/run-ai/runai-cli/pkg/authentication/types"
	"testing"
)

func TestLogin(t *testing.T) {
	token, _ := authentication.Authenticate(&types.AuthenticationParams{})
	fmt.Println(token)
}

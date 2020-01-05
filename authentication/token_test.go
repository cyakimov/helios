package authentication

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestIssueJWTWithSecret(t *testing.T) {
	token, err := IssueJWTWithSecret("test", "test@test.com", time.Now().Add(1*time.Minute))
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}

func TestValidateJWTWithSecret(t *testing.T) {
	token, err := IssueJWTWithSecret("test", "test@test.com", time.Now().Add(1*time.Minute))

	assert.NoError(t, err)
	assert.True(t, ValidateJWTWithSecret("test", token))
	// ES256 Token (unsupported)
	es256 := "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.tyh-VfuzIxCyGYDlkBA7DfyjrqmSHu6pQ2hoZuFqUSLPNY2N0mpHb3nk5K17HWP_3cYHBw7AhHale5wky6-sVA"
	assert.False(t, ValidateJWTWithSecret("test", es256))
}

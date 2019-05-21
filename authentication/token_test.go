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
}

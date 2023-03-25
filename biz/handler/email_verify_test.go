package handler

import (
	emailverify "github.com/AfterShip/email-verifier"
	"testing"
)

func TestEmailVerify(t *testing.T) {
	verifier := emailverify.NewVerifier().EnableSMTPCheck().EnableAutoUpdateDisposable()
	ret, err := verifier.Verify("xxx@gmail.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

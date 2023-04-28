package handler

import (
	emailverify "github.com/AfterShip/email-verifier"
	"testing"
)

func TestEmailVerify(t *testing.T) {
	v := emailverify.NewVerifier().EnableSMTPCheck().EnableAutoUpdateDisposable()
	ret, err := v.Verify("xxx@gmail.com")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ret)
}

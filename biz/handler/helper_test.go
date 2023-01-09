package handler

import (
	"github.com/swordandtea/lets-habit-server/biz/config"
	"github.com/swordandtea/lets-habit-server/biz/dal"
	"testing"
)

func TestTokenGenerateExtract(t *testing.T) {
	config.GlobalConfig = &config.RuntimeConfig{
		JWT: config.JWTConfig{Cypher: "test_cypher"},
	}
	uid := dal.UID("1")
	tokenStr, err := GenerateUserToken(uid)
	if err != nil {
		t.Fatal(err)
	}

	uid2, err := ExtractUserToken(tokenStr)
	if err != nil {
		t.Fatal(err)
	}

	if uid2 != uid {
		t.Fatal("uid incorrect")
	}

	tokenStr, err = GeneratePollToken(uid)
	if err != nil {
		t.Fatal(err)
	}

	uid2, err = ExtractPollToken(tokenStr)
	if err != nil {
		t.Fatal(err)
	}

	if uid2 != uid {
		t.Fatal("uid incorrect")
	}
}

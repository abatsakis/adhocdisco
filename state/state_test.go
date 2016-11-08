package statedb

import (
	"math/rand"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func TestState(t *testing.T) {
	awsSess := session.New(&aws.Config{
		Region:      aws.String("us-west-2"),
		Credentials: credentials.NewSharedCredentials("", ""),
	})
	s := State{awsSess}

	label := aws.String("testDomain" + randomString(6))
	err := s.Create(label)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer s.Delete(label)

	key := aws.String("key1")
	b, err := s.PutExclusive(label, key, aws.String("dummy"))
	if err != nil || !*b {
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	key2 := aws.String("key2")
	b, err = s.PutExclusive(label, key2, aws.String("dummy2"))
	if err != nil || !*b {
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	b, err = s.Exists(label, key2, aws.String("dummy2"))
	if err != nil || !*b {
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	b, err = s.PutExclusive(label, key, aws.String("dummy3"))
	if err != nil || *b {
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	b, err = s.Put(label, key, aws.String("dummy3"))
	if err != nil || !*b {
		if err != nil {
			t.Error(err)
		}
		t.FailNow()
	}

	v, err := s.Get(label, key)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if aws.StringValue(v) != "dummy3" {
		t.Errorf("values mismatch %s", aws.StringValue(v))
		t.FailNow()
	}

	err = s.Delete(label)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	newKey := aws.String("key3")
	_, err = s.PutExclusive(label, newKey, aws.String("dummy4"))
	if err == nil {
		t.Errorf("expected error as the domain was deleted")
		t.FailNow()
	}
}

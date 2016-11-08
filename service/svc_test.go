package service

import (
	"testing"
)

func TestSvc(t *testing.T) {
	go Run()

	CreateTokenHTTP("http://localhost:3000/token")
}

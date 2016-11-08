//go:generate go get github.com/go-martini/martini
//go:generate go get github.com/martini-contrib/throttle
//go:generate go get github.com/martini-contrib/auth

package main

import (
	"github.com/adhocdisco/cache"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/throttle"
	"time"
)

const AuthToken = "avanti"

var m *martini.Martini

func init() {

	m = martini.New()

	// Setup middleware
	m.Use(martini.Recovery())
	m.Use(martini.Logger())
	m.Use(auth.Basic(AuthToken, ""))
	m.Use(throttle.Policy(&throttle.Quota{
		Limit:  5,
		Within: time.Second,
	}))

	// Setup routes
	r := martini.NewRouter()
	r.Put(`/token`, CreateToken)
	r.Get(`/:token`, GetToken)
	r.Put(`/:token/:id`, AddToken)

	m.Action(r.Handle)
}

func AddToken(params martini.Params) (int, string) {
	return cache.AddToken(params["token"], params["id"])
}

func CreateToken() (int, string) {
	return cache.CreateToken()
}

func GetToken(params martini.Params) (int, string) {
	return cache.GetToken(params["token"])
}

func main() {
	m.Run()
}

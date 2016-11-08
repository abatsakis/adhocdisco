//go:generate go get github.com/go-martini/martini
//go:generate go get github.com/martini-contrib/throttle
//go:generate go get github.com/martini-contrib/auth

package adhocservice

import (
	"errors"
	"github.com/adhocdisco/cache"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/throttle"
	"io/ioutil"
	"net/http"
	"strings"
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
		Limit:  10,
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

func Run() {
	m.Run()
}

func CreateTokenHTTP(url string) (*string, error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	request, err := http.NewRequest("PUT", url, strings.NewReader(""))
	request.ContentLength = 0
	request.SetBasicAuth(AuthToken, "")
	response, err := netClient.Do(request)

	if err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		if response.StatusCode > 299 {
			return nil, errors.New(response.Status)
		}

		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}

		s := string(contents)
		return &s, nil
	}
}

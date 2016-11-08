//go:generate go get github.com/patrickmn/go-cache

package cache

import (
	"github.com/adhocdisco/state"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	gocache "github.com/patrickmn/go-cache"
	"log"
	"math/rand"
	"sync"
	"time"
)

var cache map[string]*gocache.Cache
var state *statedb.State

var mx *sync.Mutex

var discovery string = "discovery"
var maxResults = 5

var defaultValidity time.Duration = 5 * time.Minute

func init() {
	rand.Seed(time.Now().UnixNano())

	cache = make(map[string]*gocache.Cache)
	mx = &sync.Mutex{}
	state = &statedb.State{
		session.New(&aws.Config{
			Region:      aws.String("us-west-1"),
			Credentials: credentials.NewSharedCredentials("", ""),
		}),
	}

	err := state.Create(&discovery)
	if err != nil {
		panic(err)
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GetToken(token string) (int, string) {
	b, err := state.Exists(&discovery, &token, &token)
	if err != nil {
		return 500, err.Error()
	}
	if !*b {
		return 404, ""
	}

	mx.Lock()
	defer mx.Unlock()

	val, ok := cache[token]
	if !ok {
		return 404, ""
	}

	vals := ""
	i := 0
	for _, v := range val.Items() {
		if v.Expired() {
			continue
		}

		str := v.Object.(string)
		vals = vals + str + " "

		i = i + 1
		if i >= maxResults {
			break
		}
	}
	if len(vals) > 0 {
		vals = vals[:len(vals)-1]
	}

	return 200, vals
}

func AddToken(token string, id string) (int, string) {
	log.Println("adding id", id, " to token", token)

	b, err := state.Exists(&discovery, &token, &token)
	if err != nil {
		return 500, err.Error()
	}
	if !*b {
		return 404, ""
	}

	mx.Lock()
	defer mx.Unlock()

	_, ok := cache[token]
	if !ok {
		cache[token] = gocache.New(defaultValidity, 1*time.Minute)
	}

	val := cache[token]
	val.Set(id, id, defaultValidity)

	return 200, ""
}

func CreateToken() (int, string) {
	token := randStringRunes(64)
	_, err := state.PutExclusive(&discovery, &token, &token)
	if err != nil {
		return 500, err.Error()
	}
	log.Println("created token ", token)

	return 200, token
}

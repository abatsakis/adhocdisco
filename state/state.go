//go:generate go get github.com/aws/...

package statedb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/simpledb"
)

type State struct {
	Sess *session.Session
}

// Creates a state context
func (s *State) Create(label *string) error {
	svc := simpledb.New(s.Sess)

	_, err := svc.CreateDomain(&simpledb.CreateDomainInput{
		DomainName: label,
	})

	return err
}

// Deletes a state context
func (s *State) Delete(label *string) error {
	svc := simpledb.New(s.Sess)

	_, err := svc.DeleteDomain(&simpledb.DeleteDomainInput{
		DomainName: label,
	})

	return err
}

// Gets the value of the specified key or nil on error
func (s *State) Get(label *string, key *string) (*string, error) {
	svc := simpledb.New(s.Sess)

	params := &simpledb.GetAttributesInput{
		ConsistentRead: aws.Bool(true),
		ItemName:       key,
		AttributeNames: []*string{key},
		DomainName:     label,
	}
	resp, err := svc.GetAttributes(params)

	if err != nil {
		return nil, err
	}

	if len(resp.Attributes) == 0 {
		return nil, nil
	}

	return resp.Attributes[0].Value, nil
}

// Checks whether the specified key/value exists
func (s *State) Exists(label *string, key *string, value *string) (*bool, error) {
	r, err := s.Get(label, key)
	return aws.Bool(aws.StringValue(r) == aws.StringValue(value)), err
}

// Attempts to store the key-value combination in the specified context
//
// if the key already exists returns false, else true
func (s *State) PutExclusive(label *string, key *string, value *string) (*bool, error) {
	return s.put(label, key, value, true)
}

// Attempts to store the key-value combination in the specified context
func (s *State) Put(label *string, key *string, value *string) (*bool, error) {
	return s.put(label, key, value, false)
}

func (s *State) put(label *string, key *string, value *string, exclusive bool) (*bool, error) {
	svc := simpledb.New(s.Sess)

	params := &simpledb.PutAttributesInput{
		Attributes: []*simpledb.ReplaceableAttribute{
			{
				Name:    key,
				Value:   value,
				Replace: aws.Bool(true),
			},
		},
		DomainName: label,
		ItemName:   key,
	}

	if exclusive {
		params.Expected = &simpledb.UpdateCondition{
			Exists: aws.Bool(false),
			Name:   key,
		}
	}

	_, err := svc.PutAttributes(params)

	if err == nil {
		return aws.Bool(true), nil
	}

	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "ConditionalCheckFailed" {
			return aws.Bool(false), nil
		}
	}

	return nil, err
}

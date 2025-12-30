// db/interfaces.go
package db

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// MongoSession is our custom interface for a MongoDB session.
// It includes all the methods we need for transactions.
type MongoSession interface {
	StartTransaction(opts ...options.Lister[options.TransactionOptions]) error
	CommitTransaction(context.Context) error
	AbortTransaction(context.Context) error
	EndSession(context.Context)
}

// SessionProvider is a custom interface for a MongoDB client that can start sessions.
type SessionProvider interface {
	StartSession(opts ...options.Lister[options.SessionOptions]) (MongoSession, error)
}

// This is a concrete implementation that wraps the mongo.Client.
type mongoSessionProvider struct {
	client *mongo.Client
}

// NewSessionProvider returns a new session provider for a given client.
func NewSessionProvider(client *mongo.Client) SessionProvider {
	return &mongoSessionProvider{client: client}
}

// The wrapper for the mongo.Session that implements our interface.
type mongoSessionWrapper struct {
	mongo.Session
}

// Wrap the mongo.Client.StartSession method.
func (p *mongoSessionProvider) StartSession(opts ...options.Lister[options.SessionOptions]) (MongoSession, error) {
	session, err := p.client.StartSession(opts...)
	if err != nil {
		return nil, err
	}
	return &mongoSessionWrapper{Session: *session}, nil
}

// Implement the MongoSession interface methods.
func (s *mongoSessionWrapper) StartTransaction(opts ...options.Lister[options.TransactionOptions]) error {
	return s.Session.StartTransaction(opts...)
}

func (s *mongoSessionWrapper) CommitTransaction(ctx context.Context) error {
	return s.Session.CommitTransaction(ctx)
}

func (s *mongoSessionWrapper) AbortTransaction(ctx context.Context) error {
	return s.Session.AbortTransaction(ctx)
}

func (s *mongoSessionWrapper) EndSession(ctx context.Context) {
	s.Session.EndSession(ctx)
}

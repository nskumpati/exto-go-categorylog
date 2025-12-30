package db

import (
	"context"
	"errors"
	"time"

	"github.com/gaeaglobal/exto/server/model"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CacheUser struct {
	Email string
	User  *model.User
	Org   *model.Organization
	TTL   time.Time
}

var cacheUserStore = make(map[string]*CacheUser)

type AppDB struct {
	Client *mongo.Client
}

func NewMockAppDB() *AppDB {
	return &AppDB{}
}

func (d *AppDB) GetCoreDatabase() *mongo.Database {
	// Assuming "core" is the name of your core database.
	return d.Client.Database("sc_core")
}

func (d *AppDB) GetOrgDatabase(orgName string) *mongo.Database {
	// Assuming "organization" is the name of your organization database.
	dbName := "sc_" + orgName
	return d.Client.Database(dbName)
}

func (d *AppDB) GetUserByEmail(email string) (*CacheUser, error) {

	if cached, ok := cacheUserStore[email]; ok {
		if time.Now().Before(cached.TTL) {
			return cached, nil
		}
	}

	// Get current organization id from identity
	identitiesCol := d.GetCoreDatabase().Collection("identities")
	var identity model.Identity
	identityOpts := options.FindOne().SetProjection(bson.M{"_id": 1, "email": 1, "is_active": 1, "current_org_id": 1})
	err := identitiesCol.FindOne(context.Background(), bson.M{"email": email}, identityOpts).Decode(&identity)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// load users from org
	userOrgCol := d.GetCoreDatabase().Collection("users")
	var user model.User
	err = userOrgCol.FindOne(context.Background(), bson.M{"email": email, "org_id": identity.CurrentOrgID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user don't belong to any organization")
	}
	if !user.IsActive {
		return nil, errors.New("user is not active")
	}

	// Load organization name
	orgCol := d.GetCoreDatabase().Collection("organizations")
	var organization model.Organization
	orgOpts := options.FindOne().SetProjection(bson.M{"_id": 1, "name": 1, "slug": 1, "is_active": 1})
	err = orgCol.FindOne(context.Background(), bson.M{"_id": identity.CurrentOrgID}, orgOpts).Decode(&organization)
	if err != nil {
		return nil, errors.New("organization not found")
	}
	if !organization.IsActive {
		return nil, errors.New("organization is not active")
	}

	usr := CacheUser{
		Email: email,
		User:  &user,
		Org:   &organization,
		TTL:   time.Now().Add(5 * time.Minute), // Cache for 5 minutes
	}

	cacheUserStore[email] = &usr

	return cacheUserStore[email], nil

}

func ConnectToDatabase(database_url string) (*AppDB, error) {
	// Implement your database connection logic here.
	clientOptions := options.Client().ApplyURI(database_url)
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(context.Background(), nil); err != nil {
		return nil, err
	}
	return &AppDB{Client: client}, nil
}

func DisconnectDatabase(db *AppDB) error {
	// Implement your database disconnection logic here.
	if db.Client != nil {
		return db.Client.Disconnect(context.Background())
	}
	return nil
}

func GetDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func DeleteCacheUser(email string) {
	delete(cacheUserStore, email)
}

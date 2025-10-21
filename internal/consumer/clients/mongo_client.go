package clients

import (
	"context"

	"github.com/therehabstreet/podoai/internal/consumer/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient interface {
	FetchUserByID(ctx context.Context, userID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
}

// MongoDBClient implements DBClient
type MongoDBClient struct {
	Client *mongo.Client
}

const (
	DatabaseName = "podoai"
)

var ConsumerMongoClient *mongo.Client

func InitConsumerMongoClient(uri string) (*MongoDBClient, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ConsumerMongoClient = client
	mongoClient := &MongoDBClient{Client: client}

	// Map of collection name to slice of index definitions
	indexMap := map[string][]mongo.IndexModel{
		"users": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "email", Value: 1}},
				Options: options.Index().SetName("idx_email").SetUnique(true),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "phone", Value: 1}},
				Options: options.Index().SetName("idx_phone").SetUnique(true),
			},
		},
	}

	db := mongoClient.Client.Database(DatabaseName)
	for collName, indexes := range indexMap {
		coll := db.Collection(collName)
		if _, err := coll.Indexes().CreateMany(ctx, indexes); err != nil {
			return nil, err
		}
	}

	return mongoClient, nil
}

func (m *MongoDBClient) FetchUserByID(ctx context.Context, userID string) (*models.User, error) {
	coll := m.Client.Database(DatabaseName).Collection("users")
	var user models.User
	err := coll.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	return &user, err
}

func (m *MongoDBClient) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	coll := m.Client.Database(DatabaseName).Collection("users")
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := coll.UpdateOne(ctx, filter, update)
	return user, err
}

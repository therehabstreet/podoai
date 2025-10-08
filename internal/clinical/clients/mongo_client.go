package clients

import (
	"context"
	"time"

	"github.com/therehabstreet/podoai/internal/clinical/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient interface {
	// Clinic methods
	FetchClinicByID(ctx context.Context, id string) (*models.Clinic, error)
	CreateClinic(ctx context.Context, clinic models.Clinic) (*models.Clinic, error)
	UpdateClinic(ctx context.Context, clinic models.Clinic) (*models.Clinic, error)
	// ClinicUser methods
	CreateClinicUser(ctx context.Context, user models.ClinicUser) (*models.ClinicUser, error)
	FetchClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) (*models.ClinicUser, error)
	FetchClinicUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.ClinicUser, error)
	UpdateClinicUser(ctx context.Context, user models.ClinicUser) (*models.ClinicUser, error)
	DeleteClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) error
	ListClinicUsers(ctx context.Context, clinicID string, page, pageSize int32) ([]*models.ClinicUser, int64, error)
}

// MongoDBClient implements DBClient
type MongoDBClient struct {
	Client *mongo.Client
}

var ClinicalMongoClient *mongo.Client

func InitClinicalMongoClient(uri string) (*MongoDBClient, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ClinicalMongoClient = client
	mongoClient := &MongoDBClient{Client: client}

	// Map of collection name to slice of index definitions
	indexMap := map[string][]mongo.IndexModel{
		"clinical_patients": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "clinic_id", Value: 1}},
				Options: options.Index().SetName("idx_clinic_id"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "_id", Value: 1}},
				Options: options.Index().SetName("idx_patient_id"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "clinic_id", Value: 1}, {Key: "last_scan_date", Value: -1}},
				Options: options.Index().SetName("idx_clinic_lastscandate"),
			},
		},
		"clinic": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "id", Value: 1}},
				Options: options.Index().SetName("idx_clinic_id"),
			},
		},
		"clinical_users": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "_id", Value: 1}},
				Options: options.Index().SetName("idx_user_id"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "phone_number", Value: 1}},
				Options: options.Index().SetName("idx_phone_number"),
			},
		},
	}

	db := mongoClient.Client.Database("podoai")
	for collName, indexes := range indexMap {
		coll := db.Collection(collName)
		if _, err := coll.Indexes().CreateMany(ctx, indexes); err != nil {
			return nil, err
		}
	}

	return mongoClient, nil
}

// Clinic methods
func (m *MongoDBClient) FetchClinicByID(ctx context.Context, id string) (*models.Clinic, error) {
	coll := m.Client.Database("podoai").Collection("clinics")
	var clinic models.Clinic
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&clinic)
	return &clinic, err
}

func (m *MongoDBClient) CreateClinic(ctx context.Context, clinic models.Clinic) (*models.Clinic, error) {
	coll := m.Client.Database("podoai").Collection("clinics")
	now := time.Now()

	// Set created_at if not already set
	if clinic.CreatedAt.IsZero() {
		clinic.CreatedAt = now
	}
	// Set updated_at if not already set
	if clinic.UpdatedAt.IsZero() {
		clinic.UpdatedAt = now
	}
	_, err := coll.InsertOne(ctx, clinic)
	return &clinic, err
}

func (m *MongoDBClient) UpdateClinic(ctx context.Context, clinic models.Clinic) (*models.Clinic, error) {
	coll := m.Client.Database("podoai").Collection("clinics")
	filter := bson.M{"_id": clinic.ID}
	update := bson.M{"$set": bson.M{
		"name":    clinic.Name,
		"address": clinic.Address,
	}}
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}
	// Fetch the updated clinic to get the actual field values
	var updatedClinic models.Clinic
	err = coll.FindOne(ctx, filter).Decode(&updatedClinic)
	if err != nil {
		return nil, err
	}
	return &updatedClinic, nil
}

// ClinicUser methods
func (m *MongoDBClient) CreateClinicUser(ctx context.Context, user models.ClinicUser) (*models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	_, err := coll.InsertOne(ctx, user)
	return &user, err
}

func (m *MongoDBClient) FetchClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) (*models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	var user models.ClinicUser
	filter := bson.M{"_id": userID, "clinic_id": clinicID}
	err := coll.FindOne(ctx, filter).Decode(&user)
	return &user, err
}

func (m *MongoDBClient) FetchClinicUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	var user models.ClinicUser
	err := coll.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&user)
	return &user, err
}

func (m *MongoDBClient) UpdateClinicUser(ctx context.Context, user models.ClinicUser) (*models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := coll.UpdateOne(ctx, filter, update)
	return &user, err
}

func (m *MongoDBClient) DeleteClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) error {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	filter := bson.M{"_id": userID, "clinic_id": clinicID}
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

func (m *MongoDBClient) ListClinicUsers(ctx context.Context, clinicID string, page, pageSize int32) ([]*models.ClinicUser, int64, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	filter := bson.M{"clinic_id": clinicID}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	findOptions := options.Find().SetSkip(skip).SetLimit(limit)
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var users []*models.ClinicUser
	for cursor.Next(ctx) {
		var user models.ClinicUser
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, &user)
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		total = int64(len(users))
	}
	return users, total, nil
}

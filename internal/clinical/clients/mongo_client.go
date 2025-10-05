package clients

import (
	"context"
	"time"

	"github.com/therehabstreet/podoai/internal/clinical/models"
	commonModels "github.com/therehabstreet/podoai/internal/common/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient interface {
	FetchPatients(ctx context.Context, page, pageSize int32, sortBy, sortOrder string) ([]commonModels.Patient, int64, error)
	FetchPatientByID(ctx context.Context, id primitive.ObjectID) (commonModels.Patient, error)
	CreatePatient(ctx context.Context, patient commonModels.Patient) (commonModels.Patient, error)
	DeletePatientByID(ctx context.Context, id primitive.ObjectID) error
	// Clinic methods
	FetchClinicByID(ctx context.Context, id string) (models.Clinic, error)
	CreateClinic(ctx context.Context, clinic models.Clinic) (models.Clinic, error)
	UpdateClinic(ctx context.Context, clinic models.Clinic) (models.Clinic, error)
	// ClinicUser methods
	CreateClinicUser(ctx context.Context, user models.ClinicUser) (models.ClinicUser, error)
	FetchClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) (models.ClinicUser, error)
	FetchClinicUserByPhoneNumber(ctx context.Context, phoneNumber string) (models.ClinicUser, error)
	UpdateClinicUser(ctx context.Context, user models.ClinicUser) (models.ClinicUser, error)
	DeleteClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) error
	ListClinicUsers(ctx context.Context, clinicID string, page, pageSize int32) ([]models.ClinicUser, int64, error)
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

func (m *MongoDBClient) FetchPatients(ctx context.Context, page, pageSize int32, sortBy, sortOrder string) ([]commonModels.Patient, int64, error) {
	collection := m.Client.Database("podoai").Collection("clinical_patients")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	order := 1
	if sortOrder == "desc" {
		order = -1
	}
	sort := bson.D{}
	if sortBy != "" {
		sort = bson.D{{Key: sortBy, Value: order}}
	}

	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)

	cursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var patients []commonModels.Patient
	for cursor.Next(ctx) {
		var patient commonModels.Patient
		if err := cursor.Decode(&patient); err != nil {
			continue
		}
		patients = append(patients, patient)
	}

	totalCount, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		totalCount = int64(len(patients))
	}

	return patients, totalCount, nil
}

func (m *MongoDBClient) FetchPatientByID(ctx context.Context, id primitive.ObjectID) (commonModels.Patient, error) {
	collection := m.Client.Database("podoai").Collection("clinical_patients")
	var patient commonModels.Patient
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&patient)
	return patient, err
}

func (m *MongoDBClient) CreatePatient(ctx context.Context, patient commonModels.Patient) (commonModels.Patient, error) {
	collection := m.Client.Database("podoai").Collection("clinical_patients")
	res, err := collection.InsertOne(ctx, patient)
	if err != nil {
		return patient, err
	}
	oid, ok := res.InsertedID.(primitive.ObjectID)
	if ok {
		patient.ID = oid
	}
	return patient, nil
}

func (m *MongoDBClient) DeletePatientByID(ctx context.Context, id primitive.ObjectID) error {
	collection := m.Client.Database("podoai").Collection("clinical_patients")
	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Clinic methods
func (m *MongoDBClient) FetchClinicByID(ctx context.Context, id string) (models.Clinic, error) {
	coll := m.Client.Database("podoai").Collection("clinics")
	var clinic models.Clinic
	err := coll.FindOne(ctx, bson.M{"_id": id}).Decode(&clinic)
	return clinic, err
}

func (m *MongoDBClient) CreateClinic(ctx context.Context, clinic models.Clinic) (models.Clinic, error) {
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
	return clinic, err
}

func (m *MongoDBClient) UpdateClinic(ctx context.Context, clinic models.Clinic) (models.Clinic, error) {
	coll := m.Client.Database("podoai").Collection("clinics")
	filter := bson.M{"_id": clinic.ID}
	update := bson.M{"$set": bson.M{
		"name":    clinic.Name,
		"address": clinic.Address,
	}}
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return clinic, err
	}
	// Fetch the updated clinic to get the actual field values
	var updatedClinic models.Clinic
	err = coll.FindOne(ctx, filter).Decode(&updatedClinic)
	if err != nil {
		return clinic, err
	}
	return updatedClinic, nil
}

// ClinicUser methods
func (m *MongoDBClient) CreateClinicUser(ctx context.Context, user models.ClinicUser) (models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	_, err := coll.InsertOne(ctx, user)
	return user, err
}

func (m *MongoDBClient) FetchClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) (models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	var user models.ClinicUser
	filter := bson.M{"_id": userID, "clinic_id": clinicID}
	err := coll.FindOne(ctx, filter).Decode(&user)
	return user, err
}

func (m *MongoDBClient) FetchClinicUserByPhoneNumber(ctx context.Context, phoneNumber string) (models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	var user models.ClinicUser
	err := coll.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&user)
	return user, err
}

func (m *MongoDBClient) UpdateClinicUser(ctx context.Context, user models.ClinicUser) (models.ClinicUser, error) {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	_, err := coll.UpdateOne(ctx, filter, update)
	return user, err
}

func (m *MongoDBClient) DeleteClinicUserByIDAndClinic(ctx context.Context, userID, clinicID string) error {
	coll := m.Client.Database("podoai").Collection("clinical_users")
	filter := bson.M{"_id": userID, "clinic_id": clinicID}
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

func (m *MongoDBClient) ListClinicUsers(ctx context.Context, clinicID string, page, pageSize int32) ([]models.ClinicUser, int64, error) {
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
	var users []models.ClinicUser
	for cursor.Next(ctx) {
		var user models.ClinicUser
		if err := cursor.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		total = int64(len(users))
	}
	return users, total, nil
}

package clients

import (
	"context"
	"fmt"
	"strings"
	"time"

	clinicalModels "github.com/therehabstreet/podoai/internal/clinical/models"
	"github.com/therehabstreet/podoai/internal/common/helpers"
	"github.com/therehabstreet/podoai/internal/common/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient interface {
	FetchScans(ctx context.Context, patientID string, ownerEntityID string, page, pageSize int32, sortBy, sortOrder string) ([]*models.Scan, int64, error)
	FetchScanByID(ctx context.Context, scanID string, ownerEntityID string) (*models.Scan, error)
	CreateScan(ctx context.Context, scan *models.Scan) (*models.Scan, error)
	UpdateScan(ctx context.Context, scan *models.Scan) (*models.Scan, error)
	DeleteScanByID(ctx context.Context, scanID string, ownerEntityID string) error
	// Product methods
	FetchProductByID(ctx context.Context, productID string) (*models.Product, error)
	// Exercise-related methods
	FetchExerciseByID(ctx context.Context, exerciseID string) (*models.Exercise, error)
	// Therapy-related methods
	FetchTherapyByID(ctx context.Context, therapyID string) (*models.Therapy, error)
	// OTP methods
	StoreOTP(ctx context.Context, otp *models.OTP) error
	GetOTPByPhoneNumber(ctx context.Context, phoneNumber string) (*models.OTP, error)
	MarkOTPAsUsed(ctx context.Context, otpID string) error
	IncrementOTPAttempts(ctx context.Context, otpID string) error
	CleanupExpiredOTPs(ctx context.Context) error
	// User existence methods
	ClinicalUserExists(ctx context.Context, phoneNumber string) (bool, error)
	// User management methods
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (interface{}, error)
	CreateUser(ctx context.Context, user interface{}) (string, error)
	// Patient methods
	FetchPatientByID(ctx context.Context, patientID string, ownerEntityID string) (*models.Patient, error)
	FetchPatients(ctx context.Context, ownerEntityID string, page, pageSize int32, sortBy, sortOrder string) ([]*models.Patient, int64, error)
	SearchPatients(ctx context.Context, searchTerm, ownerEntityID string, page, pageSize int32) ([]*models.Patient, int64, error)
	CreatePatient(ctx context.Context, patient *models.Patient) (*models.Patient, error)
	UpdatePatient(ctx context.Context, patient *models.Patient) (*models.Patient, error)
	DeletePatientByID(ctx context.Context, patientID string, ownerEntityID string) error
}

// MongoDBClient implements DBClient
type MongoDBClient struct {
	Client *mongo.Client
}

const (
	DatabaseName = "podoai"
)

func InitCommonMongoClient(uri string) (*MongoDBClient, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	mongoClient := &MongoDBClient{Client: client}

	// Map of collection name to slice of index definitions
	indexMap := map[string][]mongo.IndexModel{
		"clinical_scans": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "owner_entity_id", Value: 1}, {Key: "patient_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_owner_patient_created_desc"),
			},
		},
		"consumer_scans": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "patient_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_patient_created_desc"),
			},
		},
		"clinical_otps": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "phone_number", Value: 1}},
				Options: options.Index().SetName("idx_phone_number"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_expires_at"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_created_desc"),
			},
			mongo.IndexModel{
				Keys: bson.D{
					{Key: "phone_number", Value: 1},
					{Key: "is_used", Value: 1},
					{Key: "expires_at", Value: 1},
				},
				Options: options.Index().SetName("idx_phone_used_expires"),
			},
		},
		"consumer_otps": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "phone_number", Value: 1}},
				Options: options.Index().SetName("idx_phone_number"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "expires_at", Value: 1}},
				Options: options.Index().SetName("idx_expires_at"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_created_desc"),
			},
			mongo.IndexModel{
				Keys: bson.D{
					{Key: "phone_number", Value: 1},
					{Key: "is_used", Value: 1},
					{Key: "expires_at", Value: 1},
				},
				Options: options.Index().SetName("idx_phone_used_expires"),
			},
		},
		"clinical_patients": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "owner_entity_id", Value: 1}, {Key: "_id", Value: 1}},
				Options: options.Index().SetName("idx_owner_id"),
			},
			mongo.IndexModel{
				Keys:    bson.D{{Key: "owner_entity_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_owner_created_desc"),
			},
		},
		"consumer_patients": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "owner_entity_id", Value: 1}, {Key: "_id", Value: 1}},
				Options: options.Index().SetName("idx_owner_id"),
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

func (m *MongoDBClient) FetchScans(ctx context.Context, patientID string, ownerEntityID string, page, pageSize int32, sortBy, sortOrder string) ([]*models.Scan, int64, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))

	filter := bson.M{"owner_entity_id": ownerEntityID}
	validPatientID := patientID != "" && len(strings.TrimSpace(patientID)) > 0
	if validPatientID {
		filter["patient_id"] = patientID
	}

	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// Default sort by created_at desc for date-based ordering
	sort := bson.D{{Key: "created_at", Value: -1}}
	if sortBy != "" {
		order := 1
		if sortOrder == "desc" {
			order = -1
		}
		sort = bson.D{{Key: sortBy, Value: order}}
	}

	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)
	var scans []*models.Scan
	for cursor.Next(ctx) {
		var scan models.Scan
		if err := cursor.Decode(&scan); err != nil {
			continue
		}
		scans = append(scans, &scan)
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		total = int64(len(scans))
	}
	return scans, total, nil
}

func (m *MongoDBClient) FetchScanByID(ctx context.Context, scanID string, ownerEntityID string) (*models.Scan, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	filter := bson.M{"_id": scanID, "owner_entity_id": ownerEntityID}
	var scan models.Scan
	err := coll.FindOne(ctx, filter).Decode(&scan)
	return &scan, err
}

func (m *MongoDBClient) CreateScan(ctx context.Context, scan *models.Scan) (*models.Scan, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	_, err := coll.InsertOne(ctx, scan)
	return scan, err
}

func (m *MongoDBClient) UpdateScan(ctx context.Context, scan *models.Scan) (*models.Scan, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	filter := bson.M{"_id": scan.ID, "owner_entity_id": scan.OwnerEntityID}
	update := bson.M{"$set": scan}
	_, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	// Fetch the updated scan to return the latest state
	var updatedScan models.Scan
	err = coll.FindOne(ctx, filter).Decode(&updatedScan)
	if err != nil {
		return nil, err
	}
	return &updatedScan, nil
}

func (m *MongoDBClient) DeleteScanByID(ctx context.Context, scanID string, ownerEntityID string) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	filter := bson.M{"_id": scanID, "owner_entity_id": ownerEntityID}
	_, err := coll.DeleteOne(ctx, filter)
	return err
}

// Product methods
func (m *MongoDBClient) FetchProductByID(ctx context.Context, productID string) (*models.Product, error) {
	coll := m.Client.Database(DatabaseName).Collection("products")
	var product models.Product
	err := coll.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	return &product, err
}

// Exercise methods
func (m *MongoDBClient) FetchExerciseByID(ctx context.Context, exerciseID string) (*models.Exercise, error) {
	coll := m.Client.Database(DatabaseName).Collection("exercises")
	var exercise models.Exercise
	err := coll.FindOne(ctx, bson.M{"_id": exerciseID}).Decode(&exercise)
	return &exercise, err
}

// Therapy methods
func (m *MongoDBClient) FetchTherapyByID(ctx context.Context, therapyID string) (*models.Therapy, error) {
	coll := m.Client.Database(DatabaseName).Collection("therapies")
	var therapy models.Therapy
	err := coll.FindOne(ctx, bson.M{"_id": therapyID}).Decode(&therapy)
	return &therapy, err
}

// OTP methods implementation

// StoreOTP stores a new OTP in the database
func (m *MongoDBClient) StoreOTP(ctx context.Context, otp *models.OTP) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "otps"))

	// First, invalidate any existing OTPs for this phone number
	_, err := coll.UpdateMany(ctx,
		bson.M{"phone_number": otp.PhoneNumber, "is_used": false},
		bson.M{"$set": bson.M{"is_used": true}})
	if err != nil {
		// Log error but continue with storing new OTP
	}

	// Store the new OTP
	_, err = coll.InsertOne(ctx, otp)
	return err
}

// GetOTPByPhoneNumber retrieves the latest valid OTP for a phone number
func (m *MongoDBClient) GetOTPByPhoneNumber(ctx context.Context, phoneNumber string) (*models.OTP, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "otps"))

	// Find the latest unused, non-expired OTP for this phone number
	filter := bson.M{
		"phone_number": phoneNumber,
		"is_used":      false,
		"expires_at":   bson.M{"$gt": time.Now()},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var otp models.OTP
	err := coll.FindOne(ctx, filter, opts).Decode(&otp)
	if err != nil {
		return nil, err
	}

	return &otp, nil
}

// MarkOTPAsUsed marks an OTP as used
func (m *MongoDBClient) MarkOTPAsUsed(ctx context.Context, otpID string) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "otps"))

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": otpID},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	return err
}

// IncrementOTPAttempts increments the attempt count for an OTP
func (m *MongoDBClient) IncrementOTPAttempts(ctx context.Context, otpID string) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "otps"))

	_, err := coll.UpdateOne(ctx,
		bson.M{"_id": otpID},
		bson.M{"$inc": bson.M{"attempts": 1}})

	return err
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (m *MongoDBClient) CleanupExpiredOTPs(ctx context.Context) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "otps"))

	// Delete OTPs that are either expired or used and older than 24 hours
	cutoffTime := time.Now().Add(-24 * time.Hour)
	filter := bson.M{
		"$or": []bson.M{
			{"expires_at": bson.M{"$lt": time.Now()}},
			{"is_used": true, "created_at": bson.M{"$lt": cutoffTime}},
		},
	}

	_, err := coll.DeleteMany(ctx, filter)
	return err
}

// ClinicalUserExists checks if a clinical user exists by phone number
func (m *MongoDBClient) ClinicalUserExists(ctx context.Context, phoneNumber string) (bool, error) {
	clinicalColl := m.Client.Database(DatabaseName).Collection("clinical_users")
	count, err := clinicalColl.CountDocuments(ctx, bson.M{"phone_number": phoneNumber})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserByPhoneNumber fetches user by phone number from appropriate collection
func (m *MongoDBClient) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (interface{}, error) {
	if helpers.IsClinicalApp(ctx) {
		clinicalColl := m.Client.Database(DatabaseName).Collection("clinical_users")
		var clinicalUser clinicalModels.ClinicUser
		err := clinicalColl.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&clinicalUser)
		return clinicalUser, err
	}

	consumerColl := m.Client.Database(DatabaseName).Collection("consumer_users")
	var consumerUser models.User
	err := consumerColl.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&consumerUser)
	return consumerUser, err
}

// CreateUser creates a user in the appropriate collection
func (m *MongoDBClient) CreateUser(ctx context.Context, user interface{}) (string, error) {
	if helpers.IsClinicalApp(ctx) {
		clinicalColl := m.Client.Database(DatabaseName).Collection("clinical_users")
		clinicalUser := user.(*clinicalModels.ClinicUser)
		_, err := clinicalColl.InsertOne(ctx, clinicalUser)
		return clinicalUser.ID, err
	}

	consumerColl := m.Client.Database(DatabaseName).Collection("consumer_users")
	consumerUser := user.(*models.User)
	_, err := consumerColl.InsertOne(ctx, consumerUser)
	return consumerUser.ID, err
}

// FetchPatientByID retrieves a patient by ID with owner entity validation
func (m *MongoDBClient) FetchPatientByID(ctx context.Context, patientID string, ownerEntityID string) (*models.Patient, error) {
	var patient models.Patient

	// Build filter with both patient ID and owner entity ID for security
	filter := bson.M{
		"_id":             patientID,
		"owner_entity_id": ownerEntityID,
	}

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Find the patient
	err := coll.FindOne(ctx, filter).Decode(&patient)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("patient not found")
		}
		return nil, fmt.Errorf("error fetching patient")
	}

	return &patient, nil
}

// FetchPatients retrieves patients with pagination and filtering
func (m *MongoDBClient) FetchPatients(ctx context.Context, ownerEntityID string, page, pageSize int32, sortBy, sortOrder string) ([]*models.Patient, int64, error) {
	var patients []*models.Patient

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Build filter
	filter := bson.M{"owner_entity_id": ownerEntityID}

	// Set up pagination
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)

	// Default sort by created_at desc
	sort := bson.D{{Key: "created_at", Value: -1}}
	if sortBy != "" {
		order := 1
		if sortOrder == "desc" {
			order = -1
		}
		sort = bson.D{{Key: sortBy, Value: order}}
	}

	// Get total count
	totalCount, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return patients, 0, fmt.Errorf("error counting patients")
	}

	// Find patients with pagination
	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
	cursor, err := coll.Find(ctx, filter, findOptions)
	if err != nil {
		return patients, 0, fmt.Errorf("error fetching patients")
	}
	defer cursor.Close(ctx)

	// Decode results
	for cursor.Next(ctx) {
		var patient models.Patient
		if err := cursor.Decode(&patient); err != nil {
			continue
		}
		patients = append(patients, &patient)
	}

	return patients, totalCount, nil
}

// SearchPatients searches patients by name or phone number
func (m *MongoDBClient) SearchPatients(ctx context.Context, searchTerm, ownerEntityID string, page, pageSize int32) ([]*models.Patient, int64, error) {
	var patients []*models.Patient
	if searchTerm == "" || len(strings.TrimSpace(searchTerm)) == 0 {
		return patients, 0, fmt.Errorf("search term cannot be empty")
	}

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Build search filter with regex for name and exact match for phone
	searchFilter := bson.M{
		"owner_entity_id": ownerEntityID,
		"$or": []bson.M{
			{"name": bson.M{"$regex": searchTerm, "$options": "i"}},
			{"phone_number": searchTerm},
		},
	}

	// Set up pagination
	skip := int64((page - 1) * pageSize)
	limit := int64(pageSize)
	sort := bson.D{{Key: "created_at", Value: -1}}

	// Get total count
	totalCount, err := coll.CountDocuments(ctx, searchFilter)
	if err != nil {
		return patients, 0, fmt.Errorf("error counting search results")
	}

	// Find patients with pagination
	findOptions := options.Find().SetSkip(skip).SetLimit(limit).SetSort(sort)
	cursor, err := coll.Find(ctx, searchFilter, findOptions)
	if err != nil {
		return patients, 0, fmt.Errorf("error searching patients")
	}
	defer cursor.Close(ctx)

	// Decode results
	for cursor.Next(ctx) {
		var patient models.Patient
		if err := cursor.Decode(&patient); err != nil {
			continue
		}
		patients = append(patients, &patient)
	}

	return patients, totalCount, nil
}

// CreatePatient creates a new patient
func (m *MongoDBClient) CreatePatient(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if patient.Name == "" || len(strings.TrimSpace(patient.Name)) == 0 {
		return nil, fmt.Errorf("patient name cannot be empty")
	}

	// Set created time
	patient.CreatedAt = time.Now()

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Insert patient
	_, err := coll.InsertOne(ctx, patient)
	if err != nil {
		return nil, fmt.Errorf("error creating patient")
	}

	return patient, nil
}

// UpdatePatient updates an existing patient
func (m *MongoDBClient) UpdatePatient(ctx context.Context, patient *models.Patient) (*models.Patient, error) {
	if patient.ID == "" {
		return nil, fmt.Errorf("patient ID is required")
	}
	if patient.OwnerEntityID == "" {
		return nil, fmt.Errorf("owner entity ID is required")
	}
	if patient.Name == "" || len(strings.TrimSpace(patient.Name)) == 0 {
		return nil, fmt.Errorf("patient name cannot be empty")
	}

	filter := bson.M{
		"_id":             patient.ID,
		"owner_entity_id": patient.OwnerEntityID,
	}

	update := bson.M{
		"$set": bson.M{
			"name":           patient.Name,
			"phone_number":   patient.PhoneNumber,
			"age":            patient.Age,
			"gender":         patient.Gender,
			"foot_size":      patient.FootSize,
			"total_scans":    patient.TotalScans,
			"last_scan_date": patient.LastScanDate,
		},
	}

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Update the patient
	result, err := coll.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, fmt.Errorf("error updating patient: %v", err)
	}

	if result.MatchedCount == 0 {
		return nil, fmt.Errorf("patient not found")
	}

	return patient, nil
}

// DeletePatientByID deletes a patient by ID with owner entity validation
func (m *MongoDBClient) DeletePatientByID(ctx context.Context, patientID string, ownerEntityID string) error {
	filter := bson.M{
		"_id":             patientID,
		"owner_entity_id": ownerEntityID,
	}

	// Get collection name with context-based prefix
	collectionName := getCollectionNameWithPrefix(ctx, "patients")
	coll := m.Client.Database(DatabaseName).Collection(collectionName)

	// Delete the patient
	result, err := coll.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("error deleting patient: %v", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("patient not found")
	}

	return nil
}

// getCollectionNameWithPrefix returns the collection name with app-specific prefix based on context
func getCollectionNameWithPrefix(ctx context.Context, baseCollectionName string) string {
	appType := helpers.GetAppTypeFromContext(ctx)
	return appType + "_" + baseCollectionName
}

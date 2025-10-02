package clients

import (
	"context"
	"time"

	"github.com/therehabstreet/podoai/internal/common/helpers"
	"github.com/therehabstreet/podoai/internal/common/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBClient interface {
	FetchScans(ctx context.Context, userID string, page, pageSize int32, sortBy, sortOrder string) ([]models.Scan, int64, error)
	FetchScanByID(ctx context.Context, scanID string) (models.Scan, error)
	CreateScan(ctx context.Context, scan models.Scan) (models.Scan, error)
	DeleteScanByID(ctx context.Context, scanID string) error
	// Product methods
	FetchProductByID(ctx context.Context, productID string) (models.Product, error)
	// Exercise methods
	FetchExerciseByID(ctx context.Context, exerciseID string) (models.Exercise, error)
	// Therapy methods
	FetchTherapyByID(ctx context.Context, therapyID string) (models.Therapy, error)
	// OTP methods
	StoreOTP(ctx context.Context, otp *models.OTP) error
	GetOTPByMobileNumber(ctx context.Context, mobileNumber string) (*models.OTP, error)
	MarkOTPAsUsed(ctx context.Context, otpID primitive.ObjectID) error
	IncrementOTPAttempts(ctx context.Context, otpID primitive.ObjectID) error
	CleanupExpiredOTPs(ctx context.Context) error
}

// MongoDBClient implements DBClient
type MongoDBClient struct {
	Client *mongo.Client
}

const (
	DatabaseName = "podoai"
)

var CommonMongoClient *mongo.Client

func InitCommonMongoClient(uri string) (*MongoDBClient, error) {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	CommonMongoClient = client
	mongoClient := &MongoDBClient{Client: client}

	// Map of collection name to slice of index definitions
	indexMap := map[string][]mongo.IndexModel{
		"clinical_scans": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "clinic_id", Value: 1}, {Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_clinic_user_created_desc"),
			},
		},
		"consumer_scans": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
				Options: options.Index().SetName("idx_user_created_desc"),
			},
		},
		"otps": {
			mongo.IndexModel{
				Keys:    bson.D{{Key: "mobile_number", Value: 1}},
				Options: options.Index().SetName("idx_mobile_number"),
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
					{Key: "mobile_number", Value: 1},
					{Key: "is_used", Value: 1},
					{Key: "expires_at", Value: 1},
				},
				Options: options.Index().SetName("idx_mobile_used_expires"),
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

func (m *MongoDBClient) FetchScans(ctx context.Context, userID string, page, pageSize int32, sortBy, sortOrder string) ([]models.Scan, int64, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	filter := bson.M{"user_id": userID}
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
	var scans []models.Scan
	for cursor.Next(ctx) {
		var scan models.Scan
		if err := cursor.Decode(&scan); err != nil {
			continue
		}
		scans = append(scans, scan)
	}
	total, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		total = int64(len(scans))
	}
	return scans, total, nil
}

func (m *MongoDBClient) FetchScanByID(ctx context.Context, scanID string) (models.Scan, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	var scan models.Scan
	err := coll.FindOne(ctx, bson.M{"_id": scanID}).Decode(&scan)
	return scan, err
}

func (m *MongoDBClient) CreateScan(ctx context.Context, scan models.Scan) (models.Scan, error) {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	_, err := coll.InsertOne(ctx, scan)
	return scan, err
}

func (m *MongoDBClient) DeleteScanByID(ctx context.Context, scanID string) error {
	coll := m.Client.Database(DatabaseName).Collection(getCollectionNameWithPrefix(ctx, "scans"))
	_, err := coll.DeleteOne(ctx, bson.M{"_id": scanID})
	return err
}

// Product methods
func (m *MongoDBClient) FetchProductByID(ctx context.Context, productID string) (models.Product, error) {
	coll := m.Client.Database(DatabaseName).Collection("products")
	var product models.Product
	err := coll.FindOne(ctx, bson.M{"_id": productID}).Decode(&product)
	return product, err
}

// Exercise methods
func (m *MongoDBClient) FetchExerciseByID(ctx context.Context, exerciseID string) (models.Exercise, error) {
	coll := m.Client.Database(DatabaseName).Collection("exercises")
	var exercise models.Exercise
	err := coll.FindOne(ctx, bson.M{"_id": exerciseID}).Decode(&exercise)
	return exercise, err
}

// Therapy methods
func (m *MongoDBClient) FetchTherapyByID(ctx context.Context, therapyID string) (models.Therapy, error) {
	coll := m.Client.Database(DatabaseName).Collection("therapies")
	var therapy models.Therapy
	err := coll.FindOne(ctx, bson.M{"_id": therapyID}).Decode(&therapy)
	return therapy, err
}

// getCollectionNameWithPrefix returns the collection name with app-specific prefix based on context
func getCollectionNameWithPrefix(ctx context.Context, baseCollectionName string) string {
	if appType, ok := ctx.Value(helpers.AppTypeKey).(string); ok {
		switch appType {
		case helpers.AppTypeClinical:
			return "clinical_" + baseCollectionName
		case helpers.AppTypeConsumer:
			return "consumer_" + baseCollectionName
		}
	}
	// Default fallback - return base name without prefix
	return baseCollectionName
}

// OTP methods implementation

// StoreOTP stores a new OTP in the database
func (m *MongoDBClient) StoreOTP(ctx context.Context, otp *models.OTP) error {
	coll := m.Client.Database(DatabaseName).Collection("otps")

	// First, invalidate any existing OTPs for this mobile number
	_, err := coll.UpdateMany(ctx,
		bson.M{"mobile_number": otp.MobileNumber, "is_used": false},
		bson.M{"$set": bson.M{"is_used": true}})
	if err != nil {
		// Log error but continue with storing new OTP
	}

	// Store the new OTP
	_, err = coll.InsertOne(ctx, otp)
	return err
}

// GetOTPByMobileNumber retrieves the latest valid OTP for a mobile number
func (m *MongoDBClient) GetOTPByMobileNumber(ctx context.Context, mobileNumber string) (*models.OTP, error) {
	coll := m.Client.Database(DatabaseName).Collection("otps")

	// Find the latest unused, non-expired OTP for this mobile number
	filter := bson.M{
		"mobile_number": mobileNumber,
		"is_used":       false,
		"expires_at":    bson.M{"$gt": time.Now()},
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
func (m *MongoDBClient) MarkOTPAsUsed(ctx context.Context, otpID primitive.ObjectID) error {
	coll := m.Client.Database(DatabaseName).Collection("otps")

	_, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": otpID},
		bson.M{"$set": bson.M{"is_used": true}},
	)

	return err
}

// IncrementOTPAttempts increments the attempt count for an OTP
func (m *MongoDBClient) IncrementOTPAttempts(ctx context.Context, otpID primitive.ObjectID) error {
	coll := m.Client.Database(DatabaseName).Collection("otps")

	_, err := coll.UpdateOne(ctx,
		bson.M{"_id": otpID},
		bson.M{"$inc": bson.M{"attempts": 1}})

	return err
}

// CleanupExpiredOTPs removes expired OTPs from the database
func (m *MongoDBClient) CleanupExpiredOTPs(ctx context.Context) error {
	coll := m.Client.Database(DatabaseName).Collection("otps")

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

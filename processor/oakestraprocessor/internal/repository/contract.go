package repository

import (
	"context"
	"errors"

	"github.com/smnzlnsk/opentelemetry-components/processor/oakestraprocessor/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type contractRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

func NewContractRepository(collection *mongo.Collection, logger *zap.Logger) domain.ContractRepository {
	return &contractRepository{
		collection: collection,
		logger:     logger,
	}
}

func (r *contractRepository) Create(ctx context.Context, contract domain.CalculationContract) error {
	// Create a filter to find the service document
	filter := bson.M{"service": contract.Service}

	// Update options to create a new document if it doesn't exist
	upsert := true
	updateOptions := options.UpdateOptions{
		Upsert: &upsert,
	}

	// Check if formula already exists in the document
	var existingDoc domain.ContractDocument
	err := r.collection.FindOne(ctx, filter).Decode(&existingDoc)

	// If document exists, check if formula is already in the array
	if err == nil {
		for _, existingContract := range existingDoc.Contracts {
			if existingContract.Formula == contract.Formula {
				r.logger.Info("Formula already exists for this service, skipping",
					zap.String("service", contract.Service),
					zap.String("formula", contract.Formula))
				return nil
			}
		}
	} else if err != mongo.ErrNoDocuments {
		// If an error occurred that is not "document not found"
		return err
	}

	// Add formula to the array using $addToSet (to avoid duplicates)
	update := bson.M{
		"$addToSet": bson.M{
			"contracts": contract,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update, &updateOptions)
	if err != nil {
		return err
	}

	r.logger.Info("Formula created successfully",
		zap.String("service", contract.Service),
		zap.String("formula", contract.Formula))

	return nil
}

func (r *contractRepository) Update(ctx context.Context, old domain.CalculationContract, new domain.CalculationContract) error {
	// If service names are different, we need to handle this as a delete from one service and add to another
	if old.Service != new.Service {
		// First remove from the old service
		if err := r.DeleteFormula(ctx, old); err != nil {
			return err
		}
		// Then add to the new service
		return r.Create(ctx, new)
	}

	// For same service, find the document
	filter := bson.M{"service": old.Service}

	// First check if the document exists
	var existingDoc domain.ContractDocument
	err := r.collection.FindOne(ctx, filter).Decode(&existingDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return errors.New("service not found")
		}
		return err
	}

	// Check if the old formula exists
	formulaExists := false
	for _, existingContract := range existingDoc.Contracts {
		if existingContract.Formula == old.Formula {
			formulaExists = true
			break
		}
	}

	if !formulaExists {
		return errors.New("old formula not found")
	}

	// Remove the old formula
	update := bson.M{
		"$pull": bson.M{
			"contracts": bson.M{"formula": old.Formula},
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Add the new formula
	update = bson.M{
		"$addToSet": bson.M{
			"contracts": new,
		},
	}

	_, err = r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	r.logger.Info("Formula updated successfully",
		zap.String("service", old.Service),
		zap.String("oldFormula", old.Formula),
		zap.String("newFormula", new.Formula))

	return nil
}

func (r *contractRepository) DeleteFormula(ctx context.Context, contract domain.CalculationContract) error {
	// Create filter to find the service
	filter := bson.M{"service": contract.Service}

	// Remove the specific formula from the array
	update := bson.M{
		"$pull": bson.M{
			"contracts": bson.M{"formula": contract.Formula},
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("service not found")
	}

	// Check if the formulas array is now empty, if so, consider removing the document
	var doc domain.ContractDocument
	err = r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		return err
	}

	if len(doc.Contracts) == 0 {
		r.logger.Info("No contracts left for service, removing document",
			zap.String("service", contract.Service))
		_, err = r.collection.DeleteOne(ctx, filter)
		if err != nil {
			return err
		}
	}

	r.logger.Info("Formula deleted successfully",
		zap.String("service", contract.Service),
		zap.String("formula", contract.Formula))

	return nil
}

func (r *contractRepository) DeleteContract(ctx context.Context, service string) error {
	// Create filter to find the service
	filter := bson.M{"service": service}

	// Delete the entire document
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("service not found")
	}

	r.logger.Info("Contract deleted successfully",
		zap.String("service", service))

	return nil
}

func (r *contractRepository) GetContractsForProcessor(ctx context.Context, processor string) ([]domain.ContractDocument, error) {
	// Create a projection to filter contracts by processor
	projection := bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "service", Value: 1},
			{Key: "contracts", Value: bson.D{
				{Key: "$filter", Value: bson.D{
					{Key: "input", Value: "$contracts"},
					{Key: "as", Value: "contract"},
					{Key: "cond", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$$contract.processor", processor}},
					}},
				}},
			}},
		}},
	}

	// Execute the aggregation directly with the projection
	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{projection})
	if err != nil {
		r.logger.Error("Failed to execute aggregation for processor",
			zap.String("processor", processor),
			zap.Error(err))
		return nil, err
	}
	defer cursor.Close(ctx)

	// Decode the results
	var result []domain.ContractDocument
	if err := cursor.All(ctx, &result); err != nil {
		r.logger.Error("Failed to decode contracts", zap.Error(err))
		return nil, err
	}

	return result, nil
}

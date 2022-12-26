package database

import (
	"context"
	"graphQL/graph/model"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var connectionString string = "mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&appName=mongosh+1.6.0"

type DB struct {
	client *mongo.Client
}

func Connect() *DB {
	client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
	if err != nil {
		log.Panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}
	return &DB{
		client: client,
	}
}

func (db *DB) CreateJobListing(jobInfo model.CreateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("graphql").Collection("job")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	inserted, err := jobCollection.InsertOne(ctx, bson.M{
		"title":       jobInfo.Title,
		"description": jobInfo.Description,
		"url":         jobInfo.URL,
		"company":     jobInfo.Company,
	})
	if err != nil {
		panic(err)
	}
	insertedID := inserted.InsertedID.(primitive.ObjectID).Hex()
	jobListing := model.JobListing{ID: insertedID, Title: jobInfo.Title, Description: jobInfo.Description, Company: jobInfo.Company, URL: jobInfo.Company}
	return &jobListing
}

func (db *DB) UpdateJobListing(id string, jobInfo model.UpdateJobListingInput) *model.JobListing {
	jobCollection := db.client.Database("graphql").Collection("job")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	updateJobInfo := bson.M{}
	if jobInfo.Title != nil {
		updateJobInfo["title"] = jobInfo.Title
	}
	if jobInfo.Company != nil {
		updateJobInfo["company"] = jobInfo.Company
	}
	if jobInfo.URL != nil {
		updateJobInfo["url"] = jobInfo.URL
	}
	if jobInfo.Description != nil {
		updateJobInfo["description"] = jobInfo.Description
	}
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	update := bson.M{"$set": updateJobInfo}
	result := jobCollection.FindOneAndUpdate(ctx, filter, update, options.FindOneAndUpdate().SetReturnDocument(1))

	var jobListing model.JobListing
	if err := result.Decode(&jobCollection); err != nil {
		panic(err)
	}
	return &jobListing
}

func (db *DB) DeleteJobListing(id string) *model.DeleteJobResponse {
	jobCollection := db.client.Database("graphql").Collection("job")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_id, _ := primitive.ObjectIDFromHex(id)
	filter := bson.M{"_id": _id}
	_, err := jobCollection.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	return &model.DeleteJobResponse{
		DeleteJobID: id,
	}
}

func (db *DB) GetJob(id string) *model.JobListing {
	jobCollection := db.client.Database("graphql").Collection("job")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_id, _ := primitive.ObjectIDFromHex((id))
	filter := bson.M{"_id": _id}
	var jobListing *model.JobListing
	if err := jobCollection.FindOne(ctx, filter).Decode(&jobListing); err != nil {
		log.Fatal(err.Error())
	}
	return jobListing
}

func (db *DB) GetJobs() []*model.JobListing {
	jobCollection := db.client.Database("graphql").Collection("job")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var jobListing []*model.JobListing
	cursor, err := jobCollection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	if err = cursor.All(context.TODO(), &jobListing); err != nil {
		log.Fatal(err)
	}
	return jobListing
}

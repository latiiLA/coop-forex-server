package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/latiiLA/coop-forex-server/internal/domain/model"
	"github.com/latiiLA/coop-forex-server/internal/repository"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userRepositoryTestSuite struct {
	suite.Suite
	db         *mongo.Database
	collection *mongo.Collection
	repo       model.UserRepository
	ctx        context.Context
	cancel     context.CancelFunc
}

func (s *userRepositoryTestSuite) SetupSuite() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	s.Require().NoError(err)

	s.db = client.Database("test_db")
	s.collection = s.db.Collection("users")

	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Second)
	s.repo = repository.NewUserRepository(s.db)
}

func (s *userRepositoryTestSuite) TearDownSuite() {
	s.cancel()
	_ = s.db.Drop(s.ctx)
}

func (s *userRepositoryTestSuite) TestCreateUser() {
	profileID := primitive.NewObjectID()
	roleID := primitive.NewObjectID()

	_, err := s.collection.InsertOne(s.ctx, bson.M{
		"_id":         profileID,
		"email":       "test@example.com",
		"first_name":  "Test",
		"middle_name": "User",
		"last_name":   "Demo",
	})
	s.Require().NoError(err)

	updatedBy := primitive.NewObjectID()
	user := &model.User{
		ID:        primitive.NewObjectID(),
		ProfileID: profileID,
		RoleID:    roleID,
		Username:  "testuser",
		Status:    "new",
		IsDeleted: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		CreatedBy: primitive.NewObjectID(),
		UpdatedBy: &updatedBy,
	}

	err = s.repo.Create(s.ctx, user)
	s.Require().NoError(err)

	foundUser, err := s.repo.FindByUsername(s.ctx, "testuser")
	s.Require().NoError(err)
	s.Require().Equal("testuser", foundUser.Username)
	s.Require().False(foundUser.IsDeleted)
}

func (s *userRepositoryTestSuite) TestFindByID() {
	user := &model.UserResponseDTO{
		ID:        primitive.NewObjectID(),
		Username:  "findbyuserid",
		IsDeleted: false,
	}

	_, err := s.collection.InsertOne(s.ctx, user)
	s.Require().NoError(err)

	foundUser, err := s.repo.FindByID(s.ctx, user.ID)
	s.Require().NoError(err)
	s.Require().Equal("findbyuserid", foundUser.Username)
}

func (s *userRepositoryTestSuite) TestUpdate() {
	user := &model.User{
		ID:        primitive.NewObjectID(),
		Username:  "updateuser",
		IsDeleted: false,
	}

	_, err := s.collection.InsertOne(s.ctx, user)
	s.Require().NoError(err)

	updated := &model.User{
		Username:  "updateduser",
		IsDeleted: false,
	}

	result, err := s.repo.Update(s.ctx, user.ID, updated)
	s.Require().NoError(err)
	s.Require().Equal("updateduser", result.Username)
}

func (s *userRepositoryTestSuite) TestDelete() {
	user := &model.User{
		ID:        primitive.NewObjectID(),
		Username:  "tobedeleted",
		IsDeleted: false,
		Status:    "active",
	}

	_, err := s.collection.InsertOne(s.ctx, user)
	s.Require().NoError(err)

	err = s.repo.Delete(s.ctx, user.ID, user)
	s.Require().NoError(err)

	deletedUser, err := s.repo.FindByID(s.ctx, user.ID)
	s.Require().Error(err)
	s.Require().Nil(deletedUser)
}

func (s *userRepositoryTestSuite) TestFindAll() {
	profileID := primitive.NewObjectID()
	adminID := primitive.NewObjectID()
	_, err := s.db.Collection("profiles").InsertOne(s.ctx, bson.M{
		"_id":         profileID,
		"email":       "sample@user.com",
		"first_name":  "Sample",
		"middle_name": "M",
		"last_name":   "User",
	})
	s.Require().NoError(err)

	_, err = s.collection.InsertOne(s.ctx, bson.M{
		"_id":        primitive.NewObjectID(),
		"profile_id": profileID,
		"role_id":    primitive.NewObjectID(),
		"username":   "alluser",
		"status":     "active",
		"is_deleted": false,
		"created_at": time.Now(),
		"updated_at": time.Now(),
		"created_by": adminID,
		"updated_by": &adminID,
	})
	s.Require().NoError(err)

	users, err := s.repo.FindAll(s.ctx)
	s.Require().NoError(err)
	s.Require().NotEmpty(users)
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(userRepositoryTestSuite))
}

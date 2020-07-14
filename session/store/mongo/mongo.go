package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/session/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(params storeStructuredMongo.Params) (*Store, error) {
	baseStore, err := storeStructuredMongo.NewStore(params)
	if err != nil {
		return nil, err
	}

	return &Store{
		Store: baseStore,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
}

func (s *Store) NewTokenRepository() store.TokenRepository {
	return &TokenRepository{
		s.Store.GetRepository("tokens"),
	}
}

type TokenRepository struct {
	*storeStructuredMongo.Repository
}

func (s *TokenRepository) DestroySessionsForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()

	selector := bson.M{
		"userId": userID,
	}
	removeInfo, err := s.DeleteMany(ctx, selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroySessionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy sessions for user by id")
	}
	return nil
}

package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/log"
	"github.com/tidepool-org/platform/permission"
	"github.com/tidepool-org/platform/permission/store"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

func NewStore(cfg *Config, lgr log.Logger) (*Store, error) {
	if cfg == nil {
		return nil, errors.New("config is missing")
	}

	baseStore, err := storeStructuredMongo.NewStore(cfg.Config)
	if err != nil {
		return nil, err
	}

	if err = cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "config is invalid")
	}

	return &Store{
		Store:  baseStore,
		config: cfg,
	}, nil
}

type Store struct {
	*storeStructuredMongo.Store
	config *Config
}

func (s *Store) NewPermissionsRepository() store.PermissionsRepository {
	return &PermissionsRepository{
		Repository: s.Store.GetRepository("perms"),
		config:     s.config,
	}
}

type PermissionsRepository struct {
	*storeStructuredMongo.Repository
	config *Config
}

func (p *PermissionsRepository) DestroyPermissionsForUserByID(ctx context.Context, userID string) error {
	if ctx == nil {
		return errors.New("context is missing")
	}
	if userID == "" {
		return errors.New("user id is missing")
	}

	now := time.Now()

	groupID, err := permission.GroupIDFromUserID(userID, p.config.Secret)
	if err != nil {
		return errors.Wrap(err, "unable to determine group id from user id")
	}

	selector := bson.M{
		"$or": []bson.M{
			{"groupId": groupID},
			{"userId": userID},
		},
	}
	removeInfo, err := p.DeleteMany(ctx, selector)

	loggerFields := log.Fields{"userId": userID, "removeInfo": removeInfo, "duration": time.Since(now) / time.Microsecond}
	log.LoggerFromContext(ctx).WithFields(loggerFields).WithError(err).Debug("DestroyPermissionsForUserByID")

	if err != nil {
		return errors.Wrap(err, "unable to destroy permissions for user by id")
	}
	return nil
}

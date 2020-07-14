package mongo_test

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"

	"github.com/tidepool-org/platform/blob"
	blobStoreStructured "github.com/tidepool-org/platform/blob/store/structured"
	blobStoreStructuredMongo "github.com/tidepool-org/platform/blob/store/structured/mongo"
	blobStoreStructuredTest "github.com/tidepool-org/platform/blob/store/structured/test"
	blobTest "github.com/tidepool-org/platform/blob/test"
	"github.com/tidepool-org/platform/errors"
	errorsTest "github.com/tidepool-org/platform/errors/test"
	"github.com/tidepool-org/platform/log"
	logTest "github.com/tidepool-org/platform/log/test"
	netTest "github.com/tidepool-org/platform/net/test"
	"github.com/tidepool-org/platform/page"
	"github.com/tidepool-org/platform/pointer"
	"github.com/tidepool-org/platform/request"
	requestTest "github.com/tidepool-org/platform/request/test"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
	storeStructuredMongoTest "github.com/tidepool-org/platform/store/structured/mongo/test"
	"github.com/tidepool-org/platform/test"
	userTest "github.com/tidepool-org/platform/user/test"
)

type CreatedTimeDescending blob.BlobArray

func (c CreatedTimeDescending) Len() int {
	return len(c)
}

func (c CreatedTimeDescending) Less(left int, right int) bool {
	if c[left].CreatedTime == nil {
		return true
	} else if c[right].CreatedTime == nil {
		return false
	}
	return c[right].CreatedTime.Before(*c[left].CreatedTime)
}

func (c CreatedTimeDescending) Swap(left int, right int) {
	c[left], c[right] = c[right], c[left]
}

func SelectAndSort(blobs blob.BlobArray, selector func(b *blob.Blob) bool) blob.BlobArray {
	var selected blob.BlobArray
	for _, b := range blobs {
		if selector(b) {
			selected = append(selected, b)
		}
	}
	sort.Sort(CreatedTimeDescending(selected))
	return selected
}

func AsInterfaceArray(blobs blob.BlobArray) []interface{} {
	if blobs == nil {
		return nil
	}
	array := make([]interface{}, len(blobs))
	for index, b := range blobs {
		array[index] = b
	}
	return array
}

var _ = Describe("Mongo", func() {
	var config *storeStructuredMongo.Config
	var logger *logTest.Logger
	var store *blobStoreStructuredMongo.Store
	var coll blobStoreStructured.BlobRepository

	BeforeEach(func() {
		config = storeStructuredMongoTest.NewConfig()
		logger = logTest.NewLogger()
	})

	AfterEach(func() {
		if store != nil {
			store.Terminate(nil)
		}
	})

	Context("NewStore", func() {
		It("returns an error when unsuccessful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: nil}
			store, err = blobStoreStructuredMongo.NewStore(params)
			errorsTest.ExpectEqual(err, errors.New("database config is empty"))
			Expect(store).To(BeNil())
		})

		It("returns a new store and no error when successful", func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = blobStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
		})
	})

	Context("with a new store", func() {
		var collection *mongo.Collection

		BeforeEach(func() {
			var err error
			params := storeStructuredMongo.Params{DatabaseConfig: config}
			store, err = blobStoreStructuredMongo.NewStore(params)
			Expect(err).ToNot(HaveOccurred())
			Expect(store).ToNot(BeNil())
			collection = store.GetCollection("blobs")
		})

		Context("EnsureIndexes", func() {
			It("returns successfully", func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				cursor, err := collection.Indexes().List(context.Background())
				Expect(err).ToNot(HaveOccurred())
				Expect(cursor).ToNot(BeNil())
				var indexes []storeStructuredMongoTest.MongoIndex
				err = cursor.All(context.Background(), &indexes)
				Expect(err).ToNot(HaveOccurred())

				Expect(indexes).To(ConsistOf(
					MatchFields(IgnoreExtras, Fields{
						"Key": Equal(storeStructuredMongoTest.MakeKeySlice("_id")),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("id")),
						"Background": Equal(true),
						"Unique":     Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("userId")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("mediaType")),
						"Background": Equal(true),
					}),
					MatchFields(IgnoreExtras, Fields{
						"Key":        Equal(storeStructuredMongoTest.MakeKeySlice("status")),
						"Background": Equal(true),
					}),
				))
			})
		})

		Context("NewBlobRepository", func() {
			It("returns a new session", func() {
				coll = store.NewBlobRepository()
				Expect(coll).ToNot(BeNil())
			})
		})

		Context("with a new session", func() {
			var ctx context.Context

			BeforeEach(func() {
				Expect(store.EnsureIndexes()).To(Succeed())
				coll = store.NewBlobRepository()
				ctx = log.NewContextWithLogger(context.Background(), logger)
			})

			Context("with user id", func() {
				var userID string

				BeforeEach(func() {
					userID = userTest.RandomID()
				})

				Context("List", func() {
					var filter *blob.Filter
					var pagination *page.Pagination

					BeforeEach(func() {
						filter = blob.NewFilter()
						pagination = page.NewPagination()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := coll.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := coll.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := coll.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the filter is invalid", func() {
						filter.MediaType = pointer.FromStringArray([]string{""})
						result, err := coll.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("filter is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the pagination is invalid", func() {
						pagination.Page = -1
						result, err := coll.List(ctx, userID, filter, pagination)
						errorsTest.ExpectEqual(err, errors.New("pagination is invalid"))
						Expect(result).To(BeNil())
					})

					Context("with data", func() {
						var mediaType string
						var allResult blob.BlobArray

						BeforeEach(func() {
							mediaType = netTest.RandomMediaType()
							allResult = blobTest.RandomBlobArray(8, 8)
							for index, result := range allResult {
								result.ID = pointer.FromString(blobTest.RandomID())
								result.UserID = pointer.FromString(userID)
								if (index/4)%2 == 0 {
									result.MediaType = pointer.FromString(mediaType)
								}
								result.Status = pointer.FromString(blob.Statuses()[(index/2)%2])
								if index%2 == 0 {
									result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
									result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
								}
							}
							allResult = append(allResult, blobTest.RandomBlob(), blobTest.RandomBlob())
							rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
							_, err := collection.InsertMany(context.Background(), AsInterfaceArray(allResult))
							Expect(err).ShouldNot(HaveOccurred())
						})

						It("returns no result when the user id is unknown", func() {
							userID = userTest.RandomID()
							Expect(coll.List(ctx, userID, filter, pagination)).To(SatisfyAll(Not(BeNil()), BeEmpty()))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 0})
						})

						It("returns expected result when the filter is missing", func() {
							filter = nil
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter media type is missing", func() {
							filter.MediaType = nil
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter media type is specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.MediaType == mediaType && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter status is missing", func() {
							filter.Status = nil
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to available", func() {
							filter.Status = pointer.FromStringArray([]string{blob.StatusAvailable})
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to created", func() {
							filter.Status = pointer.FromStringArray([]string{blob.StatusCreated})
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.Status == blob.StatusCreated
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})

						It("returns expected result when the filter status is set to both available and created", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool { return *b.UserID == userID && b.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 4})
						})

						It("returns expected result when the filter media type and status available are specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							filter.Status = pointer.FromStringArray([]string{blob.StatusAvailable})
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.MediaType == mediaType && *b.Status == blob.StatusAvailable
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the filter media type and status created are specified", func() {
							filter.MediaType = pointer.FromStringArray([]string{netTest.RandomMediaType(), mediaType})
							filter.Status = pointer.FromStringArray([]string{blob.StatusCreated})
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool {
									return *b.UserID == userID && b.DeletedTime == nil && *b.MediaType == mediaType && *b.Status == blob.StatusCreated
								},
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 1})
						})

						It("returns expected result when the pagination is missing", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							pagination = nil
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool { return *b.UserID == userID && b.DeletedTime == nil },
							)))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "count": 4})
						})

						It("returns expected result when the pagination limits result", func() {
							filter.Status = pointer.FromStringArray(blob.Statuses())
							pagination.Page = 1
							pagination.Size = 2
							Expect(coll.List(ctx, userID, filter, pagination)).To(Equal(SelectAndSort(allResult,
								func(b *blob.Blob) bool { return *b.UserID == userID && b.DeletedTime == nil },
							)[2:4]))
							logger.AssertDebug("List", log.Fields{"userId": userID, "filter": filter, "pagination": pagination, "count": 2})
						})
					})
				})

				Context("Create", func() {
					var create *blobStoreStructured.Create

					BeforeEach(func() {
						create = blobStoreStructuredTest.RandomCreate()
					})

					It("returns an error when the context is missing", func() {
						ctx = nil
						result, err := coll.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						result, err := coll.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						result, err := coll.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is missing", func() {
						create = nil
						result, err := coll.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is missing"))
						Expect(result).To(BeNil())
					})

					It("returns an error when the create is invalid", func() {
						create.MediaType = pointer.FromString("")
						result, err := coll.Create(ctx, userID, create)
						errorsTest.ExpectEqual(err, errors.New("create is invalid"))
						Expect(result).To(BeNil())
					})

					It("returns the result after creating", func() {
						matchAllFields := MatchAllFields(Fields{
							"ID":           PointTo(Not(BeEmpty())),
							"UserID":       PointTo(Equal(userID)),
							"DigestMD5":    BeNil(),
							"MediaType":    Equal(create.MediaType),
							"Size":         BeNil(),
							"Status":       PointTo(Equal(blob.StatusCreated)),
							"CreatedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime": BeNil(),
							"DeletedTime":  BeNil(),
							"Revision":     PointTo(Equal(0)),
						})
						result, err := coll.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := blob.BlobArray{}
						cursor, err := collection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *storeResult[0].ID})
					})

					It("returns the result after creating without media type", func() {
						create.MediaType = nil
						matchAllFields := MatchAllFields(Fields{
							"ID":           PointTo(Not(BeEmpty())),
							"UserID":       PointTo(Equal(userID)),
							"DigestMD5":    BeNil(),
							"MediaType":    BeNil(),
							"Size":         BeNil(),
							"Status":       PointTo(Equal(blob.StatusCreated)),
							"CreatedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
							"ModifiedTime": BeNil(),
							"DeletedTime":  BeNil(),
							"Revision":     PointTo(Equal(0)),
						})
						result, err := coll.Create(ctx, userID, create)
						Expect(err).ToNot(HaveOccurred())
						Expect(result).ToNot(BeNil())
						Expect(*result).To(matchAllFields)
						storeResult := blob.BlobArray{}
						cursor, err := collection.Find(context.Background(), bson.M{"id": result.ID})
						Expect(err).ToNot(HaveOccurred())
						Expect(cursor).ToNot(BeNil())
						Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
						Expect(storeResult).To(HaveLen(1))
						Expect(*storeResult[0]).To(matchAllFields)
						logger.AssertDebug("Create", log.Fields{"userId": userID, "create": create, "id": *storeResult[0].ID})
					})
				})

				Context("DeleteAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						deleted, err := coll.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						deleted, err := coll.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(deleted).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						deleted, err := coll.DeleteAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(deleted).To(BeFalse())
					})

					Context("with data", func() {
						var originals blob.BlobArray

						BeforeEach(func() {
							originals = blobTest.RandomBlobArray(4, 4)
							for index, original := range originals {
								original.UserID = pointer.FromString(userID)
								if index%2 == 0 {
									original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
									original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
								}
								_, err := collection.InsertOne(context.Background(), original)
								Expect(err).ToNot(HaveOccurred())
							}
							_, err := collection.InsertMany(context.Background(), []interface{}{blobTest.RandomBlob(), blobTest.RandomBlob()})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							logger.AssertDebug("DeleteAll", log.Fields{"userId": userID})
						})

						It("returns false and does not delete the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(coll.DeleteAll(ctx, userID)).To(BeFalse())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": originalUserID, "deletedTime": bson.M{"$exists": true}})).To(Equal(int64(2)))
							Expect(collection.CountDocuments(context.Background(), bson.M{"deletedTime": bson.M{"$exists": true}})).To(Equal(int64(2)))
						})

						It("returns true and deletes the originals when the user id matches", func() {
							Expect(coll.DeleteAll(ctx, userID)).To(BeTrue())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": userID, "deletedTime": bson.M{"$exists": true}})).To(Equal(int64(4)))
							Expect(collection.CountDocuments(context.Background(), bson.M{"deletedTime": bson.M{"$exists": true}})).To(Equal(int64(4)))
						})
					})
				})

				Context("DestroyAll", func() {
					It("returns an error when the context is missing", func() {
						ctx = nil
						destroyed, err := coll.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("context is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is missing", func() {
						userID = ""
						destroyed, err := coll.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is missing"))
						Expect(destroyed).To(BeFalse())
					})

					It("returns an error when the user id is invalid", func() {
						userID = "invalid"
						destroyed, err := coll.DestroyAll(ctx, userID)
						errorsTest.ExpectEqual(err, errors.New("user id is invalid"))
						Expect(destroyed).To(BeFalse())
					})

					Context("with data", func() {
						var originals blob.BlobArray

						BeforeEach(func() {
							originals = blobTest.RandomBlobArray(4, 4)
							for index, original := range originals {
								original.UserID = pointer.FromString(userID)
								if index%2 == 0 {
									original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
									original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
								}
								_, err := collection.InsertOne(context.Background(), original)
								Expect(err).ToNot(HaveOccurred())
							}
							_, err := collection.InsertMany(context.Background(), []interface{}{blobTest.RandomBlob(), blobTest.RandomBlob()})
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							logger.AssertDebug("DestroyAll", log.Fields{"userId": userID})
						})

						It("returns false and does not destroy the originals when the user id does not match", func() {
							originalUserID := userID
							userID = userTest.RandomID()
							Expect(coll.DestroyAll(ctx, userID)).To(BeFalse())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": originalUserID})).To(Equal(int64(4)))
							Expect(collection.CountDocuments(context.Background(), bson.M{})).To(Equal(int64(6)))
						})

						It("returns true and destroys the originals when the user id matches", func() {
							Expect(coll.DestroyAll(ctx, userID)).To(BeTrue())
							Expect(collection.CountDocuments(context.Background(), bson.M{"userId": userID})).To(Equal(int64(0)))
							Expect(collection.CountDocuments(context.Background(), bson.M{})).To(Equal(int64(2)))
						})
					})
				})
			})

			Context("Get", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := coll.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := coll.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := coll.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := coll.Get(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var allResult blob.BlobArray
					var result *blob.Blob

					BeforeEach(func() {
						allResult = blobTest.RandomBlobArray(3, 3)
						result = allResult[0]
						result.ID = pointer.FromString(id)
						rand.Shuffle(len(allResult), func(i, j int) { allResult[i], allResult[j] = allResult[j], allResult[i] })
					})

					JustBeforeEach(func() {
						_, err := collection.InsertMany(context.Background(), AsInterfaceArray(allResult))
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						logger.AssertDebug("Get", log.Fields{"id": id})
					})

					It("returns nil when the id does not exist", func() {
						id = blobTest.RandomID()
						Expect(coll.Get(ctx, id, condition)).To(BeNil())
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*result.Revision + 1)
						})

						It("returns nil", func() {
							Expect(coll.Get(ctx, id, condition)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						It("returns the result when the id exists", func() {
							Expect(coll.Get(ctx, id, condition)).To(Equal(result))
						})

						Context("when the result is marked as deleted", func() {
							BeforeEach(func() {
								result.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*result.CreatedTime, time.Now()).Truncate(time.Second))
								result.DeletedTime = pointer.CloneTime(result.ModifiedTime)
							})

							It("returns nil", func() {
								Expect(coll.Get(ctx, id, condition)).To(BeNil())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns the result with revision 0", func() {
								result.Revision = pointer.FromInt(0)
								Expect(coll.Get(ctx, id, condition)).To(Equal(result))
							})
						})
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns the result with revision 0", func() {
								result.Revision = pointer.FromInt(0)
								Expect(coll.Get(ctx, id, condition)).To(Equal(result))
							})
						})
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.CloneInt(result.Revision)
						})

						conditionAssertions()

						Context("when the revision is missing", func() {
							BeforeEach(func() {
								result.Revision = nil
							})

							It("returns nil", func() {
								Expect(coll.Get(ctx, id, condition)).To(BeNil())
							})
						})
					})
				})
			})

			Context("Update", func() {
				var id string
				var condition *request.Condition
				var update *blobStoreStructured.Update

				BeforeEach(func() {
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
					update = blobStoreStructuredTest.RandomUpdate()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is missing", func() {
					update = nil
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is missing"))
					Expect(result).To(BeNil())
				})

				It("returns an error when the update is invalid", func() {
					update.DigestMD5 = pointer.FromString("")
					result, err := coll.Update(ctx, id, condition, update)
					errorsTest.ExpectEqual(err, errors.New("update is invalid"))
					Expect(result).To(BeNil())
				})

				Context("with data", func() {
					var original *blob.Blob

					BeforeEach(func() {
						original = blobTest.RandomBlob()
						original.ID = pointer.FromString(id)
						_, err := collection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Update", log.Fields{"id": id, "condition": condition, "update": update})
						} else {
							logger.AssertDebug("Update", log.Fields{"id": id, "update": update})
						}
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns nil", func() {
							Expect(coll.Update(ctx, id, condition, update)).To(BeNil())
						})
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							It("returns updated result when the id exists", func() {
								matchAllFields := MatchAllFields(Fields{
									"ID":           PointTo(Equal(id)),
									"UserID":       Equal(original.UserID),
									"DigestMD5":    Equal(update.DigestMD5),
									"MediaType":    Equal(update.MediaType),
									"Size":         Equal(update.Size),
									"Status":       Equal(update.Status),
									"CreatedTime":  Equal(original.CreatedTime),
									"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
									"DeletedTime":  BeNil(),
									"Revision":     PointTo(Equal(*original.Revision + 1)),
								})
								result, err := coll.Update(ctx, id, condition, update)
								Expect(err).ToNot(HaveOccurred())
								Expect(result).ToNot(BeNil())
								Expect(*result).To(matchAllFields)
								storeResult := blob.BlobArray{}
								cursor, err := collection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns nil when the id does not exist", func() {
								id = blobTest.RandomID()
								Expect(coll.Update(ctx, id, condition, update)).To(BeNil())
							})
						})

						Context("without updates", func() {
							BeforeEach(func() {
								update = blobStoreStructured.NewUpdate()
							})

							It("returns original when the id exists", func() {
								Expect(coll.Update(ctx, id, condition, update)).To(Equal(original))
							})

							It("returns nil when the id does not exist", func() {
								id = blobTest.RandomID()
								Expect(coll.Update(ctx, id, condition, update)).To(BeNil())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						conditionAssertions()
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						conditionAssertions()
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.CloneInt(original.Revision)
						})

						conditionAssertions()
					})
				})
			})

			Context("Delete", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					result, err := coll.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					result, err := coll.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					result, err := coll.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(result).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					result, err := coll.Delete(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(result).To(BeFalse())
				})

				Context("with data", func() {
					var original *blob.Blob

					BeforeEach(func() {
						original = blobTest.RandomBlob()
						original.ID = pointer.FromString(id)
					})

					JustBeforeEach(func() {
						_, err := collection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Delete", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Delete", log.Fields{"id": id})
						}
					})

					When("the original is marked as deleted", func() {
						BeforeEach(func() {
							original.ModifiedTime = pointer.FromTime(test.RandomTimeFromRange(*original.CreatedTime, time.Now()).Truncate(time.Second))
							original.DeletedTime = pointer.CloneTime(original.ModifiedTime)
						})

						It("returns false", func() {
							Expect(coll.Delete(ctx, id, condition)).To(BeFalse())
						})
					})

					When("the condition revision does not match", func() {
						BeforeEach(func() {
							condition.Revision = pointer.FromInt(*original.Revision + 1)
						})

						It("returns false", func() {
							Expect(coll.Delete(ctx, id, condition)).To(BeFalse())
						})
					})

					conditionAssertions := func() {
						Context("with updates", func() {
							It("returns true when the id exists", func() {
								matchAllFields := MatchAllFields(Fields{
									"ID":           PointTo(Equal(id)),
									"UserID":       Equal(original.UserID),
									"DigestMD5":    Equal(original.DigestMD5),
									"MediaType":    Equal(original.MediaType),
									"Size":         Equal(original.Size),
									"Status":       Equal(original.Status),
									"CreatedTime":  Equal(original.CreatedTime),
									"ModifiedTime": PointTo(BeTemporally("~", time.Now(), time.Second)),
									"DeletedTime":  PointTo(BeTemporally("~", time.Now(), time.Second)),
									"Revision":     PointTo(Equal(*original.Revision + 1)),
								})
								Expect(coll.Delete(ctx, id, condition)).To(BeTrue())
								storeResult := blob.BlobArray{}
								cursor, err := collection.Find(context.Background(), bson.M{"id": id})
								Expect(err).ToNot(HaveOccurred())
								Expect(cursor).ToNot(BeNil())
								Expect(cursor.All(context.Background(), &storeResult)).To(Succeed())
								Expect(storeResult).To(HaveLen(1))
								Expect(*storeResult[0]).To(matchAllFields)
							})

							It("returns false when the id does not exist", func() {
								id = blobTest.RandomID()
								Expect(coll.Delete(ctx, id, condition)).To(BeFalse())
							})
						})
					}

					When("the condition is missing", func() {
						BeforeEach(func() {
							condition = nil
						})

						conditionAssertions()
					})

					When("the condition revision is missing", func() {
						BeforeEach(func() {
							condition.Revision = nil
						})

						conditionAssertions()
					})

					When("the condition revision matches", func() {
						BeforeEach(func() {
							condition.Revision = pointer.CloneInt(original.Revision)
						})

						conditionAssertions()
					})
				})
			})

			Context("Destroy", func() {
				var id string
				var condition *request.Condition

				BeforeEach(func() {
					id = blobTest.RandomID()
					condition = requestTest.RandomCondition()
				})

				It("returns an error when the context is missing", func() {
					ctx = nil
					deleted, err := coll.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("context is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is missing", func() {
					id = ""
					deleted, err := coll.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is missing"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the id is invalid", func() {
					id = "invalid"
					deleted, err := coll.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("id is invalid"))
					Expect(deleted).To(BeFalse())
				})

				It("returns an error when the condition is invalid", func() {
					condition.Revision = pointer.FromInt(-1)
					deleted, err := coll.Destroy(ctx, id, condition)
					errorsTest.ExpectEqual(err, errors.New("condition is invalid"))
					Expect(deleted).To(BeFalse())
				})

				Context("with data", func() {
					var original *blob.Blob

					BeforeEach(func() {
						original = blobTest.RandomBlob()
						original.ID = pointer.FromString(id)
						_, err := collection.InsertOne(context.Background(), original)
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						if condition != nil {
							logger.AssertDebug("Destroy", log.Fields{"id": id, "condition": condition})
						} else {
							logger.AssertDebug("Destroy", log.Fields{"id": id})
						}
					})

					It("returns false and does not delete the original when the id does not exist", func() {
						id = blobTest.RandomID()
						Expect(coll.Destroy(ctx, id, condition)).To(BeFalse())
						Expect(collection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns false and does not delete the original when the id exists, but the condition revision does not match", func() {
						condition.Revision = pointer.FromInt(*original.Revision + 1)
						Expect(coll.Destroy(ctx, id, condition)).To(BeFalse())
						Expect(collection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(1)))
					})

					It("returns true and deletes the original when the id exists and the condition is missing", func() {
						condition = nil
						Expect(coll.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(collection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision is missing", func() {
						condition.Revision = nil
						Expect(coll.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(collection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})

					It("returns true and deletes the original when the id exists and the condition revision matches", func() {
						condition.Revision = pointer.CloneInt(original.Revision)
						Expect(coll.Destroy(ctx, id, condition)).To(BeTrue())
						Expect(collection.CountDocuments(context.Background(), bson.M{"id": original.ID})).To(Equal(int64(0)))
					})
				})
			})
		})
	})
})

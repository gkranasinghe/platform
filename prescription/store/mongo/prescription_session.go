package mongo

import (
	"context"
	"time"

	"github.com/globalsign/mgo/bson"

	"github.com/tidepool-org/platform/page"

	"github.com/tidepool-org/platform/log"
	structureValidator "github.com/tidepool-org/platform/structure/validator"

	mgo "github.com/globalsign/mgo"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
	storeStructuredMongo "github.com/tidepool-org/platform/store/structured/mongo"
)

type PrescriptionSession struct {
	*storeStructuredMongo.Session
}

func (p *PrescriptionSession) EnsureIndexes() error {
	return p.EnsureAllIndexes([]mgo.Index{
		{Key: []string{"patientId"}, Background: true},
		{Key: []string{"prescriberId"}, Background: true},
		{Key: []string{"createdUserId"}, Background: true},
		{Key: []string{"accessCode"}, Unique: true, Background: true},
		{Key: []string{"latestRevision.attributes.email"}, Background: true, Name: "latest_patient_email"},
	})
}

func (p *PrescriptionSession) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if userID == "" {
		return nil, errors.New("userID is missing")
	}

	model := prescription.NewPrescription(userID, create)
	if err := structureValidator.New().Validate(model); err != nil {
		return nil, errors.Wrap(err, "prescription is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"userId": userID, "create": create})

	err := p.C().Insert(model)
	logger.WithFields(log.Fields{"id": model.ID, "duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("CreatePrescription")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create user restricted token")
	}

	return model, nil
}

func (p *PrescriptionSession) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if filter == nil {
		return nil, errors.New("filter is missing")
	} else if err := structureValidator.New().Validate(filter); err != nil {
		return nil, errors.Wrap(err, "filter is invalid")
	}

	if pagination == nil {
		pagination = page.NewPagination()
	} else if err := structureValidator.New().Validate(pagination); err != nil {
		return nil, errors.Wrap(err, "pagination is invalid")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"filter": filter})

	selector := newMongoSelectorFromFilter(filter)
	selector["deletedTime"] = nil

	prescriptions := prescription.Prescriptions{}
	err := p.C().Find(selector).Skip(pagination.Page * pagination.Size).Limit(pagination.Size).All(&prescriptions)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("ListPrescriptions")
	if err != nil {
		return nil, errors.Wrap(err, "unable to list prescriptions")
	}

	return prescriptions, nil
}

func (p *PrescriptionSession) GetUnclaimedPrescription(ctx context.Context, accessCode string) (*prescription.Prescription, error) {
	if ctx == nil {
		return nil, errors.New("context is missing")
	}
	if p.IsClosed() {
		return nil, errors.New("session closed")
	}
	if accessCode == "" {
		return nil, errors.New("access code is missing")
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"accessCode": accessCode})

	selector := bson.M{
		"accessCode": accessCode,
		"patientId":  nil,
		"state":      prescription.StateSubmitted,
	}

	prescr := &prescription.Prescription{}
	err := p.C().Find(selector).One(prescr)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("GetUnclaimedPrescription")
	if err == mgo.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to find unclaimed prescription")
	}

	return prescr, nil
}

func (p *PrescriptionSession) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is missing")
	}
	if p.IsClosed() {
		return false, errors.New("session closed")
	}
	if clinicianID == "" {
		return false, errors.New("clinician id is missing")
	}
	if id == "" {
		return false, errors.New("prescription id is missing")
	} else if !bson.IsObjectIdHex(id) {
		return false, nil
	}

	now := time.Now()
	logger := log.LoggerFromContext(ctx).WithFields(log.Fields{"clinicianId": clinicianID, "id": id})

	selector := bson.M{
		"_id": bson.ObjectIdHex(id),
		"$or": []bson.M{
			{"prescriberId": clinicianID},
			{"createdUserId": clinicianID},
		},
		"state": bson.M{
			"$in": []string{prescription.StateDraft, prescription.StatePending},
		},
		"deletedTime": nil,
	}
	update := bson.M{
		"$set": bson.M{
			"deletedUserId": clinicianID,
			"deletedTime":   now,
		},
	}

	err := p.C().Update(selector, update)
	logger.WithFields(log.Fields{"duration": time.Since(now) / time.Microsecond}).WithError(err).Debug("DeletePrescription")
	if err == mgo.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "unable to delete prescription")
	} else {
		return true, nil
	}
}

func newMongoSelectorFromFilter(filter *prescription.Filter) bson.M {
	selector := bson.M{}
	if filter.GetClinicianID() != "" {
		selector["$or"] = []bson.M{
			{"prescriberId": filter.GetClinicianID()},
			{"createdUserId": filter.GetClinicianID()},
		}
	}
	if filter.PatientID != "" {
		selector["patientId"] = filter.PatientID
	}
	if filter.PatientEmail != "" {
		selector["latestRevision.attributes.email"] = filter.PatientEmail
	}
	if filter.ID != "" {
		if bson.IsObjectIdHex(filter.ID) {
			selector["_id"] = bson.ObjectIdHex(filter.ID)
		} else {
			selector["_id"] = nil
		}
	}
	if filter.State != "" {
		selector["state"] = filter.State
	}
	if filter.CreatedTimeStart != nil {
		selector["createdTime"] = bson.M{"$gt": filter.CreatedTimeStart}
	}
	if filter.CreatedTimeEnd != nil {
		selector["createdTime"] = bson.M{"$lt": filter.CreatedTimeEnd}
	}
	if filter.ModifiedTimeStart != nil {
		selector["latestRevision.attributes.modifiedTime"] = bson.M{"$gt": filter.ModifiedTimeStart}
	}
	if filter.ModifiedTimeEnd != nil {
		selector["latestRevision.attributes.modifiedTime"] = bson.M{"$lt": filter.ModifiedTimeEnd}
	}

	return selector
}

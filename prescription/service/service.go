package service

import (
	"context"

	"github.com/tidepool-org/platform/user"

	"github.com/tidepool-org/platform/page"

	prescriptionStore "github.com/tidepool-org/platform/prescription/store"

	"github.com/tidepool-org/platform/errors"
	"github.com/tidepool-org/platform/prescription"
)

type PrescriptionService struct {
	prescriptionStore prescriptionStore.Store
}

func NewService(store prescriptionStore.Store) (prescription.Service, error) {
	if store == nil {
		return nil, errors.New("prescription store is missing")
	}

	return &PrescriptionService{
		prescriptionStore: store,
	}, nil
}

func (p *PrescriptionService) CreatePrescription(ctx context.Context, userID string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.CreatePrescription(ctx, userID, create)
}

func (p *PrescriptionService) ListPrescriptions(ctx context.Context, filter *prescription.Filter, pagination *page.Pagination) (prescription.Prescriptions, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ListPrescriptions(ctx, filter, pagination)
}

func (p *PrescriptionService) DeletePrescription(ctx context.Context, clinicianID string, id string) (bool, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.DeletePrescription(ctx, clinicianID, id)
}

func (p *PrescriptionService) AddRevision(ctx context.Context, usr *user.User, id string, create *prescription.RevisionCreate) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.AddRevision(ctx, usr, id, create)
}

func (p *PrescriptionService) ClaimPrescription(ctx context.Context, usr *user.User, claim *prescription.Claim) (*prescription.Prescription, error) {
	repo := p.prescriptionStore.GetPrescriptionRepository()
	return repo.ClaimPrescription(ctx, usr, claim)
}

func (p *PrescriptionService) UpdatePrescriptionState(ctx context.Context, usr *user.User, id string, update *prescription.StateUpdate) (*prescription.Prescription, error) {
	repository := p.prescriptionStore.GetPrescriptionRepository()
	return repository.UpdatePrescriptionState(ctx, usr, id, update)
}

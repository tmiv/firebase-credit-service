package firebasecredit

import (
	"context"
	"fmt"

	fb "firebase.google.com/go"
	"firebase.google.com/go/db"
)

type ChargeData struct {
	Path string `json:"path"`
	Cost int    `json:"cost,omitempty"`
}

type Service struct {
	chargeData ChargeData
	dbURL      string
}

func NewService(chargeData ChargeData, dbURL string) *Service {
	return &Service{
		chargeData: chargeData,
		dbURL:      dbURL,
	}
}

func (s *Service) createDBClient(ctx context.Context) (*db.Client, error) {
	con := fb.Config{
		DatabaseURL: s.dbURL,
	}
	fire, err := fb.NewApp(ctx, &con)
	if err != nil {
		return nil, fmt.Errorf("new firebase app failed %v", err)
	}
	fdb, err := fire.Database(ctx)
	if err != nil {
		return nil, fmt.Errorf("new firebase database failed %v", err)
	}
	return fdb, nil
}

func (s *Service) AddCredits(ctx context.Context, user string, grant int) (int, error) {
	fdb, err := s.createDBClient(ctx)
	if err != nil {
		return -1, err
	}

	var totalValue int = 0
	pathRef := fdb.NewRef(s.chargeData.Path)
	accountRef := pathRef.Child(user)
	err = accountRef.Transaction(ctx, func(tn db.TransactionNode) (interface{}, error) {
		var acc int = 0
		if err := tn.Unmarshal(&acc); err != nil {
			return nil, err
		}

		acc += grant

		totalValue = acc

		// Return the new value which will be written back to the database.
		return acc, nil
	})
	if err != nil {
		return -1, err
	}

	return totalValue, nil
}

func (s *Service) SubtractCredits(ctx context.Context, user string) (bool, int, error) {
	fdb, err := s.createDBClient(ctx)
	if err != nil {
		return false, 0, err
	}

	var totalValue int = 0
	deducted := false
	pathRef := fdb.NewRef(s.chargeData.Path)
	accountRef := pathRef.Child(user)
	err = accountRef.Transaction(ctx, func(tn db.TransactionNode) (interface{}, error) {
		var acc int = 0
		if err := tn.Unmarshal(&acc); err != nil {
			return nil, err
		}

		if acc >= s.chargeData.Cost {
			acc -= s.chargeData.Cost
			totalValue = acc
			deducted = true
		}

		return acc, nil
	})
	return deducted, totalValue, err
}

func (s *Service) RefundCredits(ctx context.Context, user string) error {
	fdb, err := s.createDBClient(ctx)
	if err != nil {
		return err
	}

	pathRef := fdb.NewRef(s.chargeData.Path)
	accountRef := pathRef.Child(user)
	err = accountRef.Transaction(ctx, func(tn db.TransactionNode) (interface{}, error) {
		var acc int = 0
		if err := tn.Unmarshal(&acc); err != nil {
			return nil, err
		}

		acc += s.chargeData.Cost

		return acc, nil
	})
	return err
}

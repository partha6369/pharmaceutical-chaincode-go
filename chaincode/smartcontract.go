/*
SPDX-License-Identifier: Apache-2.0
*/

package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type HistoryRecord struct {
	EventType string `json:"eventType"`
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Location  string `json:"location"`
	Remarks   string `json:"remarks"`
}

type DrugBatch struct {
	BatchID          string          `json:"batchID"`
	GTIN             string          `json:"gtin"`
	Quantity         float64         `json:"quantity"`
	Unit             string          `json:"unit"`
	CurrentOwner     string          `json:"currentOwner"`
	CurrentLocation  string          `json:"currentLocation"`
	Status           string          `json:"status"`
	LastEventTime    string          `json:"lastEventTime"`
	InspectionStatus string          `json:"inspectionStatus"`
	History          []HistoryRecord `json:"history"`
}

func (s *SmartContract) InitLedger(
	ctx contractapi.TransactionContextInterface,
) error {
	return nil
}

func (s *SmartContract) CreateDrugBatch(
	ctx contractapi.TransactionContextInterface,
	batchID string,
	gtin string,
	quantityText string,
	unit string,
	manufacturer string,
	location string,
	timestamp string,
) error {
	if batchID == "" {
		return fmt.Errorf("batch ID must not be empty")
	}

	exists, err := s.DrugBatchExists(ctx, batchID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("drug batch %s already exists", batchID)
	}

	quantity, err := strconv.ParseFloat(quantityText, 64)
	if err != nil {
		return fmt.Errorf(
			"invalid quantity %q for batch %s: %v",
			quantityText,
			batchID,
			err,
		)
	}

	batch := DrugBatch{
		BatchID:          batchID,
		GTIN:             gtin,
		Quantity:         quantity,
		Unit:             unit,
		CurrentOwner:     manufacturer,
		CurrentLocation:  location,
		Status:           "CREATED",
		LastEventTime:    timestamp,
		InspectionStatus: "NOT_INSPECTED",
		History: []HistoryRecord{
			{
				EventType: "CREATE_DRUG_BATCH",
				Timestamp: timestamp,
				Actor:     manufacturer,
				Location:  location,
				Remarks:   "Drug batch created",
			},
		},
	}

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to serialise drug batch %s: %v", batchID, err)
	}

	return ctx.GetStub().PutState(batchID, batchJSON)
}

func (s *SmartContract) TransferCustody(
	ctx contractapi.TransactionContextInterface,
	batchID string,
	newOwner string,
	newLocation string,
	timestamp string,
) error {
	batch, err := s.QueryProvenance(ctx, batchID)
	if err != nil {
		return err
	}

	batch.CurrentOwner = newOwner
	batch.CurrentLocation = newLocation
	batch.Status = "TRANSFERRED"
	batch.LastEventTime = timestamp
	batch.History = append(batch.History, HistoryRecord{
		EventType: "TRANSFER_CUSTODY",
		Timestamp: timestamp,
		Actor:     newOwner,
		Location:  newLocation,
		Remarks:   "Custody transferred",
	})

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to serialise drug batch %s: %v", batchID, err)
	}

	return ctx.GetStub().PutState(batchID, batchJSON)
}

func (s *SmartContract) LogQualityInspection(
	ctx contractapi.TransactionContextInterface,
	batchID string,
	inspector string,
	result string,
	remarks string,
	timestamp string,
) error {
	batch, err := s.QueryProvenance(ctx, batchID)
	if err != nil {
		return err
	}

	batch.InspectionStatus = result
	batch.Status = "INSPECTED"
	batch.LastEventTime = timestamp
	batch.History = append(batch.History, HistoryRecord{
		EventType: "QUALITY_INSPECTION",
		Timestamp: timestamp,
		Actor:     inspector,
		Location:  batch.CurrentLocation,
		Remarks:   remarks,
	})

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to serialise drug batch %s: %v", batchID, err)
	}

	return ctx.GetStub().PutState(batchID, batchJSON)
}

func (s *SmartContract) QueryProvenance(
	ctx contractapi.TransactionContextInterface,
	batchID string,
) (*DrugBatch, error) {
	batchJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return nil, fmt.Errorf("failed to read drug batch %s: %v", batchID, err)
	}
	if batchJSON == nil {
		return nil, fmt.Errorf("drug batch %s does not exist", batchID)
	}

	var batch DrugBatch
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		return nil, fmt.Errorf(
			"failed to deserialise drug batch %s: %v",
			batchID,
			err,
		)
	}

	return &batch, nil
}

func (s *SmartContract) DrugBatchExists(
	ctx contractapi.TransactionContextInterface,
	batchID string,
) (bool, error) {
	batchJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return false, fmt.Errorf(
			"failed to determine whether drug batch %s exists: %v",
			batchID,
			err,
		)
	}

	return batchJSON != nil, nil
}

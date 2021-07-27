package handlers

import (
	"fmt"
	"github.com/go-chi/render"
	zkutils "github.com/iden3/go-iden3-core/utils/zk"
	"github.com/iden3/prover-server/pkg/app/configs"
	"github.com/iden3/prover-server/pkg/app/rest"
	"github.com/iden3/prover-server/pkg/proof"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

type ZKHandler struct {
	ProverConfig configs.ProverConfig
}

type GenerateReq struct {
	CircuitName string         `json:"circuit_name"`
	Inputs      proof.ZKInputs `json:"inputs"`
}

type VerifyReq struct {
	CircuitName string              `json:"circuit_name"`
	ZKP         *zkutils.ZkProofOut `json:"zkp"`
}

type VerifyResp struct {
	Valid bool `json:"valid"`
}

func NewZKHandler(proverConfig configs.ProverConfig) *ZKHandler {
	return &ZKHandler{
		proverConfig,
	}
}

// GenerateProof
// POST /api/v1/proof/generate
func (h *ZKHandler) GenerateProof(w http.ResponseWriter, r *http.Request) {

	var req GenerateReq
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		rest.ErrorJSON(w, r, http.StatusBadRequest, err, "can't bind request", 0)
		return
	}

	circuitPath, err := getValidatedCircuitPath(h.ProverConfig.CircuitsBasePath, req.CircuitName)
	if err != nil {
		rest.ErrorJSON(w, r, http.StatusBadRequest, err, "illegal circuitPath", 0)
		return
	}

	zkProofOut, err := proof.GenerateZkProof(circuitPath, req.Inputs)

	if err != nil {
		rest.ErrorJSON(w, r, http.StatusInternalServerError, err, "can't generate identifier", 0)
		return
	}

	render.JSON(w, r, zkProofOut)
}

// VerifyProof
// POST /api/v1/proof/verify
func (h *ZKHandler) VerifyProof(w http.ResponseWriter, r *http.Request) {

	valid := false

	var req VerifyReq
	if err := render.DecodeJSON(r.Body, &req); err != nil {
		rest.ErrorJSON(w, r, http.StatusBadRequest, err, "can't bind request", 0)
		return
	}

	circuitPath, err := getValidatedCircuitPath(h.ProverConfig.CircuitsBasePath, req.CircuitName)
	if err != nil {
		rest.ErrorJSON(w, r, http.StatusBadRequest, err, "illegal circuitPath", 0)
		return
	}

	err = proof.VerifyZkProof(circuitPath, req.ZKP)
	if err == nil {
		valid = true
	}

	render.JSON(w, r, VerifyResp{Valid: valid})
}

func getValidatedCircuitPath(circuitBasePath string, circuitName string) (circuitPath string, err error) {
	// TODO: validate circuitName for illegal characters, etc

	circuitPath = circuitBasePath + "/" + circuitName
	log.Debugf("circuitPath: %s\n", filepath.Clean(circuitPath))

	if filepath.Clean(circuitPath) != circuitPath {
		return "", fmt.Errorf("illegal circuitPath")
	}

	info, err := os.Stat(circuitPath)
	fmt.Printf("%+v %v\n", info, err)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("circuitPath doesn't exist")
	}

	return circuitPath, nil
}
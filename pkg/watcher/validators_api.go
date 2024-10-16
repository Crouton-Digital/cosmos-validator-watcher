package watcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/kilnfi/cosmos-validator-watcher/pkg/metrics"
	"github.com/rs/zerolog/log"
)

type ValidatorsAPIWatcher struct {
	metrics    *metrics.Metrics
	validators []TrackedValidator
	api        *http.Client
}

func NewValidatorsAPIWatcher(validators []TrackedValidator, metrics *metrics.Metrics, api *http.Client) *ValidatorsAPIWatcher {
	return &ValidatorsAPIWatcher{
		metrics:    metrics,
		validators: validators,
		api:        api,
	}
}

func (w *ValidatorsAPIWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)

	for {
		for _, validator := range w.validators {
			log.Debug().Str("validator", validator.Account).Msg("fetching validator from api")
			if err := w.fetchInfo(ctx, validator); err != nil {
				log.Error().Err(err).Str("validator", validator.Account).Msg("failed to fetch info")
			}
		}
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *ValidatorsAPIWatcher) fetchInfo(ctx context.Context, validator TrackedValidator) error {
	// fetch all info in parallel
	errCh := make(chan error, 3)

	go func() {
		errCh <- w.fetchAccount(ctx, validator)
	}()

	go func() {
		errCh <- w.fetchValidator(ctx, validator)
	}()

	go func() {
		errCh <- w.fetchDelegator(ctx, validator)
	}()

	for i := 0; i < 3; i++ {
		if err := <-errCh; err != nil {
			log.Error().Err(err).Str("validator", validator.Account).Msg("failed to fetch info")

			return err
		}
	}

	return nil
}

// response
type APIAccountResponse struct {
	Address string      `json:"address"`
	Balance APIBalance  `json:"balance"`
	Assets  []APIAssets `json:"assets"`
}
type APIBalance struct {
	Available  int `json:"available"`
	Vesting    int `json:"vesting"`
	Delegated  int `json:"delegated"`
	Unbonding  int `json:"unbonding"`
	Reward     int `json:"reward"`
	Commission int `json:"commission"`
}
type APIAssets struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

func (w *ValidatorsAPIWatcher) fetchAccount(ctx context.Context, validator TrackedValidator) error {
	// fetch account info from api
	requestURL := "https://api.testnet.storyscan.app/accounts/" + validator.Account

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := w.api.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Str("request_url", requestURL).Int("status_code", resp.StatusCode).Msg("failed to fetch account info")

		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse response
	var account APIAccountResponse
	if err := json.Unmarshal(body, &account); err != nil {
		return err
	}

	log.Debug().Interface("account", account).Str("validator", validator.Account).Msg("fetched account info")

	// update metrics
	w.metrics.ValidatorBalanceAvailable.WithLabelValues(validator.Address, validator.Name).Set(float64(account.Balance.Available))
	w.metrics.ValidatorCommission.WithLabelValues(validator.Address, validator.Name).Set(float64(account.Balance.Commission))
	w.metrics.ValidatorBalanceDelegated.WithLabelValues(validator.Address, validator.Name).Set(float64(account.Balance.Delegated))
	w.metrics.ValidatorBalanceReward.WithLabelValues(validator.Address, validator.Name).Set(float64(account.Balance.Reward))
	w.metrics.ValidatorBalanceUnBonding.WithLabelValues(validator.Address, validator.Name).Set(float64(account.Balance.Unbonding))

	return nil
}

type APIValidator struct {
	Status             int           `json:"status"`
	Tokens             int64         `json:"tokens"`
	DelegatorShares    string        `json:"delegator_shares"`
	UnbondingTime      time.Time     `json:"unbonding_time"`
	Commission         Commission    `json:"commission"`
	MinSelfDelegation  string        `json:"min_self_delegation"`
	Participation      Participation `json:"participation"`
	SigningInfo        SigningInfo   `json:"signingInfo"`
	Uptime             Uptime        `json:"uptime"`
	VotingPowerPercent float64       `json:"votingPowerPercent"`
	CumulativeShare    float64       `json:"cumulativeShare"`
}
type CommissionRates struct {
	Rate          string `json:"rate"`
	MaxRate       string `json:"max_rate"`
	MaxChangeRate string `json:"max_change_rate"`
}
type Commission struct {
	CommissionRates CommissionRates `json:"commission_rates"`
	UpdateTime      time.Time       `json:"update_time"`
}
type Participation struct {
	Rate  int `json:"rate"`
	Total int `json:"total"`
	Voted int `json:"voted"`
}
type SigningInfo struct {
	BondedHeight int    `json:"bondedHeight"`
	JailedUntil  string `json:"jailedUntil"`
	Tombstoned   bool   `json:"tombstoned"`
}
type HistoricalUptime struct {
	EarliestHeight int `json:"earliestHeight"`
	LastSyncHeight int `json:"lastSyncHeight"`
	SuccessBlocks  int `json:"successBlocks"`
}
type WindowUptime struct {
	Uptime      float64 `json:"uptime"`
	WindowStart int     `json:"windowStart"`
	WindowEnd   int     `json:"windowEnd"`
}
type Uptime struct {
	HistoricalUptime HistoricalUptime `json:"historicalUptime"`
	WindowUptime     WindowUptime     `json:"windowUptime"`
}

func (w *ValidatorsAPIWatcher) fetchValidator(ctx context.Context, validator TrackedValidator) error {
	// fetch account info from api
	requestURL := "https://api.testnet.storyscan.app/validators/" + validator.OperatorAddress

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := w.api.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Str("request_url", requestURL).Int("status_code", resp.StatusCode).Msg("failed to fetch account info")
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse response
	var apiValidator APIValidator
	if err := json.Unmarshal(body, &apiValidator); err != nil {
		return err
	}

	log.Debug().Interface("validator", apiValidator).Str("validator", validator.Account).Msg("fetched validator info")

	// update metrics
	w.metrics.ValidatorStatus.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Status))
	w.metrics.ValidatorTokens.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Tokens))
	w.metrics.ValidatorCommissionRate.WithLabelValues(validator.Address, validator.Name).Set(MustFloat64(apiValidator.Commission.CommissionRates.Rate))
	w.metrics.ValidatorDelegatorShares.WithLabelValues(validator.Address, validator.Name).Set(MustFloat64(apiValidator.DelegatorShares))
	w.metrics.ValidatorUnbondingTime.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.UnbondingTime.Unix()))
	w.metrics.ValidatorMinSelfDelegation.WithLabelValues(validator.Address, validator.Name).Set(MustFloat64(apiValidator.MinSelfDelegation))
	w.metrics.ValidatorParticipationRate.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Participation.Rate))
	w.metrics.ValidatorParticipationTotal.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Participation.Total))
	w.metrics.ValidatorParticipationVoted.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Participation.Voted))
	w.metrics.ValidatorSigningInfoBondedHeight.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.SigningInfo.BondedHeight))
	w.metrics.ValidatorSigningInfoTombstoned.WithLabelValues(validator.Address, validator.Name).Set(metrics.BoolToFloat64(apiValidator.SigningInfo.Tombstoned))
	w.metrics.ValidatorUptimeHistoricalEarliestHeight.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Uptime.HistoricalUptime.EarliestHeight))
	w.metrics.ValidatorUptimeHistoricalLastSyncHeight.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Uptime.HistoricalUptime.LastSyncHeight))
	w.metrics.ValidatorUptimeHistoricalSuccessBlocks.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Uptime.HistoricalUptime.SuccessBlocks))
	w.metrics.ValidatorUptimeWindowUptime.WithLabelValues(validator.Address, validator.Name).Set(apiValidator.Uptime.WindowUptime.Uptime)
	w.metrics.ValidatorUptimeWindowStart.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Uptime.WindowUptime.WindowStart))
	w.metrics.ValidatorUptimeWindowEnd.WithLabelValues(validator.Address, validator.Name).Set(float64(apiValidator.Uptime.WindowUptime.WindowEnd))
	w.metrics.ValidatorVotingPowerPercent.WithLabelValues(validator.Address, validator.Name).Set(apiValidator.VotingPowerPercent)
	w.metrics.ValidatorCumulativeShare.WithLabelValues(validator.Address, validator.Name).Set(apiValidator.CumulativeShare)

	return nil
}

type APIDelegatorsResponse struct {
	ValidatorDelegators int `json:"validatorDelegators"`
}

func (w *ValidatorsAPIWatcher) fetchDelegator(ctx context.Context, validator TrackedValidator) error {
	// fetch account info from api
	requestURL := "https://api.testnet.storyscan.app/validators/" + validator.OperatorAddress + "/delegators"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}

	resp, err := w.api.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Str("request_url", requestURL).Int("status_code", resp.StatusCode).Msg("failed to fetch delegator info")
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// parse response
	var apiResponse APIDelegatorsResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return err
	}

	// log.Debug().Str("validator", validator.Account).Int("delegators", apiResponse.ValidatorDelegators).Msg("fetched delegator info")

	// update metrics
	w.metrics.ValidatorDelegators.WithLabelValues(validator.Address, validator.Name).Set(float64(apiResponse.ValidatorDelegators))

	return nil
}

func MustFloat64(val string) float64 {
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return f
}

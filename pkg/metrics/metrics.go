package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Metrics struct {
	Registry *prometheus.Registry

	// Global metrics
	ActiveSet       *prometheus.GaugeVec
	BlockHeight     *prometheus.GaugeVec
	ProposalEndTime *prometheus.GaugeVec
	SeatPrice       *prometheus.GaugeVec
	SkippedBlocks   *prometheus.CounterVec
	TrackedBlocks   *prometheus.CounterVec
	Transactions    *prometheus.CounterVec
	UpgradePlan     *prometheus.GaugeVec

	// Validator metrics
	Rank                    *prometheus.GaugeVec
	ProposedBlocks          *prometheus.CounterVec
	ValidatedBlocks         *prometheus.CounterVec
	MissedBlocks            *prometheus.CounterVec
	SoloMissedBlocks        *prometheus.CounterVec
	ConsecutiveMissedBlocks *prometheus.GaugeVec
	Tokens                  *prometheus.GaugeVec
	IsBonded                *prometheus.GaugeVec
	IsJailed                *prometheus.GaugeVec
	Commission              *prometheus.GaugeVec
	Vote                    *prometheus.GaugeVec
	LastValidatedBlockTime  *prometheus.GaugeVec

	// Metrics from the validator API watcher
	ValidatorBalanceAvailable               *prometheus.GaugeVec
	ValidatorCommission                     *prometheus.GaugeVec
	ValidatorBalanceDelegated               *prometheus.GaugeVec
	ValidatorBalanceReward                  *prometheus.GaugeVec
	ValidatorBalanceUnBonding               *prometheus.GaugeVec
	ValidatorDelegators                     *prometheus.GaugeVec
	ValidatorStatus                         *prometheus.GaugeVec
	ValidatorTokens                         *prometheus.GaugeVec
	ValidatorCommissionRate                 *prometheus.GaugeVec
	ValidatorDelegatorShares                *prometheus.GaugeVec
	ValidatorUnbondingTime                  *prometheus.GaugeVec
	ValidatorMinSelfDelegation              *prometheus.GaugeVec
	ValidatorParticipationRate              *prometheus.GaugeVec
	ValidatorSigningInfo                    *prometheus.GaugeVec
	ValidatorUptime                         *prometheus.GaugeVec
	ValidatorVotingPowerPercent             *prometheus.GaugeVec
	ValidatorCumulativeShare                *prometheus.GaugeVec
	ValidatorParticipationTotal             *prometheus.GaugeVec
	ValidatorParticipationVoted             *prometheus.GaugeVec
	ValidatorSigningInfoBondedHeight        *prometheus.GaugeVec
	ValidatorSigningInfoTombstoned          *prometheus.GaugeVec
	ValidatorUptimeHistoricalEarliestHeight *prometheus.GaugeVec
	ValidatorUptimeHistoricalLastSyncHeight *prometheus.GaugeVec
	ValidatorUptimeHistoricalSuccessBlocks  *prometheus.GaugeVec
	ValidatorUptimeWindowUptime             *prometheus.GaugeVec
	ValidatorUptimeWindowStart              *prometheus.GaugeVec
	ValidatorUptimeWindowEnd                *prometheus.GaugeVec

	// Node metrics
	NodeBlockHeight *prometheus.GaugeVec
	NodeSynced      *prometheus.GaugeVec
}

func New(namespace string) *Metrics {
	metrics := &Metrics{
		Registry: prometheus.NewRegistry(),
		// LastSignedBlockTimestamp: prometheus.NewGaugeVec(
		// 	prometheus.GaugeOpts{
		// 		Namespace: namespace,
		// 		Name:      "last_signed_block_timestamp",
		// 		Help:      "Timestamp of the last signed block",
		// 	},
		// 	[]string{"chain_id"},
		// ),
		BlockHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "block_height",
				Help:      "Latest known block height (all nodes mixed up)",
			},
			[]string{"chain_id"},
		),
		ActiveSet: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_set",
				Help:      "Number of validators in the active set",
			},
			[]string{"chain_id"},
		),
		SeatPrice: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "seat_price",
				Help:      "Min seat price to be in the active set (ie. bonded tokens of the latest validator)",
			},
			[]string{"chain_id", "denom"},
		),
		Rank: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "rank",
				Help:      "Rank of the validator",
			},
			[]string{"chain_id", "address", "name"},
		),
		ProposedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "proposed_blocks",
				Help:      "Number of proposed blocks per validator (for a bonded validator)",
			},
			[]string{"chain_id", "address", "name"},
		),
		ValidatedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "validated_blocks",
				Help:      "Number of validated blocks per validator (for a bonded validator)",
			},
			[]string{"chain_id", "address", "name"},
		),
		MissedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "missed_blocks",
				Help:      "Number of missed blocks per validator (for a bonded validator)",
			},
			[]string{"chain_id", "address", "name"},
		),
		SoloMissedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "solo_missed_blocks",
				Help:      "Number of missed blocks per validator, unless block is missed by many other validators",
			},
			[]string{"chain_id", "address", "name"},
		),
		ConsecutiveMissedBlocks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "consecutive_missed_blocks",
				Help:      "Number of consecutive missed blocks per validator (for a bonded validator)",
			},
			[]string{"chain_id", "address", "name"},
		),
		TrackedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tracked_blocks",
				Help:      "Number of blocks tracked since start",
			},
			[]string{"chain_id"},
		),
		Transactions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "transactions_total",
				Help:      "Number of transactions since start",
			},
			[]string{"chain_id"},
		),
		SkippedBlocks: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "skipped_blocks",
				Help:      "Number of blocks skipped (ie. not tracked) since start",
			},
			[]string{"chain_id"},
		),
		Tokens: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "tokens",
				Help:      "Number of staked tokens per validator",
			},
			[]string{"chain_id", "address", "name", "denom"},
		),
		IsBonded: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "is_bonded",
				Help:      "Set to 1 if the validator is bonded",
			},
			[]string{"chain_id", "address", "name"},
		),
		IsJailed: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "is_jailed",
				Help:      "Set to 1 if the validator is jailed",
			},
			[]string{"chain_id", "address", "name"},
		),
		Commission: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "commission",
				Help:      "Earned validator commission",
			},
			[]string{"chain_id", "address", "name", "denom"},
		),
		Vote: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "vote",
				Help:      "Set to 1 if the validator has voted on a proposal",
			},
			[]string{"chain_id", "address", "name", "proposal_id"},
		),
		NodeBlockHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "node_block_height",
				Help:      "Latest fetched block height for each node",
			},
			[]string{"chain_id", "node"},
		),
		NodeSynced: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "node_synced",
				Help:      "Set to 1 is the node is synced (ie. not catching-up)",
			},
			[]string{"chain_id", "node"},
		),
		UpgradePlan: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "upgrade_plan",
				Help:      "Block height of the upcoming upgrade (hard fork)",
			},
			[]string{"chain_id", "version"},
		),
		ProposalEndTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "proposal_end_time",
				Help:      "Timestamp of the voting end time of a proposal",
			},
			[]string{"chain_id", "proposal_id"},
		),
		LastValidatedBlockTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_validated_block_time",
				Help:      "Timestamp of the last validated block",
			},
			[]string{"chain_id", "address", "name"},
		),

		// Metrics from the validator API watcher
		ValidatorBalanceAvailable: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_balance_available",
				Help:      "Validator balance available",
			},
			[]string{"address"},
		),
		ValidatorCommission: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_balance_commission",
				Help:      "Validator commission",
			},
			[]string{"address"},
		),
		ValidatorBalanceDelegated: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_balance_delegated",
				Help:      "Validator balance delegated",
			},
			[]string{"address"},
		),
		ValidatorBalanceReward: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_balance_reward",
				Help:      "Validator balance reward",
			},
			[]string{"address"},
		),
		ValidatorBalanceUnBonding: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_balance_unbonding",
				Help:      "Validator balance unbonding",
			},
			[]string{"address"},
		),
		ValidatorDelegators: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_delegators",
				Help:      "Validator delegators",
			},
			[]string{"address"},
		),
		ValidatorStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_status",
				Help:      "Validator status",
			},
			[]string{"address"},
		),
		ValidatorTokens: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_tokens",
				Help:      "Validator tokens",
			},
			[]string{"address"},
		),
		ValidatorCommissionRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_commission_rate",
				Help:      "Validator commission rate",
			},
			[]string{"address"},
		),
		ValidatorDelegatorShares: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_delegator_shares",
				Help:      "Validator delegator shares",
			},
			[]string{"address"},
		),
		ValidatorUnbondingTime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_unbonding_time",
				Help:      "Validator unbonding time",
			},
			[]string{"address"},
		),
		ValidatorMinSelfDelegation: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_min_self_delegation",
				Help:      "Validator min self delegation",
			},
			[]string{"address"},
		),
		ValidatorParticipationRate: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_participation_rate",
				Help:      "Validator participation rate",
			},
			[]string{"address"},
		),
		ValidatorParticipationTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_participation_total",
				Help:      "Validator participation total",
			},
			[]string{"address"},
		),
		ValidatorSigningInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_signing_info",
				Help:      "Validator signing info",
			},
			[]string{"address"},
		),
		ValidatorUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime",
				Help:      "Validator uptime",
			},
			[]string{"address"},
		),
		ValidatorVotingPowerPercent: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_voting_power_percent",
				Help:      "Validator voting power percent",
			},
			[]string{"address"},
		),
		ValidatorCumulativeShare: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_cumulative_share",
				Help:      "Validator cumulative share",
			},
			[]string{"address"},
		),
		ValidatorParticipationVoted: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_participation_voted",
				Help:      "Validator participation voted",
			},
			[]string{"address"},
		),
		ValidatorSigningInfoBondedHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_signing_info_bonded_height",
				Help:      "Validator signing info bonded height",
			},
			[]string{"address"},
		),
		ValidatorSigningInfoTombstoned: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_signing_info_tombstoned",
				Help:      "Validator signing info tombstoned",
			},
			[]string{"address"},
		),
		ValidatorUptimeHistoricalEarliestHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_historical_earliest_height",
				Help:      "Validator uptime historical earliest height",
			},
			[]string{"address"},
		),
		ValidatorUptimeHistoricalLastSyncHeight: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_historical_last_sync_height",
				Help:      "Validator uptime historical last sync height",
			},
			[]string{"address"},
		),
		ValidatorUptimeHistoricalSuccessBlocks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_historical_success_blocks",
				Help:      "Validator uptime historical success blocks",
			},
			[]string{"address"},
		),
		ValidatorUptimeWindowUptime: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_window_uptime",
				Help:      "Validator uptime window uptime",
			},
			[]string{"address"},
		),
		ValidatorUptimeWindowStart: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_window_start",
				Help:      "Validator uptime window start",
			},
			[]string{"address"},
		),
		ValidatorUptimeWindowEnd: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "validator_uptime_window_end",
				Help:      "Validator uptime window end",
			},
			[]string{"address"},
		),
	}

	return metrics
}

func (m *Metrics) Register() {
	m.Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	m.Registry.MustRegister(collectors.NewGoCollector())

	m.Registry.MustRegister(m.BlockHeight)
	m.Registry.MustRegister(m.ActiveSet)
	m.Registry.MustRegister(m.SeatPrice)
	m.Registry.MustRegister(m.Rank)
	m.Registry.MustRegister(m.ProposedBlocks)
	m.Registry.MustRegister(m.ValidatedBlocks)
	m.Registry.MustRegister(m.MissedBlocks)
	m.Registry.MustRegister(m.SoloMissedBlocks)
	m.Registry.MustRegister(m.ConsecutiveMissedBlocks)
	m.Registry.MustRegister(m.TrackedBlocks)
	m.Registry.MustRegister(m.Transactions)
	m.Registry.MustRegister(m.SkippedBlocks)
	m.Registry.MustRegister(m.Tokens)
	m.Registry.MustRegister(m.IsBonded)
	m.Registry.MustRegister(m.Commission)
	m.Registry.MustRegister(m.IsJailed)
	m.Registry.MustRegister(m.Vote)
	m.Registry.MustRegister(m.NodeBlockHeight)
	m.Registry.MustRegister(m.NodeSynced)
	m.Registry.MustRegister(m.UpgradePlan)
	m.Registry.MustRegister(m.ProposalEndTime)
	m.Registry.MustRegister(m.LastValidatedBlockTime)

	m.Registry.MustRegister(m.ValidatorBalanceAvailable)
	m.Registry.MustRegister(m.ValidatorCommission)
	m.Registry.MustRegister(m.ValidatorBalanceDelegated)
	m.Registry.MustRegister(m.ValidatorBalanceReward)
	m.Registry.MustRegister(m.ValidatorBalanceUnBonding)
	m.Registry.MustRegister(m.ValidatorDelegators)
	m.Registry.MustRegister(m.ValidatorStatus)
	m.Registry.MustRegister(m.ValidatorTokens)
	m.Registry.MustRegister(m.ValidatorCommissionRate)
	m.Registry.MustRegister(m.ValidatorDelegatorShares)
	m.Registry.MustRegister(m.ValidatorUnbondingTime)
	m.Registry.MustRegister(m.ValidatorMinSelfDelegation)
	m.Registry.MustRegister(m.ValidatorParticipationRate)
	m.Registry.MustRegister(m.ValidatorSigningInfo)
	m.Registry.MustRegister(m.ValidatorUptime)
	m.Registry.MustRegister(m.ValidatorVotingPowerPercent)
	m.Registry.MustRegister(m.ValidatorCumulativeShare)
	m.Registry.MustRegister(m.ValidatorParticipationTotal)
	m.Registry.MustRegister(m.ValidatorParticipationVoted)
	m.Registry.MustRegister(m.ValidatorSigningInfoBondedHeight)
	m.Registry.MustRegister(m.ValidatorSigningInfoTombstoned)
	m.Registry.MustRegister(m.ValidatorUptimeHistoricalEarliestHeight)
	m.Registry.MustRegister(m.ValidatorUptimeHistoricalLastSyncHeight)
	m.Registry.MustRegister(m.ValidatorUptimeHistoricalSuccessBlocks)
	m.Registry.MustRegister(m.ValidatorUptimeWindowUptime)
	m.Registry.MustRegister(m.ValidatorUptimeWindowStart)
	m.Registry.MustRegister(m.ValidatorUptimeWindowEnd)

}

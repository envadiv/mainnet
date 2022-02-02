package main

import (
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// Account is an internal representation of a genesis passage account
type Account struct {
	Address       sdk.AccAddress
	TotalRegen    Dec
	Distributions []Distribution
}

// Distribution is an internal representation of a genesis vesting distribution of passage
type Distribution struct {
	Time  time.Time
	Regen Dec
}

func (a Account) String() string {
	return fmt.Sprintf("Account{%s, %spassage, %s}", a.Address, a.TotalRegen.String(), a.Distributions)
}

func (d Distribution) String() string {
	return fmt.Sprintf("Distribution{%s, %spassage}", d.Time.Format(time.RFC3339), d.Regen.String())
}

func (acc Account) Validate() error {
	if acc.Address.Empty() {
		return fmt.Errorf("empty address")
	}

	if !acc.TotalRegen.IsPositive() {
		return fmt.Errorf("expected positive balance, got %s", acc.TotalRegen.String())
	}

	var calcTotal Dec
	for _, dist := range acc.Distributions {
		err := dist.Validate()
		if err != nil {
			return err
		}

		calcTotal, err = calcTotal.Add(dist.Regen)
		if err != nil {
			return err
		}
	}

	if !acc.TotalRegen.IsEqual(calcTotal) {
		return fmt.Errorf("incorrect balance, expected %s, got %s", acc.TotalRegen.String(), calcTotal.String())
	}

	return nil
}

func (d Distribution) Validate() error {
	if d.Time.IsZero() {
		return fmt.Errorf("time is zero")
	}
	return nil
}

func RecordToAccount(rec Record, genesisTime time.Time) (Account, error) {
	amount := rec.TotalAmount
	distTime := rec.StartTime
	if distTime.IsZero() {
		return Account{}, fmt.Errorf("require a non-zero distribution time")
	}
	numDist := rec.NumWeeklyDistributions
	if numDist < 1 {
		return Account{}, fmt.Errorf("numDist must be >= 1, got %d", numDist)
	}

	if numDist == 1 && !distTime.After(genesisTime) {
		return Account{
			Address:    rec.Address,
			TotalRegen: amount,
			Distributions: []Distribution{
				{
					Time:  genesisTime,
					Regen: amount,
				},
			},
		}, nil
	}

	// calculate dust, which represents an upasg-integral remainder, represented as a decimal value of passage
	// from dividing `amount` by `numDist`
	distAmount, dust, err := distAmountAndDust(amount, numDist)
	if err != nil {
		return Account{}, err
	}

	var distributions []Distribution
	var genesisAmount Dec

	// collapse all pre-genesis distributions into a genesis distribution
	for ; numDist > 0 && !distTime.After(genesisTime); numDist-- {
		genesisAmount, err = genesisAmount.Add(distAmount)
		if err != nil {
			return Account{}, err
		}
		distTime = distTime.Add(OneWeek)
	}

	// if there is a genesis distribution add it
	if !genesisAmount.IsZero() {
		distributions = append(distributions, Distribution{
			Time:  genesisTime,
			Regen: genesisAmount,
		})
	}

	// add post genesis distributions
	for ; numDist > 0; numDist-- {
		distributions = append(distributions, Distribution{
			Time:  distTime,
			Regen: distAmount,
		})
		// adding one week
		distTime = distTime.Add(OneWeek)
	}

	// add dust to first distribution
	distributions[0].Regen, err = distributions[0].Regen.Add(dust)
	if err != nil {
		return Account{}, err
	}

	return Account{
		Address:       rec.Address,
		Distributions: distributions,
		TotalRegen:    amount,
	}, nil
}

func distAmountAndDust(amount Dec, numDist int) (distAmount Dec, dust Dec, err error) {
	if numDist < 1 {
		return Dec{}, Dec{}, fmt.Errorf("num must be >= 1, got %d", numDist)
	}

	if numDist == 1 {
		return amount, dust, nil
	}

	numDistDec := NewDecFromInt64(int64(numDist))

	// convert amount from passage to upasg, so we can perform integral arithmetic on upasg
	amount, err = amount.Mul(tenE6)
	if err != nil {
		return distAmount, dust, err
	}

	// each distribution is an integral amount of upasg
	distAmount, err = amount.QuoInteger(numDistDec)
	if err != nil {
		return distAmount, dust, err
	}

	dust, err = amount.Rem(numDistDec)
	if err != nil {
		return distAmount, dust, err
	}

	// convert distAmount from upasg back to passage
	distAmount, err = distAmount.Quo(tenE6)
	if err != nil {
		return distAmount, dust, err
	}

	// convert dust from upasg back to passage
	dust, err = dust.Quo(tenE6)
	if err != nil {
		return distAmount, dust, err
	}

	return distAmount, dust, nil
}

const (
	URegenDenom = "upasg"
)

var tenE6 = NewDecFromInt64(1000000)

func ToCosmosAccount(acc Account, genesisTime time.Time) (auth.AccountI, *bank.Balance, error) {
	err := acc.Validate()
	if err != nil {
		return nil, nil, err
	}

	totalCoins, err := RegenToCoins(acc.TotalRegen)
	if err != nil {
		return nil, nil, err
	}

	addrStr := acc.Address.String()
	balance := &bank.Balance{
		Address: addrStr,
		Coins:   totalCoins,
	}

	if len(acc.Distributions) == 0 {
		return &auth.BaseAccount{Address: addrStr}, balance, nil
	}

	startTime := acc.Distributions[0].Time

	// if we have one distribution and it happens before or at genesis return a basic BaseAccount
	if len(acc.Distributions) == 1 {
		if !acc.Distributions[0].Time.After(genesisTime) {
			return &auth.BaseAccount{Address: addrStr}, balance, nil
		} else {
			return &vesting.DelayedVestingAccount{
				BaseVestingAccount: &vesting.BaseVestingAccount{
					BaseAccount:     &auth.BaseAccount{Address: addrStr},
					OriginalVesting: totalCoins,
					EndTime:         acc.Distributions[0].Time.Unix(),
				},
			}, balance, nil
		}
	} else {
		periodStart := startTime

		var periods []vesting.Period
		for _, dist := range acc.Distributions {
			coins, err := RegenToCoins(dist.Regen)
			if err != nil {
				return nil, nil, err
			}

			length := dist.Time.Sub(periodStart)
			periodStart = dist.Time
			seconds := int64(math.Floor(length.Seconds()))
			periods = append(periods, vesting.Period{
				Length: seconds,
				Amount: coins,
			})
		}

		return &vesting.PeriodicVestingAccount{
			BaseVestingAccount: &vesting.BaseVestingAccount{
				BaseAccount: &auth.BaseAccount{
					Address: addrStr,
				},
				OriginalVesting: totalCoins,
				EndTime:         periodStart.Unix(),
			},
			StartTime:      startTime.Unix(),
			VestingPeriods: periods,
		}, balance, nil
	}
}

func RegenToCoins(passageAmount Dec) (sdk.Coins, error) {
	upasg, err := passageAmount.Mul(tenE6)
	if err != nil {
		return nil, err
	}

	upasgInt64, err := upasg.Int64()
	if err != nil {
		return nil, err
	}

	return sdk.NewCoins(sdk.NewCoin(URegenDenom, sdk.NewInt(upasgInt64))), nil
}

func ValidateVestingAccount(acc auth.AccountI) error {
	vacc, ok := acc.(*vesting.PeriodicVestingAccount)
	if !ok {
		return nil
	}

	orig := vacc.OriginalVesting
	time := vacc.StartTime
	var total sdk.Coins
	for _, period := range vacc.VestingPeriods {
		total = total.Add(period.Amount...)
		time += period.Length
	}

	if !orig.IsEqual(total) {
		return fmt.Errorf("vesting account error: expected %s coins, got %s", orig.String(), total.String())
	}

	if vacc.EndTime != time {
		return fmt.Errorf("vesting account error: expected %d end time, got %d", vacc.EndTime, time)
	}

	return nil
}

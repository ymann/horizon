package config

// AnonymousUserRestrictions holds limitations for anonymous users
type AnonymousUserRestrictions struct {
	MaxDailyOutcome   int64
	MaxMonthlyOutcome int64
	MaxAnnualOutcome  int64
	MaxBalance        int64
}

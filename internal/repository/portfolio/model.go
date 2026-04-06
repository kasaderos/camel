package portfolio

type portfolio struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`

	Weights map[string]float64 `db:"weights"`
	Cache   float64            `db:"cache"`

	PolicyID string `db:"policy_id"`
}

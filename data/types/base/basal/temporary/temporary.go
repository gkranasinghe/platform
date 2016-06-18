package temporary

/* CHECKLIST
 * [ ] Uses interfaces as appropriate
 * [ ] Private package variables use underscore prefix
 * [ ] All parameters validated
 * [ ] All errors handled
 * [ ] Reviewed for concurrency safety
 * [ ] Code complete
 * [ ] Full test coverage
 */

import (
	"github.com/tidepool-org/platform/data"
	"github.com/tidepool-org/platform/data/types/base/basal"
)

type Temporary struct {
	basal.Basal `bson:",inline"`

	Duration *int     `json:"duration,omitempty" bson:"duration,omitempty"`
	Rate     *float64 `json:"rate,omitempty" bson:"rate,omitempty"`
	Percent  *float64 `json:"percent,omitempty" bson:"percent,omitempty"`
}

func DeliveryType() string {
	return "temp"
}

func NewDatum() data.Datum {
	return New()
}

func New() *Temporary {
	return &Temporary{}
}

func Init() *Temporary {
	temporary := New()
	temporary.Init()
	return temporary
}

func (t *Temporary) Init() {
	t.Basal.Init()
	t.Basal.DeliveryType = DeliveryType()

	t.Duration = nil
	t.Rate = nil
	t.Percent = nil
}

func (t *Temporary) Parse(parser data.ObjectParser) error {
	if err := t.Basal.Parse(parser); err != nil {
		return err
	}

	t.Duration = parser.ParseInteger("duration")
	t.Rate = parser.ParseFloat("rate")
	t.Percent = parser.ParseFloat("percent")

	return nil
}

func (t *Temporary) Validate(validator data.Validator) error {
	if err := t.Basal.Validate(validator); err != nil {
		return err
	}

	validator.ValidateInteger("duration", t.Duration).Exists().InRange(0, 86400000)
	validator.ValidateFloat("rate", t.Rate).Exists().InRange(0.0, 20.0)
	validator.ValidateFloat("percent", t.Percent).InRange(0.0, 10.0)

	return nil
}

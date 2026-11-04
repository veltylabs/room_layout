package roomlayout

import (
	"webtyp.com/fmt"
	"webtyp.com/time"
)

type CategoryReader interface {
	CategoryOptions(tenantID string) ([]fmt.KeyValue, error)
}

type OccupantReader interface {
	OccupantOptions(tenantID string) ([]fmt.KeyValue, error)
}

// HoursReader es la parte de business_calendar que este módulo usa: si la clínica abre ese día y entre qué minutos.
// Ejemplo: lim, err := h.DayLimits(date)
type HoursReader interface {
	DayLimits(date time.Unix) (time.DayLimits, error)
}

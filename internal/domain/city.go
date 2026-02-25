package domain

type City struct {
	ID        int64
	Name      string
	Slug      string
	Latitude  float64
	Longitude float64
}

func (c *City) Validate() error {
	if c.Name == "" {
		return ErrInvalidInput
	}
	if c.Slug == "" {
		return ErrInvalidInput
	}
	return nil
}

package models

type Auto struct {
	ID            string
	URL           string
	Brand         string
	Model         string
	Price         string
	PriceMark     string
	Generation    string
	Complectation string
	Mileage       string
	NoMileageRF   string
	Color         string
	BodyType      string
	Power         string
	FuelType      string
	EngineVolume  string
	Photos        []string
}

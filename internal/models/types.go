package models

// Auto содержит информацию об автомобиле, включая идентификатор, URL,
// характеристики (марка, модель, цена, комплектация, пробег, цвет, мощность, топливо, объём двигателя)
// и ссылки на фотографии.
type Auto struct {
	ID            string   // Уникальный идентификатор автомобиля
	URL           string   // Ссылка на страницу автомобиля
	Brand         string   // Марка
	Model         string   // Модель
	Price         string   // Цена
	PriceMark     string   // Дополнительная информация о цене
	Generation    string   // Поколение
	Complectation string   // Комплектация
	Mileage       string   // Пробег
	NoMileageRF   string   // Пробег отсутствует для РФ
	Color         string   // Цвет
	BodyType      string   // Тип кузова
	Power         string   // Мощность двигателя
	FuelType      string   // Тип топлива
	EngineVolume  string   // Объём двигателя
	Photos        []string // URL фотографий
}

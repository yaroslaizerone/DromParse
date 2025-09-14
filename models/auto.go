package models

// Auto представляет объявление об автомобиле, включая его характеристики,
// технические данные, цену и фотографии.
type Auto struct {
	ID            string   // Уникальный идентификатор объявления
	URL           string   // Ссылка на объявление
	Brand         string   // Марка автомобиля
	Model         string   // Модель автомобиля
	Price         string   // Стоимость автомобиля
	PriceMark     string   // Тип цены (например, "нормальная цена" или "хорошая цена")
	Generation    string   // Поколение модели
	Complectation string   // Комплектация автомобиля
	Mileage       string   // Пробег автомобиля (км)
	NoMileageRF   string   // Флаг "без пробега по РФ" ("да" или "нет")
	Color         string   // Цвет автомобиля
	BodyType      string   // Тип кузова
	Power         string   // Мощность двигателя (л.с.)
	FuelType      string   // Тип топлива
	EngineVolume  string   // Объем двигателя
	Photos        []string // Список путей к фотографиям автомобиля
}

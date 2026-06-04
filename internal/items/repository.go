package items

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository interface {
	Upsert(items []Item) (int, error)
	GetByID(id uint64) (*Item, error)

	GetRandomNames(count int) ([]string, error)
	GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db: db,
	}
}

func (r *repository) Upsert(items []Item) (int, error) {
	const batchSize = 100
	total := 0

	for i := 0; i < len(items); i += batchSize {
		end := min(i+batchSize, len(items))
		batch := items[i:end]

		result := r.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "external_id"},
				{Name: "set_external_id"},
				{Name: "tcg"},
				{Name: "code"},
				{Name: "lang"},
				{Name: "rarity"},
				{Name: "edition"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"edition",
				"print_url_small",
				"print_url_large",
				"set_image",
				"quantity_per_set",
				"quantity_per_box",
				"updated_at",
			}),
		}).Create(&batch)

		if result.Error != nil {
			return 0, result.Error
		}
		total += int(result.RowsAffected)
	}

	return total, nil
}

func (r *repository) GetByID(id uint64) (*Item, error) {
	var item Item
	result := r.db.Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, result.Error
	}

	r.db.Model(&item).UpdateColumn("wanted", gorm.Expr("wanted + 1"))

	return &item, nil
}

func (r *repository) GetSuggestions(input SuggestionInput) ([]SuggestionDTO, error) {
	var suggestions []SuggestionDTO

	// El SELECT incluye todos los campos del DTO más wanted para el ORDER BY externo
	baseSelect := `
		id,
		external_id,
		set_external_id,
		type,
		tcg,
		name,
		code,
		rarity,
		rarity_code,
		set_name,
		set_code,
		lang,
		edition,
		wanted,
		COALESCE(
			NULLIF(print_url_small, ''),
			images->0->>'image_url_small',
			set_image,
			''
		) AS image
	`

	// Construimos filtros opcionales de TCG e Idioma comunes
	filterQuery := r.db.Table("items")
	if input.TCG != "" {
		filterQuery = filterQuery.Where("tcg = ?", input.TCG)
	}
	if input.Lang != "" {
		filterQuery = filterQuery.Where("lang = ?", input.Lang)
	}

	namePattern := "%" + input.Input + "%"
	otherPattern := input.Input + "%"

	// Creamos subconsultas independientes
	sub1 := filterQuery.Session(&gorm.Session{}).
		Select("1 as priority, "+baseSelect).
		Where("name ILIKE ?", namePattern).
		Limit(10)

	sub2 := filterQuery.Session(&gorm.Session{}).
		Select("2 as priority, "+baseSelect).
		Where("code ILIKE ?", otherPattern).
		Limit(10)

	sub3 := filterQuery.Session(&gorm.Session{}).
		Select("3 as priority, "+baseSelect).
		Where("archetype ILIKE ?", otherPattern).
		Limit(10)

	sub4 := filterQuery.Session(&gorm.Session{}).
		Select("4 as priority, "+baseSelect).
		Where("external_id ILIKE ?", otherPattern).
		Limit(10)

	// Unimos los resultados mediante un Raw query con UNION.
	unionSQL := r.db.Raw(`
		SELECT * FROM (
			(?) UNION (?) UNION (?) UNION (?)
		) AS combined_suggestions
		ORDER BY
		    combined_suggestions.priority ASC,
		    combined_suggestions.wanted DESC
		LIMIT 10
	`, sub1, sub2, sub3, sub4)

	if err := unionSQL.Scan(&suggestions).Error; err != nil {
		return nil, err
	}

	return suggestions, nil
}

func (r *repository) GetRandomNames(count int) ([]string, error) {
	var names []string

	subQuery := r.db.Table("items").
		Select("DISTINCT name").
		Where("lang = ?", "EN")

	result := r.db.Table("(?) as t", subQuery).
		Select("name").
		Order("RANDOM()").
		Limit(count).
		Find(&names)

	return names, result.Error
}
